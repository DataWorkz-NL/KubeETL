package provider

import (
	"context"
	"fmt"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/listers"
	"github.com/dataworkz/kubeetl/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConnectionProvider interface {
	ProvideWorkflowSecret(workflowName, workflowNamespace string) error
}

func NewConnectionProvider(client client.Client) ConnectionProvider {
	return &connectionProvider{
		client:           client,
		workflowLister:   listers.NewWorkflowLister(client),
		connectionLister: listers.NewConnectionLister(client),
	}
}

type connectionProvider struct {
	client           client.Client
	workflowLister   listers.WorkflowLister
	connectionLister listers.ConnectionLister
}

func (cp *connectionProvider) ProvideWorkflowSecret(workflowName, workflowNamespace string) error {
	ctx := context.Background()
	wf, err := cp.workflowLister.Find(ctx, workflowNamespace, workflowName)
	if err != nil {
		return fmt.Errorf("failed to find workflow with name %s: %w", workflowName, err)
	}

	secret := corev1.Secret{}
	key := types.NamespacedName{
		Name:      v1alpha1.ConnectionSecret(wf.Name, wf.Namespace).Name,
		Namespace: v1alpha1.ConnectionSecret(wf.Name, wf.Namespace).Namespace,
	}
	if err := cp.client.Get(ctx, key, &secret); err != nil {
		return fmt.Errorf("failed to find connection secret with name %s: %w", key.Name, err)
	}

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
func (cp *connectionProvider) populateSecret(ctx context.Context, secret *corev1.Secret, wf *v1alpha1.Workflow) error {
	secret.StringData = make(map[string]string)
	for _, iv := range wf.Spec.InjectableValues {
		conn, err := cp.connectionLister.Find(ctx, wf.Namespace, iv.ConnectionRef.Name)
		if err != nil {
			return fmt.Errorf("failed to find Connection %s: %w", iv.ConnectionRef.Name, err)
		}

		credValues := make(map[string]string, len(conn.Spec.Credentials))

		for name := range conn.Spec.Credentials {
			credReader := util.NewCredentialReader(cp.client, conn)
			data, err := credReader.ReadValue(ctx, name)
			if err != nil {
				return fmt.Errorf("failed to read credential value %s in Connection %s: %w", name, iv.ConnectionRef.Name, err)
			}

			credValues[name] = data
		}

		content, err := iv.Content.Render(credValues)
		if err != nil {
			return fmt.Errorf("failed to render content for InjectableValue %s: %w", iv.Name, err)
		}

		secret.StringData[iv.Name] = content
	}

	return nil
}
