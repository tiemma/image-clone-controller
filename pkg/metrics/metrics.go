package metrics

import (
	"github.com/Tiemma/image-clone-controller/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	ctrlMetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ImageCloneTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "image_clone_total",
			Help: "Number of unique images successfully cloned",
		},
	)

	failedImageClones = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "image_clone_failures",
			Help: "Number of failed image clones",
		},
		[]string{"name", "namespace", "kind", "image", "err_type"},
	)
)

func UpdateFailedImageClonesMetric(name, namespace, kind, image string, errType errors.ErrType) {
	failedImageClones.WithLabelValues(name, namespace, kind, image, string(errType)).Add(1)
}

func Init() {
	// Register custom metrics with the global prometheus registry
	ctrlMetrics.Registry.MustRegister(ImageCloneTotal, failedImageClones)
}
