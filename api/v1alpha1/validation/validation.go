package validation

import (
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

// ValidateValue validates the string value against a Validation.
// It returns a field.Errorlist based on the path passed.
func ValidateValue(value string, path *field.Path, validation v1alpha1.Validation) field.ErrorList {

	var errList field.ErrorList
	length := int32(len(value))
	if validation.MaxLength != nil {
		if length > *validation.MaxLength {
			apiErr := field.Invalid(path, value, "Value above MaxLength")
			errList = append(errList, apiErr)
		}
	}

	if validation.MinLength != nil {
		if length < *validation.MinLength {
			apiErr := field.Invalid(path, value, "Value below MinLength")
			errList = append(errList, apiErr)
		}
	}

	if validation.Regex != nil {
		matched, err := regexp.Match(string(*validation.Regex), []byte(value))
		if err != nil {
			apiErr := field.InternalError(path, err)
			errList = append(errList, apiErr)
		}

		if !matched {
			apiErr := field.Invalid(path, value, "Value does not match regex pattern")
			errList = append(errList, apiErr)
		}
	}

	return errList
}
