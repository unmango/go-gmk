package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// AddFunction panics on arguments make would reject, which no makefile can
// trigger. The fixture calls it with each bad argument behind a recover and
// returns the message, so the specs can assert on the guard from outside.
var _ = DescribeTable("AddFunction rejects",
	func(key, want string) {
		Expect(runMake(recipe("'$(fixture-panics " + key + ")'"))).To(Equal(want))
	},
	Entry("an empty name", "empty-name", "gmk: function name must be 1 to 255 bytes"),
	Entry("a name over 255 bytes", "long-name", "gmk: function name must be 1 to 255 bytes"),
	Entry("a nil function", "nil-func", "gmk: function must not be nil"),
	Entry("a negative minimum", "neg-min", "gmk: minArgs must be 0 to 255"),
	Entry("a minimum over 255", "big-min", "gmk: minArgs must be 0 to 255"),
	Entry("a maximum over 255", "big-max", "gmk: maxArgs must be 0 to 255"),
	Entry("a minimum above the maximum", "min-gt-max", "gmk: minArgs must not exceed maxArgs"),
)
