/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhooks

import (
	"context"
	"fmt"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/api/v1alpha1/validation"
	"github.com/dataworkz/kubeetl/listers"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-v1alpha1-connection,mutating=false,failurePolicy=fail,groups=etl.dataworkz.nl,resources=connections,versions=v1alpha1,name=connection.dataworkz.nl

// SetupValidatingConnectionWebhookWithManager registers the validating web hook for connections with the manager
func SetupValidatingConnectionWebhookWithManager(mgr ctrl.Manager) error {
	client := mgr.GetClient()
	decoder, err := admission.NewDecoder(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("unable to create decoder: %w", err)
	}
	hook := &connectionValidatorHook{
		client:               client,
		decoder:              decoder,
		connectionTypeLister: listers.NewConnectionTypeLister(client),
	}

	hookserver := mgr.GetWebhookServer()
	hookserver.Register("/validate-v1alpha1-connection", &admission.Webhook{Handler: hook})
	return nil
}

type connectionValidatorHook struct {
	client               client.Client
	decoder              *admission.Decoder
	connectionTypeLister listers.ConnectionTypeLister
}

func (hook *connectionValidatorHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.Log.WithName("webhooks").WithName("connection-validator-hook")
	log.Info("Admission webhook request")
	con := v1alpha1.Connection{}
	if err := hook.decoder.Decode(req, &con); err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to decode admission request: %w", err))
	}

	conType, err := hook.connectionTypeLister.Find(ctx, req.Namespace, con.Spec.Type)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to list ConnectionType: %w", err))
	}

	errs := validation.ValidateConnection(con, *conType)
	if errs != nil {
		return admission.Errored(http.StatusBadRequest, errs.ToAggregate())
	}

	return admission.Allowed("valid Connection resource passed to the API")
}
