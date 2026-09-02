package test_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// repoRoot is the directory holding the vendored gnumake.h. The specs run from
// test/, so the header the fixture builds need is one level up.
const repoRoot = ".."

// buildEnv returns the environment a fixture build runs in. The repo root is
// added to CGO_CFLAGS so go test works without the devshell staging the header
// somewhere the default include path finds it.
func buildEnv() []string {
	GinkgoHelper()

	root, err := filepath.Abs(repoRoot)
	Expect(err).NotTo(HaveOccurred())

	cflags := strings.TrimSpace(os.Getenv("CGO_CFLAGS") + " -I" + root)

	return append(os.Environ(), "CGO_ENABLED=1", "CGO_CFLAGS="+cflags)
}

// buildPlugin compiles the fixture package under testdata into a shared object
// named soName and returns its path.
//
// The name matters: make derives the setup symbol it looks for from the
// object's base name, so a fixture whose setup is named plugin_gmk_setup has
// to be written to plugin.so.
func buildPlugin(pkg, soName string) string {
	GinkgoHelper()

	path := filepath.Join(GinkgoT().TempDir(), soName)

	build := exec.Command("go", "build",
		"-buildmode=c-shared",
		"-o", path,
		"./testdata/"+pkg,
	)
	build.Env = buildEnv()

	out, err := build.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "building fixture %s: %s", pkg, out)

	return path
}

// makeOpts describes one invocation of make.
type makeOpts struct {
	// Body is appended to the load directive to form the makefile. When Load
	// is set it names the object to load instead of the main fixture; an
	// empty Load loads the main fixture, and NoLoad omits the directive.
	Body string
	// Load overrides which object the generated makefile loads.
	Load string
	// NoLoad writes the body with no load directive at all.
	NoLoad bool
	// Target is the goal make is asked to build. It defaults to all.
	//
	// It is always passed explicitly because a rule the plugin evaluates from
	// its setup function becomes the first target make has seen, and so would
	// otherwise silently become the default goal.
	Target string
	// Args are extra arguments passed to make, such as -j4.
	Args []string
	// Env entries are appended to the make process environment.
	Env []string
	// Files are extra files written into the run directory, keyed by name.
	Files map[string]string
	// Dir is the directory make runs in. An empty Dir gets a fresh temp
	// directory, which is what every spec that only reads stdout wants.
	// Setting it lets a spec read back a file the makefile wrote.
	Dir string
}

// makeResult is everything a spec can learn about one make run.
type makeResult struct {
	Stdout string
	Stderr string
	// Err is make's exit error, nil when it succeeded.
	Err error
	// State carries the resource usage of the make process, which the leak
	// specs read through SysUsage.
	State *os.ProcessState
}

// Output is both streams joined, which is what the specs asserting on a
// diagnostic want, since make splits its output between them.
func (r makeResult) Output() string {
	return r.Stdout + r.Stderr
}

// runMakeRaw runs make and returns the result without asserting anything about
// it, so it is the entry point the failure and resource specs need.
func runMakeRaw(opts makeOpts) makeResult {
	GinkgoHelper()

	dir := opts.Dir
	if dir == "" {
		dir = GinkgoT().TempDir()
	}

	contents := opts.Body
	if !opts.NoLoad {
		object := opts.Load
		if object == "" {
			object = pluginPath
		}
		contents = "load " + object + "\n\n" + opts.Body
	}

	makefile := filepath.Join(dir, "Makefile")
	Expect(os.WriteFile(makefile, []byte(contents), 0o644)).To(Succeed())

	for name, body := range opts.Files {
		Expect(os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)).To(Succeed())
	}

	target := opts.Target
	if target == "" {
		target = "all"
	}

	args := append([]string{"--no-print-directory", "-f", makefile}, opts.Args...)
	args = append(args, target)

	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command("make", args...)
	cmd.Dir = dir
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Env = append(os.Environ(), opts.Env...)

	err := cmd.Run()

	return makeResult{
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
		Err:    err,
		State:  cmd.ProcessState,
	}
}

// runMake writes body to a makefile that loads the main fixture and runs it,
// asserts make succeeded, and returns everything it printed.
func runMake(body string) string {
	GinkgoHelper()

	return runMakeWith(makeOpts{Body: body})
}

// runMakeWith is runMake with the full option set.
func runMakeWith(opts makeOpts) string {
	GinkgoHelper()

	res := runMakeRaw(opts)
	Expect(res.Err).NotTo(HaveOccurred(), "make said:\n%s", res.Output())

	return strings.TrimSpace(res.Output())
}

// runMakeFail asserts make exited non-zero and returns what it printed. The
// negative specs care about the diagnostic, which make writes to stderr, so
// both streams come back joined.
func runMakeFail(opts makeOpts) string {
	GinkgoHelper()

	res := runMakeRaw(opts)
	Expect(res.Err).To(HaveOccurred(), "expected make to fail, it printed:\n%s", res.Output())

	return res.Output()
}

// recipe builds a makefile whose default target echoes each line given. Most
// specs only want to see what one expansion produced, and writing the tab-led
// recipe by hand in every spec obscures that.
func recipe(lines ...string) string {
	var b strings.Builder
	b.WriteString("all:\n")
	for _, line := range lines {
		fmt.Fprintf(&b, "\t@echo %s\n", line)
	}

	return b.String()
}
