package labels

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringSet", func() {
	var ss StringSet
	var emptySs StringSet
	BeforeEach(func() {
		ss = StringSet("").Add("1").Add("2")
		emptySs = StringSet("")
	})

	It("Should be possible to add elements to an existing StringSet", func() {

		Expect(string(ss)).To(Equal("1,2"))
	})

	It("Should be possible to split a StringSet into a slice", func() {
		res := ss.Split()
		Expect(res).To(Equal([]string{"1", "2"}))
	})

	It("Should be possible to check whether an element exists", func() {
		Expect(ss.Contains("1")).To(BeTrue())
		Expect(ss.Contains("3")).To(BeFalse())
	})

	It("Should be possible to remove elements from a StringSet", func() {
		Expect(ss.Remove("3")).To(Equal(ss))
		Expect(ss.Remove("1")).To(Equal(NewStringSet("2")))
	})

	It("Should be possible to check a StringSet for emptyness", func() {
		Expect(ss.IsEmpty()).To(BeFalse())
		Expect(emptySs.IsEmpty()).To(BeTrue())
	})
})