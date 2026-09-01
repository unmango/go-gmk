#include <gnumake.h>

#include "_cgo_export.h"
#include "bridge.h"

char *gnumake_go_func_bridge(const char *nm, unsigned int argc, char **argv) {
	return goFuncBridge((char *)nm, argc, argv);
}
