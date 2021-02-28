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

package controllers

import (
	"context"
	"github.com/Tiemma/image-clone-controller/pkg/metrics"
	"time"

	"github.com/Tiemma/image-clone-controller/pkg/docker"
	"github.com/Tiemma/image-clone-controller/pkg/env"
	"github.com/Tiemma/image-clone-controller/pkg/errors"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DaemonSetReconciler reconciles a DaemonSet object
type DaemonSetReconciler struct {
	client.Client
	Log               logr.Logger
	Scheme            *runtime.Scheme
	KubeServerVersion string
	RetryDelay        time.Duration
}

// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get;update;patch

func (r *DaemonSetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	_ = r.Log.WithValues("daemonSet", req.NamespacedName)

	daemonSet := &appsv1.DaemonSet{}

	if err := r.Client.Get(ctx, req.NamespacedName, daemonSet); err != nil {
		metrics.UpdateFailedImageClonesMetric(daemonSet.Name, daemonSet.Namespace, daemonSet.Kind, "", errors.SpecGet)

		return ctrl.Result{
			RequeueAfter: r.RetryDelay,
		}, errors.ErrorGettingResource("DaemonSet", err)
	}

	if env.IsSkippableNamespace(daemonSet.Kind, daemonSet.Namespace) {
		return ctrl.Result{}, nil
	}

	image, errType := docker.MustCacheAndModifyPodImage(&daemonSet.Spec.Template.Spec, r.KubeServerVersion)
	if errType != "" {
		metrics.UpdateFailedImageClonesMetric(daemonSet.Name, daemonSet.Namespace, daemonSet.Kind, image, errType)
		return ctrl.Result{
			RequeueAfter: r.RetryDelay,
		}, errors.ErrorCloningImage(image, errType)
	}

	if err := r.Client.Update(ctx, daemonSet); err != nil {
		metrics.UpdateFailedImageClonesMetric(daemonSet.Name, daemonSet.Namespace, daemonSet.Kind, "", errors.SpecUpdate)
		return ctrl.Result{
			RequeueAfter: r.RetryDelay,
		}, errors.ErrorUpdatingResource(daemonSet.Name, daemonSet.Namespace, daemonSet.Kind, err)
	}

	return ctrl.Result{}, nil
}

func (r *DaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Complete(r)
}
