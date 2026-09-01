// Command badsetup is a fixture whose setup function reports failure.
//
// make treats a zero return from the setup symbol as "this object refused to
// initialize" and stops. The object is otherwise well formed, so a spec that
// loads it isolates the return value from every other reason a load can fail.
//
// The specs must write it to badsetup.so, since make derives the setup symbol
// from the object's base name.
package main

/*
#include <gnumake.h>
*/
import "C"

import (
	"github.com/unmango/go-gmk"
)

func main() {}

//export badsetup_gmk_setup
func badsetup_gmk_setup(*C.gmk_floc) C.int {
	// Evaluated before the refusal so a spec can tell "setup never ran" from
	// "setup ran and said no".
	gmk.Eval("BADSETUP_RAN := yes", nil)

	return 0
}
