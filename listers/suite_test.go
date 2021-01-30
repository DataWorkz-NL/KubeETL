package listers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestListers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Listers Suite")
}
