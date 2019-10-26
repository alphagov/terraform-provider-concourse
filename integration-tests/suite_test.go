package integration_tests

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}

var _ = BeforeSuite(SetupSuite)
var _ = AfterSuite(TeardownSuite)
