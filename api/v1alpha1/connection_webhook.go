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

package v1alpha1

import (
	"context"
	"fmt"

	// "k8s.io/apimachinery/pkg/runtime"
	// fieldval "k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	// "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func SetupValidatingConnectionWebhookWithManager(mgr ctrl.Manager) {

}

type connectionValidatorHook struct {
	client  client.Client
	decoder *admission.Decoder
}

func (hook *connectionValidatorHook) Handle(ctx context.Context, req admission.Request) (*admission.Response, error) {
	log := logf.Log.WithName("webhooks").WithName("connection-validator-hook")
	log.Info("Admission webhook request")
	con := &Connection{}
	if err := hook.decoder.Decode(req, con); err != nil {
		return nil, fmt.Errorf("unable to decode admission request: %w", err)
	}

	typeList := &ConnectionTypeList{}
	if err := hook.client.List(ctx, typeList, &client.ListOptions{Namespace: con.GetNamespace()}); err != nil {
		return nil, fmt.Errorf("unable to list ConnectionTypes: %w", err)
	}

	// Find ConnectionType
	// var conType ConnectionType
	// for _, ct := range typeList.Items {
	// 	if ct.Name == con.Spec.Type {
	// 		conType = ct
	// 	}
	// }

	return nil, nil
	//TODO Check required types
}
