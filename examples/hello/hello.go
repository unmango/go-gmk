//go:build makeplugin

// Command hello is an example GNU Make plugin.
//
// Build it as a shared object and load it from a makefile:
//
//	go build -tags makeplugin -buildmode=c-shared -o hello.so ./examples/hello
//	echo 'load ./hello.so' > Makefile
//
// make derives the setup symbol from the object's base name, so hello.so must
// export hello_gmk_setup.
package main

/*
#include <gnumake.h>
*/
import "C"

import (
	"strconv"
	"strings"

	gnumake "github.com/unmango/gnumake-go"
)

func main() {}

//export hello_gmk_setup
func hello_gmk_setup(*C.gmk_floc) C.int {
	gnumake.AddFunction("hello-upper", upper, 1, 1, gnumake.FuncDefault)
	gnumake.AddFunction("hello-join", join, 1, 0, gnumake.FuncDefault)
	gnumake.AddFunction("hello-repeat", repeat, 1, 1, gnumake.FuncDefault)
	gnumake.AddFunction("hello-expand", expand, 1, 1, gnumake.FuncDefault)

	gnumake.Eval("HELLO_FROM_PLUGIN := yes", nil)

	return 1
}

func upper(_ string, args []string) string {
	return strings.ToUpper(args[0])
}

// join proves the whole argument vector arrives, not only the first entry.
func join(_ string, args []string) string {
	return strings.Join(args, "+")
}

// maxRepeat bounds repeat so a makefile cannot ask the plugin to exhaust
// memory.
const maxRepeat = 1 << 16

// repeat returns args[0] words.
func repeat(_ string, args []string) string {
	n, err := strconv.Atoi(strings.TrimSpace(args[0]))
	if err != nil || n < 0 || n > maxRepeat {
		return ""
	}

	return strings.TrimSpace(strings.Repeat("x ", n))
}

func expand(_ string, args []string) string {
	return gnumake.Expand(args[0])
}
