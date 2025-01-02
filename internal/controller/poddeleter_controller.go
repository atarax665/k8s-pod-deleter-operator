/*
Copyright 2025.

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

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	lifetimeLabel = "pod.kubernetes.io/lifetime"
)

type PodDeleterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is the core logic for the operator
func (r *PodDeleterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the pod object
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the pod has the `lifetimeLabel`
	lifetimeStr, exists := pod.Labels[lifetimeLabel]
	if !exists {
		logger.Info("Pod does not have lifetime label, skipping")
		return ctrl.Result{}, nil
	}

	// Parse the label value to get the lifetime in seconds
	lifetimeSecond, err := time.ParseDuration(lifetimeStr + "s")
	if err != nil {
		logger.Error(err, "Invalid lifetime label value", "value", lifetimeStr)
		return ctrl.Result{}, nil
	}

	// Find the container with the maximum start time
	var maxStartTime *metav1.Time
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Running != nil {
			if maxStartTime == nil || status.State.Running.StartedAt.After(maxStartTime.Time) {
				maxStartTime = &status.State.Running.StartedAt
			}
		}
	}

	// Skip pods without any running containers
	if maxStartTime == nil {
		logger.Info("Pod has no running containers, skipping")
		return ctrl.Result{}, nil
	}

	// Calculate the expiry time
	expiryTime := maxStartTime.Add(lifetimeSecond)
	if time.Now().After(expiryTime) {
		logger.Info("Pod lifetime expired, deleting pod", "pod", pod.Name)
		if err := r.Delete(ctx, &pod); err != nil {
			logger.Error(err, "Failed to delete pod", "pod", pod.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PodDeleterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	go func() {
		if err := r.recheckExpiredPods(context.Background()); err != nil {
			log.FromContext(context.Background()).Error(err, "Failed to run periodic expired pod recheck")
		}
	}()

    return ctrl.NewControllerManagedBy(mgr).
        For(&corev1.Pod{}).
        WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
        Complete(r)
}

// recheckExpiredPods periodically checks for expired pods and triggers deletions
func (r *PodDeleterReconciler) recheckExpiredPods(ctx context.Context) error {
    logger := log.FromContext(ctx)
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            logger.Info("Stopping periodic expired pod recheck")
            return nil
        case <-ticker.C:
            logger.Info("Periodic expired pod recheck started")

            var podList corev1.PodList
            if err := r.List(ctx, &podList); err != nil {
                logger.Error(err, "Failed to list pods")
                continue
            }
            for _, pod := range podList.Items {
                req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&pod)}
                if _, err := r.Reconcile(ctx, req); err != nil {
                    logger.Error(err, "Failed to reconcile pod", "pod", pod.Name)
                }
            }
        }
    }
}

