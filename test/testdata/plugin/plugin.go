// Command plugin is the fixture the specs load into make.
//
// It lives under testdata so that go build ./... skips it: a main package
// cannot be linked against gmk_ symbols, which exist only inside a running
// make. The specs build it explicitly with -buildmode=c-shared.
//
// make derives the setup symbol from the object's base name, so the specs
// must write it to plugin.so.
package main

/*
#include <gnumake.h>
*/
import "C"

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/unmango/go-gmk"
)

func main() {}

//export plugin_gmk_setup
func plugin_gmk_setup(*C.gmk_floc) C.int {
	setupRuns++

	gmk.AddFunction("fixture-setup-runs", setupRunCount, 0, 0, gmk.FuncDefault)
	gmk.AddFunction("fixture-upper", upper, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-join", join, 1, 0, gmk.FuncDefault)
	gmk.AddFunction("fixture-repeat", repeat, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-expand", expand, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-count", count, 0, 0, gmk.FuncDefault)
	gmk.AddFunction("fixture-arity", join, 2, 3, gmk.FuncDefault)
	gmk.AddFunction("fixture-name", name, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-alias", name, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-special", special, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-alloc", alloc, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-alloc-empty", allocEmpty, 0, 0, gmk.FuncDefault)
	gmk.AddFunction("fixture-free-empty", freeEmpty, 0, 0, gmk.FuncDefault)
	gmk.AddFunction("fixture-panics", panics, 1, 1, gmk.FuncDefault)
	gmk.AddFunction("fixture-eval", evalNow, 1, 1, gmk.FuncNoExpand)
	gmk.AddFunction("fixture-eval-at", evalAt, 1, 1, gmk.FuncNoExpand)

	// raw and expanded are the same Go function under two registrations that
	// differ only in flags, so a spec can compare what each one receives.
	gmk.AddFunction("fixture-raw", first, 1, 1, gmk.FuncNoExpand)
	gmk.AddFunction("fixture-expanded", first, 1, 1, gmk.FuncDefault)

	// Registering twice must replace the Go function without calling into
	// make again, which make would reject as a duplicate.
	gmk.AddFunction("fixture-replaced", constant("first"), 0, 0, gmk.FuncDefault)
	gmk.AddFunction("fixture-replaced", constant("second"), 0, 0, gmk.FuncDefault)

	gmk.Eval("FIXTURE_FROM_PLUGIN := yes", nil)

	// A rule evaluated here is the first target make has seen, so it becomes
	// the default goal. The specs always name their target for that reason.
	gmk.Eval("fixture-target:\n\t@echo from-plugin-rule", nil)

	return 1
}

// setupRuns counts how many times make has called the setup symbol in this
// process, which is how a spec tells a repeated load directive from a repeated
// initialization.
var setupRuns int

func setupRunCount(string, []string) string {
	return strconv.Itoa(setupRuns)
}

func upper(_ string, args []string) string {
	return strings.ToUpper(args[0])
}

// join proves the whole argument vector arrives, not only the first entry. The
// separator is not a shell metacharacter so the specs can echo the result.
func join(_ string, args []string) string {
	return strings.Join(args, "+")
}

// count reports how many arguments arrived, which join cannot distinguish from
// a call whose arguments were all empty.
func count(_ string, args []string) string {
	return strconv.Itoa(len(args))
}

// name returns the name make dispatched under, so a spec can register one Go
// function under two names and see which one was called.
func name(nm string, _ []string) string {
	return nm
}

func first(_ string, args []string) string {
	return args[0]
}

func constant(s string) gmk.Func {
	return func(string, []string) string { return s }
}

// maxRepeat bounds repeat so a makefile cannot ask the fixture to exhaust
// memory.
const maxRepeat = 1 << 16

// repeat returns args[0] words, which is longer than any buffer make holds on
// the plugin's behalf.
func repeat(_ string, args []string) string {
	n, err := strconv.Atoi(strings.TrimSpace(args[0]))
	if err != nil || n < 0 || n > maxRepeat {
		return ""
	}

	return strings.TrimSpace(strings.Repeat("x ", n))
}

func expand(_ string, args []string) string {
	return gmk.Expand(args[0])
}

// specials are results that stress how allocResult sizes and terminates the
// buffer it hands make. A spec writes each one to a file rather than echoing
// it, so a shell never gets the chance to reinterpret it.
var specials = map[string]string{
	"newline":   "a\nb",
	"tab":       "a\tb",
	"dollar":    "a$b",
	"hash":      "a#b",
	"backslash": `a\`,
	"utf8":      "héllo→ß",
	"spaces":    "  padded  ",
	"empty":     "",
}

func special(_ string, args []string) string {
	return specials[strings.TrimSpace(args[0])]
}

// alloc takes a buffer from make, writes every byte of it, reads it back into
// a Go string, and releases it. Returning the content proves the buffer was
// both writable and still intact at the moment it was copied.
func alloc(_ string, args []string) string {
	n, err := strconv.Atoi(strings.TrimSpace(args[0]))
	if err != nil || n < 0 || n > maxRepeat {
		return "bad-argument"
	}

	buf := gmk.Alloc(n)
	if n > 0 && buf == nil {
		return "nil-buffer"
	}
	if len(buf) != n {
		return fmt.Sprintf("wrong-length-%d", len(buf))
	}

	for i := range buf {
		buf[i] = 'x'
	}

	s := string(buf)
	gmk.Free(buf)

	return s
}

// allocEmpty reports what Alloc does with sizes make would reject.
func allocEmpty(string, []string) string {
	return fmt.Sprintf("%d %d",
		len(gmk.Alloc(0)),
		len(gmk.Alloc(-1)),
	)
}

// freeEmpty proves Free tolerates what Alloc returns for those sizes, rather
// than handing make a pointer into the Go heap.
func freeEmpty(string, []string) string {
	gmk.Free(nil)
	gmk.Free([]byte{})
	gmk.Free(gmk.Alloc(0))

	return "ok"
}

// badCalls are the AddFunction arguments make would reject. They cannot be
// reached from a makefile, so the fixture makes the panic observable instead.
var badCalls = map[string]func(){
	"empty-name": func() { gmk.AddFunction("", first, 1, 1, gmk.FuncDefault) },
	"long-name":  func() { gmk.AddFunction(strings.Repeat("n", 256), first, 1, 1, gmk.FuncDefault) },
	"nil-func":   func() { gmk.AddFunction("fixture-nil", nil, 1, 1, gmk.FuncDefault) },
	"neg-min":    func() { gmk.AddFunction("fixture-neg", first, -1, 1, gmk.FuncDefault) },
	"big-min":    func() { gmk.AddFunction("fixture-big", first, 256, 0, gmk.FuncDefault) },
	"big-max":    func() { gmk.AddFunction("fixture-max", first, 1, 256, gmk.FuncDefault) },
	"min-gt-max": func() { gmk.AddFunction("fixture-order", first, 3, 2, gmk.FuncDefault) },
}

func panics(_ string, args []string) (msg string) {
	call, ok := badCalls[strings.TrimSpace(args[0])]
	if !ok {
		return "unknown-case"
	}

	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprint(r)
		}
	}()

	call()

	return "no-panic"
}

// evalNow calls Eval from inside an expansion rather than from setup, which is
// the reentrant path: make is already expanding when the parse begins.
func evalNow(_ string, args []string) string {
	gmk.Eval(makefileText(args[0]), nil)

	return ""
}

// makefileText turns the escapes a makefile can pass through a function
// argument into the newlines and tabs a recipe needs.
func makefileText(s string) string {
	return strings.NewReplacer(`\n`, "\n", `\t`, "\t").Replace(s)
}

// evalAt attributes the evaluated text to a makefile that does not exist, so a
// spec can read the location back out of the diagnostic make prints.
func evalAt(_ string, args []string) string {
	gmk.Eval(args[0], &gmk.FileLocation{
		FileName:   "phony.mk",
		LineNumber: 42,
	})

	return ""
}
