package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Error",
	func(key, want string) {
		out := runMakeFail(makeOpts{Body: recipe("$(fixture-error " + key + ")")})

		Expect(out).To(ContainSubstring("Makefile:4: *** " + want + ".  Stop."))
		Expect(out).NotTo(ContainSubstring("after-error"))
	},
	Entry("with plain text", "plain", "something went wrong"),
	Entry("with a reference, left unexpanded", "dollar", "costs $(HOME) and $$"),
	Entry("with an unbalanced parenthesis", "paren", "unbalanced ) here ("),
	Entry("with a comma", "comma", "a, b"),
)

var _ = DescribeTable("Warning",
	func(key, want string) {
		out := runMakeWith(makeOpts{Body: recipe("$(fixture-warning " + key + ")")})

		Expect(out).To(ContainSubstring("Makefile:4: " + want))
		Expect(out).To(ContainSubstring("after-warning"))
	},
	Entry("with plain text", "plain", "something went wrong"),
	Entry("with a reference, left unexpanded", "dollar", "costs $(HOME) and $$"),
	Entry("with an unbalanced parenthesis", "paren", "unbalanced ) here ("),
)
