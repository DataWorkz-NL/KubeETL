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

// +kubebuilder:webhook:verbs=create;update,path=/validate-v1alpha1-dataset,mutating=false,failurePolicy=fail,groups=etl.dataworkz.nl,resources=datasets,versions=v1alpha1,sideEffects=None,name=dataset.dataworkz.nl,admissionReviewVersions=v1beta1

// SetupValidatingDataSetWebhookWithManager registers the validating web hook for DataSet with the manager
func SetupValidatingDataSetWebhookWithManager(mgr ctrl.Manager) error {
	client := mgr.GetClient()
	decoder, err := admission.NewDecoder(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("unable to create decoder: %w", err)
	}
	hook := &datasetValidatorHook{
		client:            client,
		decoder:           decoder,
		dataSetTypeLister: listers.NewDataSetTypeLister(client),
	}

	hookserver := mgr.GetWebhookServer()
	hookserver.Register("/validate-v1alpha1-dataset", &admission.Webhook{Handler: hook})
	return nil
}

type datasetValidatorHook struct {
	client            client.Client
	decoder           *admission.Decoder
	dataSetTypeLister listers.DataSetTypeLister
}

func (hook *datasetValidatorHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.Log.WithName("webhooks").WithName("validate-dataset")
	log.Info("Admission webhook request")
	ds := v1alpha1.DataSet{}
	if err := hook.decoder.Decode(req, &ds); err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to decode admission request: %w", err))
	}

	dtype := ds.Spec.Type
	dsType, err := hook.dataSetTypeLister.Find(ctx, req.Namespace, dtype)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to list DataSetType: %w", err))
	}

	if dsType == nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("Unknown DataSetType: %v", dtype))
	}

	errs := validation.ValidateDataSet(ds, *dsType)
	if errs != nil {
		return admission.Errored(http.StatusBadRequest, errs.ToAggregate())
	}

	return admission.Allowed("valid DataSet resource passed to the API")
}
