package test_test

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs are slower than the rest and are labelled so a quick run can
// skip them with ginkgo --label-filter='!slow'.
var _ = Describe("Under load", Label("slow"), func() {
	It("should expand in every branch of a parallel build", func() {
		var b strings.Builder
		b.WriteString("all:")
		for i := range 8 {
			fmt.Fprintf(&b, " job-%d", i)
		}
		b.WriteString("\n\njob-%:\n\t@echo $(fixture-upper j$*)\n")

		out := runMakeWith(makeOpts{Body: b.String(), Args: []string{"-j8"}})

		for i := range 8 {
			Expect(out).To(ContainSubstring(fmt.Sprintf("J%d", i)))
		}
	})

	It("should load into parallel sub-makes independently", func() {
		// Eight make processes each dlopen the object and start their own Go
		// runtime, which the single-process specs never exercise.
		var b strings.Builder
		b.WriteString("all:")
		for i := range 8 {
			fmt.Fprintf(&b, " job-%d", i)
		}
		b.WriteString("\n\njob-%:\n\t@$(MAKE) --no-print-directory -f inner.mk V=$* inner\n")

		out := runMakeWith(makeOpts{
			NoLoad: true,
			Body:   b.String(),
			Args:   []string{"-j8"},
			Files: map[string]string{
				"inner.mk": "load " + pluginPath +
					"\ninner:\n\t@echo $(fixture-upper j$(V))\n",
			},
		})

		for i := range 8 {
			Expect(out).To(ContainSubstring(fmt.Sprintf("J%d", i)))
		}
	})

	It("should return a result at the fixture's size ceiling", func() {
		Expect(runMake(recipe("$(words $(fixture-repeat 65536))"))).To(Equal("65536"))
	})
})

var _ = Describe("Repeated expansion", Label("slow"), func() {
	// Expand frees the buffer make allocates for it. Nothing in a single call
	// can show that, so the check is whether make's peak memory tracks the
	// number of calls.
	BeforeEach(func() {
		if runtime.GOOS != "linux" {
			Skip("maximum resident set size is reported in different units per platform")
		}
	})

	// expansions builds a makefile that expands a kilobyte of text n times.
	//
	// Each result is thrown away by $(if ...,,) rather than collected. A
	// foreach that keeps them holds every result until the loop ends, which
	// costs the same n kilobytes a leak would and would hide one completely.
	expansions := func(n int) string {
		return "VAL := " + strings.Repeat("v", 1024) + "\n" +
			"NUMS := $(shell seq 1 " + fmt.Sprint(n) + ")\n" +
			"X := $(foreach i,$(NUMS),$(if $(fixture-expand $$(VAL)),,))\n" +
			"all:;@true\n"
	}

	maxRSSKiB := func(n int) int64 {
		GinkgoHelper()

		res := runMakeRaw(makeOpts{Body: expansions(n)})
		Expect(res.Err).NotTo(HaveOccurred(), "make said:\n%s", res.Output())

		usage, ok := res.State.SysUsage().(*syscall.Rusage)
		Expect(ok).To(BeTrue(), "resource usage was not a *syscall.Rusage")

		return usage.Maxrss
	}

	It("should not grow make's peak memory with the number of calls", func() {
		const (
			baseline = 100_000
			heavy    = 400_000
			// Retaining every buffer would cost the 300k extra calls some
			// 300MB. The slack covers the Go heap settling at a larger size
			// for the longer run, which is tens of megabytes.
			slackKiB = 64 << 10
		)

		Expect(maxRSSKiB(heavy)).To(BeNumerically("<", maxRSSKiB(baseline)+slackKiB))
	})
})
