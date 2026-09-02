package test_test

import (
	"os/exec"
	"regexp"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGnumake(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gnumake Suite")
}

// The bindings resolve gmk_expand and friends from the make process that loads
// them, so they cannot be exercised in a plain test binary. Every spec builds a
// fixture plugin as a shared object, loads it from a makefile, and reads back
// what make printed.
var (
	// pluginPath is the main fixture, which registers everything the specs
	// call. Its setup symbol is plugin_gmk_setup, so it must be plugin.so.
	pluginPath string
	// badSetupPath is a fixture whose setup returns 0.
	badSetupPath string
	// noGPLPath is a fixture that never imports this package, and so never
	// picks up the plugin_is_GPL_compatible symbol bridge.c defines.
	noGPLPath string
)

var _ = BeforeSuite(func() {
	requireLoadableMake()

	pluginPath = buildPlugin("plugin", "plugin.so")
	badSetupPath = buildPlugin("badsetup", "badsetup.so")
	noGPLPath = buildPlugin("nogpl", "nogpl.so")
})

var makeVersion = regexp.MustCompile(`GNU Make (\d+)\.(\d+)`)

// requireLoadableMake skips the whole suite when the make on PATH cannot load
// objects, which is a build-time option. Without the check every spec fails
// identically and none of them says why.
func requireLoadableMake() {
	out, err := exec.Command("make", "--version").Output()
	if err != nil {
		Skip("make is not on PATH: " + err.Error())
	}

	m := makeVersion.FindSubmatch(out)
	if m == nil {
		Skip("make on PATH is not GNU make: " + string(out))
	}

	major, _ := strconv.Atoi(string(m[1]))
	minor, _ := strconv.Atoi(string(m[2]))
	if major < 4 || (major == 4 && minor < 3) {
		Skip("the load directive needs GNU make 4.3 or later, found " + string(m[0]))
	}
}
