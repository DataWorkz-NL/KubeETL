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

// ValidateConnection validates whether a v1alpha1.Connection adheres to the definition
// of its type defined by a v1alpha1.ConnectionType
func ValidateConnection(con v1alpha1.Connection, conType v1alpha1.ConnectionType) field.ErrorList {
	// Transform to fieldmap for quick lookup
	fieldMap := make(map[string]*v1alpha1.CredentialFieldSpec)
	for _, credField := range conType.Spec.Fields {
		fieldMap[credField.Name] = &credField
	}

	// Perform field validation
	var errList field.ErrorList
	for k, v := range con.Spec.Credentials {
		credField := fieldMap[k]
		path := field.NewPath("spec").Child("credentials").Child(k)
		if credField == nil {
			if !conType.Spec.AllowExtraFields {
				err := field.Invalid(path, v, "ConnectionType does not allow extra fields")
				errList = append(errList, err)
			}
			continue
		}

		// Perform validation, not possible with reference valueFrom
		if v.Value != "" {
			errs := ValidateValue(v.Value, path, *credField.Validation)
			errList = append(errList, errs...)
		}
	}

	return errList
}
