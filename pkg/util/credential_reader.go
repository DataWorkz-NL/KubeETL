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
		return cr.readValueSource(ctx, value.ValueFrom)
	}
	return "", fmt.Errorf("credential value %s in connection %s does not contain a value or value source", name, cr.connection.Name)
}

func (cr *credentialReader) readValueSource(ctx context.Context, valueSource *v1alpha1.ValueSource) (string, error) {
	if valueSource.ConfigMapKeyRef != nil {
		return cr.readConfigMapKey(ctx, valueSource.ConfigMapKeyRef)
	}
	if valueSource.SecretKeyRef != nil {
		return cr.readSecretKey(ctx, valueSource.SecretKeyRef)
	}

	return "", errors.New("no configmapkeyref or secretkeyref found")
}

func (cr *credentialReader) readSecretKey(ctx context.Context, selector *corev1.SecretKeySelector) (string, error) {
	secret := &corev1.Secret{}
	err := cr.client.Get(ctx, client.ObjectKey{Name: selector.Name, Namespace: cr.connection.Namespace}, secret)
	if err != nil {
		return "", fmt.Errorf("readSecretKey failed: %w", err)
	}

	data, ok := secret.Data[selector.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", selector.Key, selector.Name)
	}

	return string(data), nil
}

func (cr *credentialReader) readConfigMapKey(ctx context.Context, selector *corev1.ConfigMapKeySelector) (string, error) {
	cm := &corev1.ConfigMap{}
	err := cr.client.Get(ctx, client.ObjectKey{Name: selector.Name, Namespace: cr.connection.Namespace}, cm)
	if err != nil {
		return "", fmt.Errorf("readConfigMapKey failed: %w", err)
	}

	data, ok := cm.Data[selector.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap %s", selector.Key, selector.Name)
	}

	return data, nil
}
