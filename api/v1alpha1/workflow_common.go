package v1alpha1

import (
	"crypto/md5"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NameWithHash returns the workflow name with an md5-hash added as a suffix.
// This is used prevent naming conflicts when creating Workflow-related resources,
// such as Connection Secrets
func NameWithHash(name string) string {
	m := md5.New()
	h := m.Sum([]byte(name))

	return fmt.Sprintf("%s-%x", name, h)
}

func ConnectionVolumeName(parentResourceName string) string {
	return NameWithHash(parentResourceName)
}

func ConnectionSecret(parentResourceName, namespace string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      NameWithHash(parentResourceName),
		},
	}
}
