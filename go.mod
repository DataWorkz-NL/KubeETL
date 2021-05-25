module github.com/dataworkz/kubeetl

go 1.16

require (
	github.com/argoproj/argo/v2 v2.12.10
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.1
	sigs.k8s.io/controller-tools v0.5.0 // indirect
	sigs.k8s.io/yaml v1.2.0
)
