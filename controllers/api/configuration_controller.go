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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/AnaisUrlichs/security-controller/apis/api/v1alpha1"
	kapps "k8s.io/api/apps/v1"
	kcore "k8s.io/api/core/v1"
)

// ConfigurationReconciler reconciles a Configuration object
type ConfigurationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	reconcileErrorInterval   = 10 * time.Second
	reconcileSuccessInterval = 120 * time.Second
	annotationName           = "anaisurl.com/misconfiguration"
)

// +kubebuilder:rbac:groups=api.core.anaisurl.com,resources=configurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.core.anaisurl.com,resources=configurations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.core.anaisurl.com,resources=configurations/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;list;watch;create;update;patch;delete
//
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info("Reconciling deployments")

	mdConf := &apiv1alpha1.Configuration{}

	if err := r.Client.Get(ctx, req.NamespacedName, mdConf); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info("No Misconfiguration Configuration found.")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, "Failed to get the Misconfiguration Configuration")
		return r.finishReconcile(err, false)
	}

	if !mdConf.ObjectMeta.DeletionTimestamp.IsZero() {
		// Stop reconciliation as the item is being deleted
		return r.finishReconcile(nil, false)
	}

	// Get list of deployments
	deploymentList := &kapps.DeploymentList{}
	var mdDeploymentList []kapps.Deployment

	if err := r.List(ctx, deploymentList); err != nil {
		return r.finishReconcile(err, false)
	}

	cmExists := false
	// Get list of deployments with annotation
	for _, cm := range deploymentList.Items {
		val, ok := cm.GetAnnotations()["anaisurl.com/misconfiguration"]
		if ok && val == "true" {
			cmExists = true
			mdDeploymentList = append(mdDeploymentList, cm)
		}
	}

	if cmExists == true {
		for _, cm := range mdDeploymentList {

			err := r.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, &cm)
			deploymentStatus := cm.Status.Conditions[0].Type

			// Update Deployment Spec
			log.Info("Reconciling deployments" + cm.Name)
			if err != nil && errors.IsNotFound(err) {
				log.V(1).Info("Deplpyment is not found")
				return r.finishReconcile(err, true)
			} else if err == nil && deploymentStatus == "Available" {
				cm.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = mdConf.Spec.ContainerPort
				cm.Spec.Template.Spec.Containers[0].Image = strings.Split(cm.Spec.Template.Spec.Containers[0].Image, ":")[0] + ":" + mdConf.Spec.ImageTag
				cm.Spec.Template.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = &mdConf.Spec.AllowPrivilegeEscalation
				cm.Spec.Template.Spec.Containers[0].SecurityContext.RunAsNonRoot = &mdConf.Spec.RunAsNonRoot
				cm.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = &mdConf.Spec.ReadOnlyRootFilesystem
				cm.Spec.Template.Spec.Containers[0].Resources.Requests[kcore.ResourceCPU] = mdConf.Spec.CPURequests
				cm.Spec.Template.Spec.Containers[0].Resources.Limits[kcore.ResourceCPU] = mdConf.Spec.CPULimits
				cm.Spec.Template.Spec.Containers[0].Resources.Requests[kcore.ResourceMemory] = mdConf.Spec.MemoryRequests
				cm.Spec.Template.Spec.Containers[0].Resources.Limits[kcore.ResourceMemory] = mdConf.Spec.MemoryLimits
				cm.Annotations["anaisurl.com/last-updated"] = time.Now().Format(time.RFC3339)

				val := "false"
				cm.Annotations["anaisurl.com/misconfiguration"] = val

				err := r.Client.Update(ctx, &cm)
				if err != nil {
					r.finishReconcile(err, true)
				}
			}
		}
	} else {
		if err := r.List(ctx, deploymentList); err != nil {
			return r.finishReconcile(err, false)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ConfigurationReconciler) finishReconcile(err error, requeueImmediate bool) (ctrl.Result, error) {
	if err != nil {
		interval := reconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling Deployments with error: %w")
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	interval := reconcileSuccessInterval
	if requeueImmediate {
		interval = 0
	}
	r.Log.Info("Finished Reconciling Deployment")
	return ctrl.Result{Requeue: true, RequeueAfter: interval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Configuration{}).
		Owns(&kapps.Deployment{}).
		Complete(r)
}
