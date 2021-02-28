package env

import (
	"fmt"
	"github.com/Tiemma/image-clone-controller/pkg/errors"
	"os"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	NamespacesToSkip = "NAMESPACES_TO_SKIP"
	DelayPeriod      = "DELAY_PERIOD"
	IsDevEnv         = "IS_DEV_ENV"
	Kubeconfig       = "KUBECONFIG"
	RepoURL          = "REPO_URL"
	DockerConfig     = "DOCKER_CONFIG"
)

var (
	logger                  = ctrl.Log.WithValues("pkg", "env")
	defaultNamespacesToSkip = []string{"kube-system"}

	RequiredEnvVariables = []string{RepoURL, DockerConfig}
)

func splitCommaSeparatedString(str string) []string {
	strs := strings.Split(strings.ReplaceAll(str, " ", ""), ",")
	var res []string
	for _, s := range strs {
		if s != "" {
			res = append(res, s)
		}
	}

	return res
}

func getSkippableNamespaces() []string {
	// We can ignore duplicates as the sample set is too small to
	// bring out any performance issues
	skippableNamespaces := splitCommaSeparatedString(os.Getenv(NamespacesToSkip))
	skippableNamespaces = append(skippableNamespaces, defaultNamespacesToSkip...)

	return skippableNamespaces
}

func IsSkippableNamespace(workloadKind string, namespace string) bool {
	for _, ns := range getSkippableNamespaces() {
		if ns == namespace {
			logger.Info(fmt.Sprintf("Config set to ignore workloads of type %s in the %s namespace, skipping...", workloadKind, namespace))
			return true
		}
	}

	return false
}

func MustValidateRequiredEnvsExist() {
	for _, key := range RequiredEnvVariables {
		if os.Getenv(key) == "" {
			errors.HandleErr(fmt.Errorf("%s env key must be set", key))
		}
	}
}
