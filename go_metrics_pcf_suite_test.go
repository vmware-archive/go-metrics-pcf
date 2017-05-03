package pcfmetrics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoMetricsPcf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoMetricsPcf Suite")
}
