// Package gmk provides Go bindings for the GNU Make loadable object API.
//
// The functions here are only callable from inside a make process that has
// loaded the compiled object with the load directive. Build a plugin with
// go build -buildmode=c-shared and export plugin_is_GPL_compatible, then load
// it from a makefile:
//
//	load plugin.so
//
// See https://www.gnu.org/software/make/manual/html_node/Loading-Objects.html.
package gmk

/*
#include <gnumake.h>
#include <stdlib.h>

#include "bridge.h"
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/unmango/go-gmk/internal/gnumake"
)

// Argument expansion modes for [AddFunction].
const (
	// FuncDefault expands each argument before the function is called.
	FuncDefault uint = C.GMK_FUNC_DEFAULT
	// FuncNoExpand passes each argument to the function unexpanded.
	FuncNoExpand uint = C.GMK_FUNC_NOEXPAND
)

// FileLocation is the makefile position an evaluation is attributed to.
type FileLocation struct {
	FileName   string
	LineNumber uint
}

// Func is a function that make can call. The arguments are the expanded
// argument text unless the function was registered with [FuncNoExpand], and
// the returned string is substituted for the call.
type Func func(name string, args []string) string

var (
	funcsMu sync.RWMutex
	funcs   = map[string]Func{}
)

// Alloc returns a buffer of n bytes owned by make. Pass it back to [Free] when
// it is no longer needed, or hand it to make as a function result.
func Alloc(n int) []byte {
	if n <= 0 {
		return nil
	}
	p := gnumake.Alloc(uint32(n))
	if p == nil {
		return nil
	}
	return unsafe.Slice(p, n)
}

// Free releases a buffer returned by [Alloc].
func Free(b []byte) {
	if len(b) == 0 {
		return
	}
	C.gmk_free((*C.char)(unsafe.Pointer(&b[0])))
}

// Expand expands variable and function references in s the way make would and
// returns the result. The buffer make allocates for the result is freed before
// returning, so the returned string is owned by Go.
func Expand(s string) string {
	p := gnumake.Expand(s)
	if p == nil {
		return ""
	}

	// The generated Free takes a slice, which cannot describe a buffer whose
	// length is only known from its terminator, so the C call is direct.
	res := (*C.char)(unsafe.Pointer(p))
	defer C.gmk_free(res)

	return C.GoString(res)
}

// Eval parses buf as makefile text, as though it appeared at loc. A nil loc
// attributes the text to the object that called Eval.
func Eval(buf string, loc *FileLocation) {
	cbuf := C.CString(buf)
	defer C.free(unsafe.Pointer(cbuf))

	if loc == nil {
		C.gmk_eval(cbuf, nil)
		return
	}

	cname := C.CString(loc.FileName)
	defer C.free(unsafe.Pointer(cname))

	floc := C.gmk_floc{
		filenm: cname,
		lineno: C.ulong(loc.LineNumber),
	}

	C.gmk_eval(cbuf, &floc)
}

// maxNameLen and maxArgCount are the limits gnumake.h documents for
// gmk_add_function.
const (
	maxNameLen  = 255
	maxArgCount = 255
)

// AddFunction registers fn with make under name, callable as $(name ...) with
// between minArgs and maxArgs arguments. A maxArgs of 0 means no upper bound.
// flags is [FuncDefault] or [FuncNoExpand].
//
// Re-registering a name replaces the Go function without calling into make
// again, since make rejects a duplicate registration.
//
// AddFunction panics on an argument make would reject. It is called from a
// plugin's setup, where the arguments are constants and a bad one is a
// programming error rather than something a makefile can trigger.
func AddFunction(name string, fn Func, minArgs, maxArgs int, flags uint) {
	switch {
	case name == "" || len(name) > maxNameLen:
		panic("gmk: function name must be 1 to 255 bytes")
	case fn == nil:
		panic("gmk: function must not be nil")
	case minArgs < 0 || minArgs > maxArgCount:
		panic("gmk: minArgs must be 0 to 255")
	case maxArgs < 0 || maxArgs > maxArgCount:
		panic("gmk: maxArgs must be 0 to 255")
	case maxArgs != 0 && minArgs > maxArgs:
		panic("gmk: minArgs must not exceed maxArgs")
	}

	funcsMu.Lock()
	_, registered := funcs[name]
	funcs[name] = fn
	funcsMu.Unlock()

	if registered {
		return
	}

	// make stores the name pointer rather than copying it, so this allocation
	// is deliberately never freed.
	cname := C.CString(name)

	C.gmk_add_function(cname,
		C.gmk_func_ptr(C.go_gmk_func_bridge),
		C.uint(minArgs),
		C.uint(maxArgs),
		C.uint(flags),
	)
}

//export goFuncBridge
func goFuncBridge(nm *C.char, argc C.uint, argv **C.char) *C.char {
	name := C.GoString(nm)

	funcsMu.RLock()
	fn := funcs[name]
	funcsMu.RUnlock()

	if fn == nil {
		return nil
	}

	return allocResult(fn(name, packArgs(argc, argv)))
}

// packArgs copies the argument vector make passes into Go strings. c-for-go
// leaves the destination slice nil, so this is written by hand.
//
// The result always holds argc entries. A nil entry becomes an empty string
// rather than truncating, so an argument's position never shifts.
func packArgs(argc C.uint, argv **C.char) []string {
	if argv == nil || argc == 0 {
		return nil
	}

	cargs := unsafe.Slice(argv, int(argc))
	args := make([]string, len(cargs))
	for i, carg := range cargs {
		if carg != nil {
			args[i] = C.GoString(carg)
		}
	}

	return args
}

// allocResult copies s into memory make owns. make frees a function result
// itself, so returning a pointer into the Go heap would hand it a pointer it
// must not free.
func allocResult(s string) *C.char {
	buf := C.gmk_alloc(C.uint(len(s) + 1))
	if buf == nil {
		return nil
	}

	dst := unsafe.Slice((*byte)(unsafe.Pointer(buf)), len(s)+1)
	copy(dst, s)
	dst[len(s)] = 0

	return buf
}
