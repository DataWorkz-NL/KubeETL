package labels

import "strings"

// StringSet handles comma seperated strings in label values
// and ensures uniqueness of the values
type StringSet string


// NewStringSet creates a new StringSet from
// the given strings.
func NewStringSet(el ...string) StringSet {
	return StringSet(strings.Join(el, ","))
}

// Split returns a slice of strings
func (s StringSet) Split() []string {
	return strings.Split(string(s), ",")
}

// Contains returns whether the StringSet contains
// the given value
func (s StringSet) Contains(val string) bool {
	l := s.Split()
	i := find(l, val)
	return i != -1
}

// Add adds the given value to the StringSet
// if it wasn't already added.
func (s StringSet) Add(val string) StringSet {
	if s.Contains(val) {
		return s
	}

	if string(s) == "" {
		return StringSet(val)
	}

	return StringSet(strings.Join([]string{string(s), val}, ","))
}

// Remove removes the value from the StringSet
func (s StringSet) Remove(val string) StringSet {
	l := s.Split()
	i := find(l, val)
	if i == -1 {
		return s
	}

	return StringSet(strings.Join(append(l[:i], l[i+1:]...), ","))
}

// TODO move to seperate slice utils
func find(l []string, val string) int {
	for i, v := range l {
		if v == val {
			return i
		}
	}

	return -1
}