package labels

func HasLabel(labels map[string]string, key string) bool {
	_, ok := labels[key]
	return ok
}

func GetLabelValue(labels map[string]string, key string) string {
	if !HasLabel(labels, key) {
		return ""
	}
	return labels[key]
}

func AddLabel(labels map[string]string, key, value string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[key] = value
	return labels
}

func RemoveLabel(labels map[string]string, key string) map[string]string {
	delete(labels, key)
	return labels
}
