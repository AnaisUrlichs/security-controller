/*
Copyright 2023 AnaisUrlichs.
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

package apps

import (
	"context"
	"time"

	kapps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	deployment := &kapps.Deployment{}

	// Get list of deployments with annotation
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Deployment", "name", deployment.Name, "namespace", deployment.Namespace, "annotations", deployment.Annotations)

	lastUpdated, has := deployment.GetAnnotations()["anaisurl.com/last-updated"]
	val, ok := deployment.GetAnnotations()["anaisurl.com/misconfiguration"]

	// check if lastUpdated is more than 1 minutes
	if ok && val == "false" && has {

		lastUpdatedTime, err := time.Parse(time.RFC3339, lastUpdated)

		if time.Now().Sub(lastUpdatedTime) > 5*time.Minute {
			val = "true"
			// Update deployment
			deployment.SetAnnotations(map[string]string{"anaisurl.com/misconfiguration": val})
			deployment.Annotations["anaisurl.com/last-updated"] = time.Now().Format(time.RFC3339)

			err := r.Client.Update(ctx, deployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		if err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kapps.Deployment{}).
		Complete(r)
}
