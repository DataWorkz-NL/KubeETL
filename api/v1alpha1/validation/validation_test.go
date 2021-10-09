package validation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
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

	BeforeEach(func() {
		reg := v1alpha1.ValidationRegex("^[0-9]*$")
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

var _ = Describe("ValidateConnection", func() {
	var conType v1alpha1.ConnectionType

	BeforeEach(func() {
		reg := v1alpha1.ValidationRegex("^[0-9]*$")
		val := v1alpha1.Validation{
			MinLength: pointer.Int32Ptr(3),
			MaxLength: pointer.Int32Ptr(9),
			Regex:     &reg,
		}

		conType = v1alpha1.ConnectionType{
			Spec: v1alpha1.ConnectionTypeSpec{
				Name: "Test",
				Fields: []v1alpha1.CredentialFieldSpec{
					{
						Name:       "username",
						Validation: &val,
					},
					{
						Name:      "password",
						Sensitive: true,
					},
				},
			},
		}
	})

	Context("Validating a correct v1alpha1.Connection", func() {
		con := v1alpha1.Connection{
			Spec: v1alpha1.ConnectionSpec{
				Type: "Test",
				Credentials: v1alpha1.Credentials{
					"username": v1alpha1.Value{
						Value: "123456",
					},
					"password": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "password",
							},
						},
					},
				},
			},
		}
		It("should return no errors", func() {
			errs := ValidateConnection(con, conType)
			Expect(errs).To(BeNil())
		})
	})

	Context("Validating a v1alpha1.Connection with a disallowed extra field", func() {
		con := v1alpha1.Connection{
			Spec: v1alpha1.ConnectionSpec{
				Type: "Test",
				Credentials: v1alpha1.Credentials{
					"username": v1alpha1.Value{
						Value: "123456",
					},
					"password": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "password",
							},
						},
					},
					"nonsense": v1alpha1.Value{
						Value: "nonsense",
					},
				},
			},
		}
		It("should return one error indicating there is an invalid field", func() {
			errs := ValidateConnection(con, conType)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("spec.credentials.nonsense: Invalid value: v1alpha1.Value{Value:\"nonsense\", ValueFrom:(*v1alpha1.ValueSource)(nil)}: ConnectionType does not allow extra fields"))
		})
	})

	Context("Validating a v1alpha1.Connection with an invalid value", func() {
		con := v1alpha1.Connection{
			Spec: v1alpha1.ConnectionSpec{
				Type: "Test",
				Credentials: v1alpha1.Credentials{
					"username": v1alpha1.Value{
						Value: "12",
					},
					"password": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "password",
							},
						},
					},
				},
			},
		}
		It("should return one error indicating the value is invalid", func() {
			errs := ValidateConnection(con, conType)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("spec.credentials.username: Invalid value: \"12\": Value below MinLength"))
		})
	})

	Context("Validating a v1alpha1.Connection with a plain text secret", func() {
		con := v1alpha1.Connection{
			Spec: v1alpha1.ConnectionSpec{
				Type: "Test",
				Credentials: v1alpha1.Credentials{
					"username": v1alpha1.Value{
						Value: "123456",
					},
					"password": v1alpha1.Value{
						Value: "secret",
					},
				},
			},
		}
		It("should return one error indicating the field is sensitive", func() {
			errs := ValidateConnection(con, conType)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("spec.credentials.password: Invalid value: v1alpha1.Value{Value:\"secret\", ValueFrom:(*v1alpha1.ValueSource)(nil)}: Field is sensitive, only SecretKeyRef is allowed"))
		})
	})
})

var _ = Describe("ValidateDataSet", func() {
	var dsType v1alpha1.DataSetType

	BeforeEach(func() {
		reg := v1alpha1.ValidationRegex("^[a-zA-Z0-9_]*$")
		val := v1alpha1.Validation{
			MinLength: pointer.Int32Ptr(3),
			MaxLength: pointer.Int32Ptr(9),
			Regex:     &reg,
		}

		dsType = v1alpha1.DataSetType{
			Spec: v1alpha1.DataSetTypeSpec{
				MetadataFields: v1alpha1.MetadataValidation{
					Fields: []v1alpha1.CredentialFieldSpec{
						{
							Name:       "db_name",
							Validation: &val,
						},
						{
							Name:      "metadata_db_url",
							Sensitive: true,
						},
					},
					AllowExtraFields: false,
				},
			},
		}
	})

	Context("Validating a correct v1alpha1.DataSet", func() {
		ds := v1alpha1.DataSet{
			Spec: v1alpha1.DataSetSpec{
				Type: "Test",
				Metadata: v1alpha1.Credentials{
					"db_name": v1alpha1.Value{
						Value: "my_db",
					},
					"metadata_db_url": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "metadata_db_url",
							},
						},
					},
				},
			},
		}
		It("should return no errors", func() {
			errs := ValidateDataSet(ds, dsType)
			Expect(errs).To(BeNil())
		})
	})

	Context("Validating a v1alpha1.DataSet with a disallowed extra field", func() {
		ds := v1alpha1.DataSet{
			Spec: v1alpha1.DataSetSpec{
				Type: "Test",
				Metadata: v1alpha1.Credentials{
					"db_name": v1alpha1.Value{
						Value: "my_db",
					},
					"metadata_db_url": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "metadata_db_url",
							},
						},
					},
					"nonsense": v1alpha1.Value{
						Value: "nonsense",
					},
				},
			},
		}
		It("should return one error indicating there is an invalid field", func() {
			errs := ValidateDataSet(ds, dsType)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("spec.metadata.nonsense: Invalid value: v1alpha1.Value{Value:\"nonsense\", ValueFrom:(*v1alpha1.ValueSource)(nil)}: DataSetType does not allow extra fields"))
		})
	})

	Context("Validating a v1alpha1.DataSet with a allowed extra field", func() {
		ds := v1alpha1.DataSet{
			Spec: v1alpha1.DataSetSpec{
				Type: "Test",
				Metadata: v1alpha1.Credentials{
					"db_name": v1alpha1.Value{
						Value: "my_db",
					},
					"metadata_db_url": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "metadata_db_url",
							},
						},
					},
					"nonsense": v1alpha1.Value{
						Value: "nonsense",
					},
				},
			},
		}
		It("should return one error indicating there is an invalid field", func() {
			dsType.Spec.MetadataFields.AllowExtraFields = true
			errs := ValidateDataSet(ds, dsType)
			Expect(errs).To(BeNil())
		})
	})

	Context("Validating a v1alpha1.DataSet with an invalid value", func() {
		ds := v1alpha1.DataSet{
			Spec: v1alpha1.DataSetSpec{
				Type: "Test",
				Metadata: v1alpha1.Credentials{
					"db_name": v1alpha1.Value{
						Value: "db",
					},
					"metadata_db_url": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &v1.SecretKeySelector{
								Key: "metadata_db_url",
							},
						},
					},
				},
			},
		}
		It("should return one error indicating the value is invalid", func() {
			errs := ValidateDataSet(ds, dsType)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("spec.metadata.db_name: Invalid value: \"db\": Value below MinLength"))
		})
	})

	Context("Validating a v1alpha1.DataSet with a plain text secret", func() {
		ds := v1alpha1.DataSet{
			Spec: v1alpha1.DataSetSpec{
				Type: "Test",
				Metadata: v1alpha1.Credentials{
					"db_name": v1alpha1.Value{
						Value: "my_db",
					},
					"metadata_db_url": v1alpha1.Value{
						Value: "secret",
					},
				},
			},
		}
		It("should return one error indicating the field is sensitive", func() {
			errs := ValidateDataSet(ds, dsType)
			Expect(errs).To(Not(BeNil()))
			Expect(errs.ToAggregate().Error()).To(Equal("spec.metadata.metadata_db_url: Invalid value: v1alpha1.Value{Value:\"secret\", ValueFrom:(*v1alpha1.ValueSource)(nil)}: Field is sensitive, only SecretKeyRef is allowed"))
		})
	})
})
