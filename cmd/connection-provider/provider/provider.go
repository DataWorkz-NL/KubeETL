package provider

import (
	"context"
	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/listers"
	"github.com/dataworkz/kubeetl/pkg/util"
	corev1 "k8s.io/api/core/v1"
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
		return err
	}

	secret := corev1.Secret{}
	err = cp.client.Get(ctx, wf.SecretName(), &secret)
	if err != nil {
		return err
	}

	err = cp.populateSecret(ctx, &secret, wf)
	if err != nil {
		return err
	}

	err = cp.client.Update(ctx, &secret)
	if err != nil {
		return err
	}

	return nil
}

// populateSecret renders the template for each InjectableValue in a Workflow and adds the result to the provided secret
// using the name of the InjectableValue as a key
func (cp *connectionProvider) populateSecret(ctx context.Context, secret *corev1.Secret, wf *v1alpha1.Workflow) error {
	for _, iv := range wf.Spec.InjectableValues {
		conn, err := cp.connectionLister.Find(ctx, wf.Namespace, iv.ConnectionRef.Name)
		if err != nil {
			return err
		}

		credValues := make(map[string]string, len(conn.Spec.Credentials))

		for name := range conn.Spec.Credentials {
			credReader := util.NewCredentialReader(cp.client, conn)
			data, err := credReader.ReadValue(ctx, name)
			if err != nil {
				return err
			}

			credValues[name] = data
		}

		content, err := iv.Content.Render(credValues)
		if err != nil {
			return err
		}

		secret.StringData[iv.Name] = content
	}

	return nil
}
