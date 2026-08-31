package test_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGnumake(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gnumake Suite")
}
