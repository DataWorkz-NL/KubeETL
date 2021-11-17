package util

import (
	"context"
	"errors"
	"fmt"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CredentialReader interface {
	ReadValue(ctx context.Context, name string) (string, error)
}

type credentialReader struct {
	client     client.Client
	connection *v1alpha1.Connection
}

func NewCredentialReader(client client.Client, connection *v1alpha1.Connection) CredentialReader {
	return &credentialReader{
		client:     client,
		connection: connection,
	}
}

func (cr *credentialReader) ReadValue(ctx context.Context, name string) (string, error) {
	value, ok := cr.connection.Spec.Credentials[name]
	if !ok {
		return "", fmt.Errorf("credential value %s not found in connection %s", name, cr.connection.Name)
	}

	if value.Value != "" {
		return value.Value, nil
	}
	if value.ValueFrom != nil {
		return readValueSource(ctx, cr.client, cr.connection.Namespace, value.ValueFrom)
	}
	return "", fmt.Errorf("credential value %s in connection %s does not contain a value or value source", name, cr.connection.Name)
}

type dataSetReader struct {
	client  client.Client
	dataset *v1alpha1.DataSet
}

func NewDataSetCredentialReader(client client.Client, dataset *v1alpha1.DataSet) CredentialReader {
	return &dataSetReader{
		client:  client,
		dataset: dataset,
	}
}

func (cr *dataSetReader) ReadValue(ctx context.Context, name string) (string, error) {
	value, ok := cr.dataset.Spec.Metadata[name]
	if !ok {
		return "", fmt.Errorf("credential value %s not found in dataset %s", name, cr.dataset.Name)
	}

	if value.Value != "" {
		return value.Value, nil
	}
	if value.ValueFrom != nil {
		return readValueSource(ctx, cr.client, cr.dataset.Namespace, value.ValueFrom)
	}
	return "", fmt.Errorf("credential value %s in dataset %s does not contain a value or value source", name, cr.dataset.Name)
}

func readValueSource(ctx context.Context, cl client.Client, namespace string, valueSource *v1alpha1.ValueSource) (string, error) {
	if valueSource.ConfigMapKeyRef != nil {
		return readConfigMapKey(ctx, cl, namespace, valueSource.ConfigMapKeyRef)
	}
	if valueSource.SecretKeyRef != nil {
		return readSecretKey(ctx, cl, namespace, valueSource.SecretKeyRef)
	}

	return "", errors.New("no configmapkeyref or secretkeyref found")
}

func readSecretKey(ctx context.Context, cl client.Client, namespace string, selector *corev1.SecretKeySelector) (string, error) {
	secret := &corev1.Secret{}
	err := cl.Get(ctx, client.ObjectKey{Name: selector.Name, Namespace: namespace}, secret)
	if err != nil {
		return "", fmt.Errorf("readSecretKey failed: %w", err)
	}

	data, ok := secret.Data[selector.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", selector.Key, selector.Name)
	}

	return string(data), nil
}

func readConfigMapKey(ctx context.Context, cl client.Client, namespace string, selector *corev1.ConfigMapKeySelector) (string, error) {
	cm := &corev1.ConfigMap{}
	err := cl.Get(ctx, client.ObjectKey{Name: selector.Name, Namespace: namespace}, cm)
	if err != nil {
		return "", fmt.Errorf("readConfigMapKey failed: %w", err)
	}

	data, ok := cm.Data[selector.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap %s", selector.Key, selector.Name)
	}

	return data, nil
}
