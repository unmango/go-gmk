package test_test

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Arguments", func() {
	It("should pass every argument through, not just the first", func() {
		Expect(runMake(recipe("$(fixture-join a,b,c)"))).To(Equal("a+b+c"))
	})

	It("should pass a single argument through", func() {
		Expect(runMake(recipe("$(fixture-join solo)"))).To(Equal("solo"))
	})

	It("should keep an empty argument in its position", func() {
		// A dropped empty argument would shift c into b's slot rather than
		// showing up as a gap.
		Expect(runMake(recipe("$(fixture-join a,,c)"))).To(Equal("a++c"))
	})

	It("should receive one empty argument for a call that looks empty", func() {
		// make needs whitespace after the name to see a function call at all,
		// so the emptiest call it can make still carries one empty argument
		// and the bridge never sees an argc of zero.
		Expect(runMake(recipe("[$(fixture-count )]"))).To(Equal("[1]"))
	})

	It("should count the arguments make actually split", func() {
		Expect(runMake(recipe("[$(fixture-count a,b,c)]"))).To(Equal("[3]"))
	})

	It("should carry the full vector when the function has no upper bound", func() {
		args := strings.Repeat("a,", maxArgCount-1) + "a"

		Expect(runMake(recipe("[$(fixture-count " + args + ")]"))).
			To(Equal("[" + strconv.Itoa(maxArgCount) + "]"))
	})

	It("should let make reject a call with too few arguments", func() {
		out := runMakeFail(makeOpts{Body: recipe("$(fixture-arity a)")})

		Expect(out).To(ContainSubstring("insufficient number of arguments"))
		Expect(out).To(ContainSubstring("fixture-arity"))
	})

	It("should fold arguments past the maximum into the last one", func() {
		// make stops splitting at maxArgs rather than erroring, so the tail
		// arrives with its commas intact.
		Expect(runMake(recipe("$(fixture-arity a,b,c,d)"))).To(Equal("a+b+c,d"))
	})
})

// maxArgCount is the limit gnumake.h documents for gmk_add_function.
const maxArgCount = 255

var _ = Describe("FuncNoExpand", func() {
	// The results are quoted in the recipe because an unexpanded argument
	// still looks like a variable reference to the shell.
	const body = "V := hi\nall:\n\t@echo '[$(fixture-raw $(V))] [$(fixture-expanded $(V))]'\n"

	It("should hand the function its argument text unexpanded", func() {
		Expect(runMake(body)).To(Equal(`[$(V)] [hi]`))
	})
})
