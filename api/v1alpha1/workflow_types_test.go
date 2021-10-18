package v1alpha1

import (
	"errors"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContentTemplate", func() {
	Context("Rendering a valid template", func() {
		var ct ContentTemplate

		BeforeEach(func() {
			ct = "{{.Value1}} {{.Value2}}"
		})

		It("Should fail if templated keys are missing", func() {
			data := map[string]interface{}{
				"Value1": "foo",
			}

			_, err := ct.Render(data)
			Expect(err).To(HaveOccurred())
			var execErr template.ExecError
			Expect(errors.As(err, &execErr)).To(BeTrue())
			Expect(execErr.Error()).To(Equal(`template: content:1:14: executing "content" at <.Value2>: map has no entry for key "Value2"`))
		})

		It("Should render succesfully with all keys provided", func() {
			data := map[string]interface{}{
				"Value1": "foo",
				"Value2": "bar",
			}

			res, err := ct.Render(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("foo bar"))
		})
	})

	Context("Rendering a nested template", func() {
		var ct ContentTemplate

		BeforeEach(func() {
			ct = "{{.Nested.Value1}} {{.Nested.Value2}}"
		})

		It("Should render succesfully with all keys provided", func() {
			data := map[string]interface{}{
				"Nested": map[string]interface{}{
					"Value1": "foo",
					"Value2": "bar",
				},
			}

			res, err := ct.Render(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("foo bar"))
		})
	})

	Context("Rendering an invalid template", func() {
		// doesn't need to be much more expansive than this
		// this verifies that invalid template errors are passed along
		It("Should fail on an invalid template", func() {
			var ct ContentTemplate = "{{.value1-}"

			_, err := ct.Render(nil)
			Expect(err).To(HaveOccurred())
		})
	})
})
