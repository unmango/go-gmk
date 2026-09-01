package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loading the object", func() {
	It("should run a makefile that loads the plugin", func() {
		Expect(runMake(recipe("loaded"))).To(Equal("loaded"))
	})

	It("should fail when the object does not exist", func() {
		out := runMakeFail(makeOpts{
			Load: "/nonexistent.so",
			Body: recipe("unreachable"),
		})

		Expect(out).To(ContainSubstring("nonexistent.so"))
	})

	It("should refuse an object that does not declare GPL compatibility", func() {
		out := runMakeFail(makeOpts{
			Load: noGPLPath,
			Body: recipe("unreachable"),
		})

		Expect(out).To(ContainSubstring("not declared to be GPL compatible"))
	})

	It("should fail the load when setup reports failure", func() {
		out := runMakeFail(makeOpts{
			Load: badSetupPath,
			Body: recipe("unreachable"),
		})

		Expect(out).To(ContainSubstring("failed to load"))
	})

	It("should call setup once no matter how often the object is loaded", func() {
		body := "load " + pluginPath + "\nload " + pluginPath + "\n" +
			recipe("$(fixture-setup-runs )")

		Expect(runMakeWith(makeOpts{NoLoad: true, Body: body})).To(Equal("1"))
	})

	It("should load into a sub-make as its own process", func() {
		out := runMakeWith(makeOpts{
			NoLoad: true,
			Body:   "all:\n\t@$(MAKE) --no-print-directory -f inner.mk inner\n",
			Files: map[string]string{
				"inner.mk": "load " + pluginPath + "\n" +
					"inner:\n\t@echo $(fixture-upper sub) $(fixture-setup-runs )\n",
			},
		})

		Expect(out).To(Equal("SUB 1"))
	})
})
