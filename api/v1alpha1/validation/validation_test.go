package validation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("ValidateValue", func() {
	var val v1alpha1.Validation
	correct := "123456"
	shortString := "12"
	longString := "1234567890"
	invalidString := "123abc"

	reg := v1alpha1.ValidationRegex("^[0-9]*$")

	BeforeEach(func() {
		val = v1alpha1.Validation{
			MinLength: pointer.Int32Ptr(3),
			MaxLength: pointer.Int32Ptr(9),
			Regex:     &reg,
		}
	})

	Context("Validating a correct string", func() {
		It("should return no errors", func() {
			errs := ValidateValue(correct, field.NewPath("test"), val)
			Expect(errs).To(BeNil())
		})
	})

	Context("Validating a too short string", func() {
		It("should return one error", func() {
			errs := ValidateValue(shortString, field.NewPath("test"), val)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("test: Invalid value: \"12\": Value below MinLength"))
		})
	})

	Context("Validating a too long string", func() {
		It("should return one error", func() {
			errs := ValidateValue(longString, field.NewPath("test"), val)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("test: Invalid value: \"1234567890\": Value above MaxLength"))
		})
	})

	Context("Validating an invalid string", func() {
		It("should return one error", func() {
			errs := ValidateValue(invalidString, field.NewPath("test"), val)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("test: Invalid value: \"123abc\": Value does not match regex pattern"))
		})
	})
})
