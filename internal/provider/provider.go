package provider

import (
	"context"
	"fmt"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/listers"
	"github.com/dataworkz/kubeetl/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/sirupsen/logrus"
)

type SecretProvider interface {
	ProvideWorkflowSecret(workflowName, workflowNamespace string) error
}

func NewSecretProvider(client client.Client) SecretProvider {
	return &secretProvider{
		client:           client,
		workflowLister:   listers.NewWorkflowLister(client),
		connectionLister: listers.NewConnectionLister(client),
		datasetLister:    listers.NewDataSetLister(client),
	}
}

type secretProvider struct {
	client           client.Client
	workflowLister   listers.WorkflowLister
	connectionLister listers.ConnectionLister
	datasetLister    listers.DataSetLister
}

func (cp *secretProvider) ProvideWorkflowSecret(workflowName, workflowNamespace string) error {
	ctx := context.Background()
	wf, err := cp.workflowLister.Find(ctx, workflowNamespace, workflowName)
	if err != nil {
		return fmt.Errorf("failed to find workflow with name %s: %w", workflowName, err)
	}

	secret := v1alpha1.ConnectionSecret(wf.Namespace, wf.Name)
	ownerRef := metav1.OwnerReference{
		APIVersion: wf.APIVersion,
		Kind:       wf.Kind,
		UID:        wf.GetUID(),
		Name:       wf.GetName(),
	}
	secret.OwnerReferences = append(secret.OwnerReferences, ownerRef)

	log.Infof("creating connection secret %s/%s", secret.Namespace, secret.Name)
	if err := cp.populateSecret(ctx, &secret, wf); err != nil {
		return fmt.Errorf("failed to populate connection secret: %w", err)
	}

	if err := cp.client.Update(ctx, &secret); err != nil {
		return fmt.Errorf("failed to update connection secret: %w", err)
	}

	return nil
}

// populateSecret renders the template for each InjectableValue in a Workflow and adds the result to the provided secret
// using the name of the InjectableValue as a key
func (sp *secretProvider) populateSecret(ctx context.Context, secret *corev1.Secret, wf *v1alpha1.Workflow) error {
	secret.StringData = make(map[string]string)
	for _, iv := range wf.Spec.InjectableValues {
		if iv.ConnectionRef.Name != "" {
			content, err := sp.renderConnectionValue(ctx, secret, wf, iv)
			if err != nil {
				return err
			}

			secret.StringData[iv.Name] = content
		} else if iv.DataSetRef.Name != "" {
			content, err := sp.renderDataSetValue(ctx, secret, wf, iv)
			if err != nil {
				return err
			}

			secret.StringData[iv.Name] = content
		}
	}

	return nil
}

func (sp *secretProvider) renderConnectionValue(ctx context.Context, secret *corev1.Secret, wf *v1alpha1.Workflow, iv v1alpha1.InjectableValue) (string, error) {
	conn, err := sp.connectionLister.Find(ctx, wf.Namespace, iv.ConnectionRef.Name)
	if err != nil {
		return "", fmt.Errorf("failed to find Connection %s: %w", iv.ConnectionRef.Name, err)
	}

	credValues, err := sp.createCredentialsMap(ctx, conn, iv)
	if err != nil {
		return "", err
	}

	content, err := iv.Content.Render(credValues)
	if err != nil {
		return "", fmt.Errorf("failed to render content for InjectableValue %s: %w", iv.Name, err)
	}

	return content, nil
}

func (sp *secretProvider) createCredentialsMap(ctx context.Context, conn *v1alpha1.Connection, iv v1alpha1.InjectableValue) (map[string]string, error) {
	credValues := make(map[string]string, len(conn.Spec.Credentials))

	for name := range conn.Spec.Credentials {
		credReader := util.NewCredentialReader(sp.client, conn)
		data, err := credReader.ReadValue(ctx, name)
		if err != nil {
			return credValues, fmt.Errorf("failed to read credential value %s in Connection %s: %w", name, iv.ConnectionRef.Name, err)
		}

		credValues[name] = data
	}

	return credValues, nil
}

func (sp *secretProvider) renderDataSetValue(ctx context.Context, secret *corev1.Secret, wf *v1alpha1.Workflow, iv v1alpha1.InjectableValue) (string, error) {
	ds, err := sp.datasetLister.Find(ctx, wf.Namespace, iv.DataSetRef.Name)
	if err != nil {
		return "", fmt.Errorf("failed to find DataSet %s: %w", iv.DataSetRef.Name, err)
	}

	credValues := make(map[string]string, len(ds.Spec.Metadata))

	for name := range ds.Spec.Metadata {
		credReader := util.NewDataSetCredentialReader(sp.client, ds)
		data, err := credReader.ReadValue(ctx, name)
		if err != nil {
			return "", fmt.Errorf("failed to read credential value %s in Connection %s: %w", name, iv.ConnectionRef.Name, err)
		}

		credValues[name] = data
	}

	injectedValues := make(map[string]map[string]string)
	injectedValues["metadata"] = credValues

	if ds.Spec.Connection.ConnectionFrom != nil {
		conn, err := sp.connectionLister.Find(ctx, wf.Namespace, iv.ConnectionRef.Name)
		if err != nil {
			return "", fmt.Errorf("failed to find Connection for DataSet %s: %w", ds.Spec.Connection.ConnectionFrom.Name, err)
		}

		connValues, err := sp.createCredentialsMap(ctx, conn, iv)
		if err != nil {
			return "", err
		}

		injectedValues["connection"] = connValues
	}

	content, err := iv.Content.Render(injectedValues)
	if err != nil {
		return "", fmt.Errorf("failed to render content for InjectableValue %s: %w", iv.Name, err)
	}

	return content, nil
}
