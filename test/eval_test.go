package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Eval", func() {
	It("should define a variable from the plugin's setup", func() {
		Expect(runMake(recipe("$(FIXTURE_FROM_PLUGIN)"))).To(Equal("yes"))
	})

	It("should define a rule from the plugin's setup", func() {
		out := runMakeWith(makeOpts{
			Target: "fixture-target",
			Body:   recipe("unreachable"),
		})

		Expect(out).To(Equal("from-plugin-rule"))
	})

	It("should define a variable from inside an expansion", func() {
		body := "X := $(fixture-eval RUNTIME_VAR := ok)\n" + recipe("$(RUNTIME_VAR)")

		Expect(runMake(body)).To(Equal("ok"))
	})

	It("should define a rule with a recipe from inside an expansion", func() {
		body := `X := $(fixture-eval evaled:\n\t@echo from-runtime-eval)` +
			"\nall: evaled\n"

		Expect(runMake(body)).To(Equal("from-runtime-eval"))
	})
})

var _ = Describe("Eval with a FileLocation", func() {
	// The fixture attributes its text to phony.mk line 42, a file that does
	// not exist, so the location in the diagnostic can only have come from the
	// gmk_floc the binding built.
	const garbage = "X := $(fixture-eval-at garbage line)\nall:;@true\n"

	It("should attribute a parse error to the location it was given", func() {
		out := runMakeFail(makeOpts{Body: garbage})

		Expect(out).To(ContainSubstring("phony.mk:42"))
	})

	It("should attribute the error to the real makefile when the location is nil", func() {
		out := runMakeFail(makeOpts{
			Body: "X := $(fixture-eval garbage line)\nall:;@true\n",
		})

		Expect(out).NotTo(ContainSubstring("phony.mk"))
		Expect(out).To(ContainSubstring("Makefile:"))
	})
})
