package webhooks

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestWebhooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhooks Suite")
}

var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	failPolicy := admissionregistrationv1beta1.Fail
	conWebhookPath := "/validate-v1alpha1-connection"
	dsWebhookPath := "/validate-v1alpha1-dataset"

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			ValidatingWebhooks: []client.Object{
				&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-validation-webhook-config",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "ValidatingWebhookConfiguration",
						APIVersion: "admissionregistration.k8s.io/v1beta1",
					},
					Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{
						{
							Name:          "connection.dataworkz.nl",
							FailurePolicy: &failPolicy,
							ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
								Service: &admissionregistrationv1beta1.ServiceReference{
									Name:      "deployment-validation-service",
									Namespace: "default",
									Path:      &conWebhookPath,
								},
							},
							Rules: []admissionregistrationv1beta1.RuleWithOperations{
								{
									Operations: []admissionregistrationv1beta1.OperationType{
										admissionregistrationv1beta1.Create,
										admissionregistrationv1beta1.Update,
									},
									Rule: admissionregistrationv1beta1.Rule{
										APIGroups:   []string{"etl.dataworkz.nl"},
										APIVersions: []string{"v1alpha1"},
										Resources:   []string{"connections"},
									},
								},
							},
						},
						{
							Name:          "dataset.dataworkz.nl",
							FailurePolicy: &failPolicy,
							ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
								Service: &admissionregistrationv1beta1.ServiceReference{
									Name:      "deployment-validation-service",
									Namespace: "default",
									Path:      &dsWebhookPath,
								},
							},
							Rules: []admissionregistrationv1beta1.RuleWithOperations{
								{
									Operations: []admissionregistrationv1beta1.OperationType{
										admissionregistrationv1beta1.Create,
										admissionregistrationv1beta1.Update,
									},
									Rule: admissionregistrationv1beta1.Rule{
										APIGroups:   []string{"etl.dataworkz.nl"},
										APIVersions: []string{"v1alpha1"},
										Resources:   []string{"datasets"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Port:    testEnv.WebhookInstallOptions.LocalServingPort,
		Host:    testEnv.WebhookInstallOptions.LocalServingHost,
		CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
	})

	By("running webhook server")
	err = SetupValidatingConnectionWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = SetupValidatingDataSetWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
