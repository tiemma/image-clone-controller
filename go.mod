module github.com/Tiemma/image-clone-controller

go 1.13

require (
	github.com/go-logr/logr v0.2.0
	github.com/google/go-containerregistry v0.4.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/prometheus/client_golang v1.0.0
	golang.org/x/mod v0.3.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

replace github.com/go-logr/logr v0.2.0 => github.com/go-logr/logr v0.1.0
