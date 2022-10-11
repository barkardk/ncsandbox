/*
Copyright 2022.

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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	addPodNameLabelAnnotation = "barkardk.info/add-pod-name-label"
	podNameLabel              = "barkardk.info/pod-name"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	// Fetch the pod from the kubernetes api
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			// ignore not-found errors
			return ctrl.Result{}, nil
		}
		l.Error(err, "unable to fetch pod")
		return ctrl.Result{}, err

	}

	// Add or remove the label
	labelShouldBePresent := pod.Annotations[addPodNameLabelAnnotation] == "true"
	labelIsPresent := pod.Labels[podNameLabel] == pod.Name
	if labelShouldBePresent == labelIsPresent {
		// the pod desired and actual state matches and no more is required by the operator
		l.Info("no update required")
		return ctrl.Result{}, nil
	}

	if labelShouldBePresent {
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels[podNameLabel] = pod.Name
		l.Info("adding label")
	} else {
		// if the labels should not be set but is , remove it
		delete(pod.Labels, podNameLabel)
		l.Info("removing label")
	}

	// Update the pod in the kubernetes api
	if err := r.Update(ctx, &pod); err != nil {
		if apierrors.IsConflict(err) {
			// the pod has beeen updated since we read it
			// requeue the pod and reconcile again
			return ctrl.Result{Requeue: true}, nil
		}
		l.Error(err, "unable to update pod")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
