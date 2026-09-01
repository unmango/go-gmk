# go-gmk

Go bindings for the GNU Make loadable object API.

The package binds the `gmk_` functions `gnumake.h` declares, which are what a shared object loaded with make's `load` directive can call.

<https://www.gnu.org/software/make/manual/html_node/Loading-Objects.html>

## Usage

A plugin is a `main` package built as a shared object.
make derives the setup symbol from the object's base name, so `hello.so` must export `hello_gmk_setup`.

```go
//go:build makeplugin

package main

/*
#include <gnumake.h>
*/
import "C"

import (
	"strings"

	"github.com/unmango/go-gmk"
)

func main() {}

//export hello_gmk_setup
func hello_gmk_setup(*C.gmk_floc) C.int {
	gmk.AddFunction("hello-upper", upper, 1, 1, gmk.FuncDefault)
	gmk.Eval("HELLO_FROM_PLUGIN := yes", nil)
	return 1
}

func upper(_ string, args []string) string {
	return strings.ToUpper(args[0])
}
```

Build it and load it:

```sh
go build -tags makeplugin -buildmode=c-shared -o hello.so ./examples/hello
```

make searches the dynamic loader's path for a bare name, so give it an explicit relative path:

```make
load ./hello.so

all:
	@echo $(hello-upper abc)      # ABC
	@echo $(HELLO_FROM_PLUGIN)    # yes
```

The build tag keeps the plugin out of `go build ./...`, which cannot link a `main` package against symbols that only exist inside make.

`plugin_is_GPL_compatible` is defined by this package, so importing it satisfies make's check.

## Layout

`internal/gnumake` holds the raw bindings, generated from `gnumake.h` with [c-for-go](https://github.com/xlab/c-for-go).
Run `make gen` to regenerate them.

The root package, `gmk`, is written by hand.
`gmk_add_function` needs a callback bridge that copies its result into `gmk_alloc` memory, and `gmk_expand` returns a NUL-terminated buffer the caller must free, neither of which c-for-go can express.

## Development

```sh
nix develop      # go, ginkgo, gomod2nix, and gnumake.h on CGO_CFLAGS
make test        # ginkgo run -r
make build       # nix build .#
```

The specs build fixture plugins from `test/testdata` and run real make against them, since the bindings resolve their symbols from the loading process.
`test/testdata/plugin` registers everything the specs call; `badsetup` and `nogpl` are the two ways a load can fail.

The suite skips itself when the make on `PATH` is older than 4.3, which is where the `load` directive arrived.
Specs labelled `slow` cover parallel builds and memory growth:

```sh
ginkgo run --label-filter='!slow' ./test   # the fast subset
```
