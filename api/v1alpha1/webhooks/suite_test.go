package webhooks

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"path/filepath"
	"testing"
	"time"

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
	caBundle, err := ioutil.ReadFile("certs/ca.pem")
	Expect(err).ShouldNot(HaveOccurred())

	failPolicy := admissionregistrationv1beta1.Fail
	url := "https://127.0.0.1:8443/validate-v1alpha1-connection"

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
								CABundle: caBundle,
								URL:      &url,
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
					},
				},
			},
		},
	}

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})

	By("running webhook server")
	err = SetupValidatingConnectionWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	webhookServer := k8sManager.GetWebhookServer()
	certDir, err := filepath.Abs("./certs")
	Expect(err).ToNot(HaveOccurred())

	webhookServer.CertDir = certDir
	webhookServer.CertName = "server.pem"
	webhookServer.KeyName = "server-key.pem"
	webhookServer.Host = "127.0.0.1"
	webhookServer.Port = 8443

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	d := &net.Dialer{Timeout: time.Second}
	Eventually(func() error {
		conn, err := tls.DialWithDialer(d, "tcp", "127.0.0.1:8443", &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
