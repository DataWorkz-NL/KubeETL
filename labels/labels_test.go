package labels

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("HasLabel",
	func(labels map[string]string, key string, hasLabel bool) {
		res := HasLabel(labels, key)
		Expect(res).To(Equal(hasLabel))
	},
	Entry("Empty labels should return false", map[string]string{}, "etl.dataworkz.nl/healthcheck", false),
	Entry("Unmatched labels should return false", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}, "etl.dataworkz.nl/healthcheck", false),
	Entry("Matched labels should return true", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}, "etl.dataworkz.nl/healthcheck", true),
)

var _ = DescribeTable("RemoveLabel",
	func(labels map[string]string, key string, newLabels map[string]string) {
		res := RemoveLabel(labels, key)
		Expect(res).To(Equal(newLabels))
	},
	Entry("Empty labels should return empty labels", map[string]string{}, "etl.dataworkz.nl/healthcheck", map[string]string{}),
	Entry("Unmatched labels should return labels unaltered", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}, "etl.dataworkz.nl/healthcheck", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}),
	Entry("Matched labels should return altered labels", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}, "etl.dataworkz.nl/healthcheck", map[string]string{}),
)

var _ = DescribeTable("GetLabelValue",
	func(labels map[string]string, key string, expectedVal string) {
		res := GetLabelValue(labels, key)
		Expect(res).To(Equal(expectedVal))
	},
	Entry("Empty labels should return empty string", map[string]string{}, "etl.dataworkz.nl/healthcheck", ""),
	Entry("Unmatched labels should return empty string", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}, "etl.dataworkz.nl/healthcheck", ""),
	Entry("Matched labels should return the label value", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}, "etl.dataworkz.nl/healthcheck", "value"),
)

var _ = DescribeTable("AddLabel",
	func(labels map[string]string, key string, value string, expectedLabels map[string]string) {
		res := AddLabel(labels, key, value)
		Expect(res).To(Equal(expectedLabels))
	},
	Entry("Nil labels should return new label set with new labels", nil, "etl.dataworkz.nl/healthcheck", "value", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}),
	Entry("Empty labels should return new label set with new labels", map[string]string{}, "etl.dataworkz.nl/healthcheck", "value", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}),
	Entry("Non-conflicting labels should return new merged labels", map[string]string{"etl.dataworkz.nl/otherlabel": "value"}, "etl.dataworkz.nl/healthcheck", "value", map[string]string{"etl.dataworkz.nl/otherlabel": "value", "etl.dataworkz.nl/healthcheck": "value"}),
	Entry("Matched labels should update the label value", map[string]string{"etl.dataworkz.nl/healthcheck": "value"}, "etl.dataworkz.nl/healthcheck", "value2", map[string]string{"etl.dataworkz.nl/healthcheck": "value2"}),
)
