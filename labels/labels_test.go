package labels

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("HasLabel",
	func(labels map[string]string, expected string, hasLabel bool) {
		ok := HasLabel(labels, expected)
		Expect(ok).To(Equal(hasLabel))
	},
	Entry("Empty labels should return false", map[string]string{}, "etl.dataworkz.nl/healthcheck", false),
	Entry("Unmatched labels should return false", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}, "etl.dataworkz.nl/healthcheck", false),
	Entry("Matched labels should return true", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}, "etl.dataworkz.nl/healthcheck", true),
)

var _ = DescribeTable("GetLabelValue",
	func(labels map[string]string, expected string, expectedVal string) {
		val := GetLabelValue(labels, expected)
		Expect(val).To(Equal(expectedVal))
	},
	Entry("Empty labels should return empty string", map[string]string{}, "etl.dataworkz.nl/healthcheck", ""),
	Entry("Unmatched labels should return empty string", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}, "etl.dataworkz.nl/healthcheck", ""),
	Entry("Matched labels should return the label value", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}, "etl.dataworkz.nl/healthcheck", "value"),
)