// Command nogpl is a fixture that never imports this package.
//
// bridge.c defines plugin_is_GPL_compatible so that importing gmk is enough to
// satisfy make's license check. This fixture is the control: it is a valid
// shared object with a valid setup symbol and nothing else, so the only reason
// make can reject it is the missing symbol.
//
// The specs must write it to nogpl.so, since make derives the setup symbol
// from the object's base name.
package main

import "C"

import "unsafe"

func main() {}

// The location argument is taken as an opaque pointer rather than a
// *C.gmk_floc, so the fixture does not need gnumake.h either.
//
//export nogpl_gmk_setup
func nogpl_gmk_setup(unsafe.Pointer) C.int {
	return 1
}
