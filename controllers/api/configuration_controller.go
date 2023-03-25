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

package api

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/AnaisUrlichs/security-controller/apis/api/v1alpha1"
	kapps "k8s.io/api/apps/v1"
)

// ConfigurationReconciler reconciles a Configuration object
type ConfigurationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.core.anaisurl.com,resources=configurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.core.anaisurl.com,resources=configurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.core.anaisurl.com,resources=configurations/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Configuration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	log := r.Log.WithValues("simpleDeployment", req.NamespacedName)

	var misconfDeployment v1alpha1.Configuration
	if err := r.Get(ctx, req.NamespacedName, &misconfDeployment); err != nil {
		log.Error(err, "unable to fetch Misconfiguratin-labelled Deployment")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add the Deployment spec defined to the cluster
	deployment := &kapps.Deployment{}

	// Set the information you care about
	deployment.Spec.Template.Spec.Containers[0].Ports[0].HostPort = misconfDeployment.Spec.ContainerPort

	// Reconcilation when changes occur to the deployment
	// if err := controllerutil.SetControllerReference(&misconfDeployment, deployment, r.scheme); err != nil {
	//    return ctrl.Result{}, err
	//}

	// create deployment if it does not exist
	// update deployment if it changes
	foundDeployment := &kapps.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		l.V(1).Info("Creating Deployment", "deployment", deployment.Name)
		err = r.Create(ctx, deployment)
	} else if err == nil {
		if foundDeployment.Spec.Replicas != deployment.Spec.Replicas {
			foundDeployment.Spec.Replicas = deployment.Spec.Replicas
			l.V(1).Info("Updating Deployment", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
		}
	}

	config := &v1alpha1.Configuration{}
	errt := r.Get(ctx, req.NamespacedName, config)

	if errt != nil {
		l.Error(err, "unable to fetch Configuration")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Reconciling Configuration", "name", config.Name, "namespace", config.Namespace, "labels", config.Labels)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Configuration{}).
		Owns(&kapps.Deployment{}).
		Complete(r)
}
