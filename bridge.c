#include <gnumake.h>

#include "_cgo_export.h"
#include "bridge.h"

char *go_gmk_func_bridge(const char *nm, unsigned int argc, char **argv) {
	return goFuncBridge((char *)nm, argc, argv);
}

/* make refuses to load an object that does not export this symbol. Defining it
   here means every plugin that imports this package satisfies the check. */
int plugin_is_GPL_compatible;
