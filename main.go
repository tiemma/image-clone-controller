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

package main

import (
	"flag"
	"fmt"
	"github.com/Tiemma/image-clone-controller/controllers"
	"github.com/Tiemma/image-clone-controller/pkg/env"
	"github.com/Tiemma/image-clone-controller/pkg/errors"
	"github.com/Tiemma/image-clone-controller/pkg/metrics"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strconv"
	"time"
	// +kubebuilder:scaffold:imports
)

var (
	scheme                         = runtime.NewScheme()
	setupLog                       = ctrl.Log.WithName("setup")
	defaultRetryDelayMinutes int64 = 5
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = appsv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func getDelayPeriod() int64 {
	delayPeriod := os.Getenv(env.DelayPeriod)
	if delayPeriod == "" {
		return defaultRetryDelayMinutes
	}

	delay, err := strconv.Atoi(delayPeriod)
	if err != nil {
		setupLog.Error(err, "specified delay period is not valid")
	}

	if delay <= 0 {
		errors.HandleErr(fmt.Errorf("delay must be positive and greater than 0"))
	}

	return int64(delay)
}

func getKubeConfig() *rest.Config {
	if os.Getenv(env.IsDevEnv) == "true" {
		configPath := filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)
		if kubeConfigPath := os.Getenv(env.Kubeconfig); kubeConfigPath != "" {
			configPath = kubeConfigPath
		}
		kubeconfig, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			setupLog.Error(err, "cannot build client config")
		}

		return kubeconfig
	}

	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		setupLog.Error(err, "cannot obtain cluster config")
	}

	return kubeconfig
}

func getKubeServerVersion() *version.Info {
	clientSet, err := kubernetes.NewForConfig(getKubeConfig())
	if err != nil {
		setupLog.Error(err, "cannot create client from cluster config")
		os.Exit(1)
	}

	kubeVersion, err := clientSet.ServerVersion()
	if err != nil {
		setupLog.Error(err, "cannot obtain server version")
		os.Exit(1)
	}

	return kubeVersion
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "78654e12.bakman.build",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	env.MustValidateRequiredEnvsExist()

	kubeVersion := getKubeServerVersion()
	retryDelayDuration := time.Duration(getDelayPeriod()) * time.Minute
	metrics.Init()

	if err = (&controllers.DaemonSetReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("DaemonSet"),
		Scheme:            mgr.GetScheme(),
		KubeServerVersion: kubeVersion.GitVersion,
		RetryDelay:        retryDelayDuration,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DaemonSet")
		os.Exit(1)
	}

	if err = (&controllers.DeploymentReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Deployment"),
		Scheme:            mgr.GetScheme(),
		KubeServerVersion: kubeVersion.GitVersion,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
