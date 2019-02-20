package grouter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoUrlRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoUrlRouter Suite")
}
