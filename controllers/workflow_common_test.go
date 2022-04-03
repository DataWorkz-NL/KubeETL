package controllers

import (
	"context"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"
)

type workflowTestResources struct {
	connKey   types.NamespacedName
	secretRef v1.LocalObjectReference
	cmRef     v1.LocalObjectReference
}

func beforeEachWorkflowTest(ctx context.Context, resources *workflowTestResources) {
	resources.connKey = types.NamespacedName{
		Name:      randomSuffix("default-connection"),
		Namespace: "default",
	}

	resources.secretRef = v1.LocalObjectReference{
		Name: randomSuffix("default-secret"),
	}

	resources.cmRef = v1.LocalObjectReference{
		Name: randomSuffix("default-cm"),
	}

	connSpec := api.ConnectionSpec{
		Credentials: api.Credentials{
			"host": api.Value{Value: "localhost"},
			"user": api.Value{
				ValueFrom: &api.ValueSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: resources.cmRef,
						Key:                  "user",
					},
				},
			},
			"password": api.Value{
				ValueFrom: &api.ValueSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: resources.secretRef,
						Key:                  "password",
					},
				},
			},
		},
	}

	conn := api.Connection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.connKey.Name,
			Namespace: resources.connKey.Namespace,
		},
		Spec: connSpec,
	}

	Expect(k8sClient.Create(ctx, &conn)).Should(Succeed())
}

func afterEachWorkflowTest(ctx context.Context, resources *workflowTestResources) {
	var conn api.Connection
	Expect(k8sClient.Get(ctx, resources.connKey, &conn)).Should(Succeed())
	Expect(k8sClient.Delete(ctx, &conn)).Should(Succeed())
}
