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
	"strconv"
	"strings"

	gnumake "github.com/unmango/gnumake-go"
)

func main() {}

//export plugin_gmk_setup
func plugin_gmk_setup(*C.gmk_floc) C.int {
	gnumake.AddFunction("fixture-upper", upper, 1, 1, gnumake.FuncDefault)
	gnumake.AddFunction("fixture-join", join, 1, 0, gnumake.FuncDefault)
	gnumake.AddFunction("fixture-repeat", repeat, 1, 1, gnumake.FuncDefault)
	gnumake.AddFunction("fixture-expand", expand, 1, 1, gnumake.FuncDefault)

	// Registering twice must replace the Go function without calling into
	// make again, which make would reject as a duplicate.
	gnumake.AddFunction("fixture-upper", upper, 1, 1, gnumake.FuncDefault)

	gnumake.Eval("FIXTURE_FROM_PLUGIN := yes", nil)

	return 1
}

func upper(_ string, args []string) string {
	return strings.ToUpper(args[0])
}

// join proves the whole argument vector arrives, not only the first entry. The
// separator is not a shell metacharacter so the specs can echo the result.
func join(_ string, args []string) string {
	return strings.Join(args, "+")
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
	return gnumake.Expand(args[0])
}
