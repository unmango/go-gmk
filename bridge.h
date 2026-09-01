#ifndef GO_GMK_BRIDGE_H
#define GO_GMK_BRIDGE_H

/* Proxy passed to gmk_add_function. A Go function value cannot be used as a C
   function pointer, so make calls this and it forwards to the Go bridge. */
char *go_gmk_func_bridge(const char *nm, unsigned int argc, char **argv);

#endif
