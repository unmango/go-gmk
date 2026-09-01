package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Expand", func() {
	It("should expand a variable reference the way make would", func() {
		body := "GREETING := hi\n" + recipe("$(fixture-expand $$(GREETING))")

		Expect(runMake(body)).To(Equal("hi"))
	})

	It("should expand a function call", func() {
		Expect(runMake(recipe("$(fixture-expand $$(subst a,b,aaa))"))).To(Equal("bbb"))
	})

	It("should expand a reference through another variable", func() {
		body := "INNER := deep\nOUTER = $(INNER)\n" + recipe("$(fixture-expand $$(OUTER))")

		Expect(runMake(body)).To(Equal("deep"))
	})

	It("should expand an undefined variable to nothing", func() {
		Expect(runMake(recipe("[$(fixture-expand $$(NOPE))]"))).To(Equal("[]"))
	})

	It("should expand text with no references to itself", func() {
		Expect(runMake(recipe("$(fixture-expand plain)"))).To(Equal("plain"))
	})

	It("should expand empty text to nothing", func() {
		Expect(runMake(recipe("[$(fixture-expand )]"))).To(Equal("[]"))
	})

	It("should expand a call back into the plugin", func() {
		// make is inside a plugin function, expanding text that calls another
		// plugin function. The bridge has to be reentrant for this to return.
		Expect(runMake(recipe("$(fixture-expand $$(fixture-upper abc))"))).To(Equal("ABC"))
	})
})
