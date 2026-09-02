package test_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Alloc", func() {
	It("should return a writable buffer of the requested size", func() {
		Expect(runMake(recipe("[$(fixture-alloc 8)]"))).To(Equal("[xxxxxxxx]"))
	})

	It("should return nil rather than a buffer for a non-positive size", func() {
		// Both sizes come back as a zero-length slice, which is what the
		// fixture measures; a nil pointer wrapped in a slice would panic on
		// the first write instead.
		Expect(runMake(recipe("[$(fixture-alloc-empty )]"))).To(Equal("[0 0]"))
	})

	It("should hand back memory make owns, not the Go heap", func() {
		// A large buffer written end to end and read back intact is the only
		// observable difference between a real gmk_alloc and a Go allocation
		// make would later try to free.
		const size = 1 << 16

		dir := GinkgoT().TempDir()
		runMakeWith(makeOpts{
			Dir:  dir,
			Body: "$(file >out.txt,$(fixture-alloc 65536))\nall:;@true\n",
		})

		got, err := os.ReadFile(filepath.Join(dir, "out.txt"))
		Expect(err).NotTo(HaveOccurred())

		Expect(strings.TrimSuffix(string(got), "\n")).To(Equal(strings.Repeat("x", size)))
	})
})

var _ = Describe("Free", func() {
	It("should ignore a nil or empty buffer", func() {
		Expect(runMake(recipe("[$(fixture-free-empty )]"))).To(Equal("[ok]"))
	})
})
