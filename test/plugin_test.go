package test_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The bindings resolve gmk_expand and friends from the make process that loads
// them, so they cannot be exercised in a plain test binary. Every spec builds
// the fixture plugin as a shared object, loads it from a makefile, and reads
// back what make printed.

var pluginPath string

var _ = BeforeSuite(func() {
	dir := GinkgoT().TempDir()
	pluginPath = filepath.Join(dir, "plugin.so")

	build := exec.Command("go", "build",
		"-buildmode=c-shared",
		"-o", pluginPath,
		"./testdata/plugin",
	)
	build.Env = append(os.Environ(), "CGO_ENABLED=1")

	out, err := build.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "building the fixture plugin: %s", out)
})

// runMake writes body to a makefile that loads the plugin and runs it, and
// returns everything make printed.
func runMake(body string) string {
	GinkgoHelper()

	dir := GinkgoT().TempDir()
	makefile := filepath.Join(dir, "Makefile")
	contents := "load " + pluginPath + "\n\n" + body

	Expect(os.WriteFile(makefile, []byte(contents), 0o644)).To(Succeed())

	cmd := exec.Command("make", "--no-print-directory", "-f", makefile)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "make said: %s", out)

	return strings.TrimSpace(string(out))
}

var _ = Describe("Loading the object", func() {
	It("should run a makefile that loads the plugin", func() {
		Expect(runMake("all:\n\t@echo loaded\n")).To(Equal("loaded"))
	})

	It("should fail the load when the object does not export the setup symbol", func() {
		dir := GinkgoT().TempDir()
		makefile := filepath.Join(dir, "Makefile")
		Expect(os.WriteFile(makefile, []byte("load /nonexistent.so\nall:;@true\n"), 0o644)).To(Succeed())

		cmd := exec.Command("make", "-f", makefile)
		cmd.Dir = dir

		out, err := cmd.CombinedOutput()
		Expect(err).To(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("nonexistent.so"))
	})
})

var _ = Describe("AddFunction", func() {
	It("should substitute the value the Go function returns", func() {
		Expect(runMake("all:\n\t@echo $(fixture-upper abc)\n")).To(Equal("ABC"))
	})

	It("should pass every argument through, not just the first", func() {
		Expect(runMake("all:\n\t@echo $(fixture-join a,b,c)\n")).To(Equal("a+b+c"))
	})

	It("should pass a single argument through", func() {
		Expect(runMake("all:\n\t@echo $(fixture-join solo)\n")).To(Equal("solo"))
	})

	It("should return an empty result without upsetting make", func() {
		Expect(runMake("all:\n\t@echo [$(fixture-upper )]\n")).To(Equal("[]"))
	})

	It("should survive a result longer than make's own buffer", func() {
		out := runMake("all:\n\t@echo $(words $(fixture-repeat 4096))\n")
		Expect(out).To(Equal("4096"))
	})

	It("should keep working after the same name is registered twice", func() {
		Expect(runMake("all:\n\t@echo $(fixture-upper abc)\n")).To(Equal("ABC"))
	})
})

var _ = Describe("Expand", func() {
	It("should expand a variable reference the way make would", func() {
		body := "GREETING := hi\nall:\n\t@echo $(fixture-expand $$(GREETING))\n"
		Expect(runMake(body)).To(Equal("hi"))
	})

	It("should expand a function call", func() {
		body := "all:\n\t@echo $(fixture-expand $$(subst a,b,aaa))\n"
		Expect(runMake(body)).To(Equal("bbb"))
	})
})

var _ = Describe("Eval", func() {
	It("should define a variable from the plugin's setup", func() {
		Expect(runMake("all:\n\t@echo $(FIXTURE_FROM_PLUGIN)\n")).To(Equal("yes"))
	})
})
