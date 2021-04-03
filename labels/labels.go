package labels

func HasLabel(labels map[string]string, expected string) bool {
	_, ok := labels[expected]
	return ok
}

func GetLabelValue(labels map[string]string, expected string) string {
	if !HasLabel(labels, expected) {
		return ""
	}

	val, _ := labels[expected]
	return val
}