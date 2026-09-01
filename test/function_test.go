package test_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddFunction", func() {
	It("should substitute the value the Go function returns", func() {
		Expect(runMake(recipe("$(fixture-upper abc)"))).To(Equal("ABC"))
	})

	It("should return an empty result without upsetting make", func() {
		Expect(runMake(recipe("[$(fixture-upper )]"))).To(Equal("[]"))
	})

	It("should survive a result longer than make's own buffer", func() {
		Expect(runMake(recipe("$(words $(fixture-repeat 4096))"))).To(Equal("4096"))
	})

	It("should tell the Go function which name make dispatched under", func() {
		// Both names resolve to the same Go function, so the only way to
		// distinguish them is the name the bridge passes through.
		out := runMake(recipe("$(fixture-name x) $(fixture-alias x)"))

		Expect(out).To(Equal("fixture-name fixture-alias"))
	})

	It("should replace the Go function when a name is registered twice", func() {
		// The fixture registers this name with two different functions. make
		// rejects a duplicate registration, so the second one has to take
		// effect without calling into make again.
		Expect(runMake(recipe("$(fixture-replaced )"))).To(Equal("second"))
	})
})

// A result make hands to a shell is at the mercy of that shell, so these specs
// route it through $(file ...) and read the bytes back instead.
var _ = DescribeTable("A result containing",
	func(key, want string) {
		dir := GinkgoT().TempDir()

		runMakeWith(makeOpts{
			Dir:  dir,
			Body: "$(file >out.txt,$(fixture-special " + key + "))\nall:;@true\n",
		})

		got, err := os.ReadFile(filepath.Join(dir, "out.txt"))
		Expect(err).NotTo(HaveOccurred())

		// $(file ...) appends a newline of its own.
		Expect(string(got)).To(Equal(want + "\n"))
	},
	Entry("a newline", "newline", "a\nb"),
	Entry("a tab", "tab", "a\tb"),
	Entry("a dollar sign", "dollar", "a$b"),
	Entry("a comment character", "hash", "a#b"),
	Entry("a trailing backslash", "backslash", `a\`),
	Entry("multi-byte UTF-8", "utf8", "héllo→ß"),
	Entry("leading and trailing spaces", "spaces", "  padded  "),
	Entry("nothing at all", "empty", ""),
)
