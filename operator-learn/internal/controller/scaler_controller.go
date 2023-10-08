/*
Copyright 2023.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/example/memcached-operator/api/v1alpha1"
	v1App "k8s.io/api/apps/v1"
)

// ScalerReconciler reconciles a Scaler object
type ScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.banny.com,resources=scalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.banny.com,resources=scalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.banny.com,resources=scalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	// log req.NameSpace and req.Name to see the name and namespace of the scaler object

	scaler := &apiv1alpha1.Scaler{}
	err := r.Get(ctx, req.NamespacedName, scaler)
	if err != nil {
		return ctrl.Result{}, err
	}

	startTime := scaler.Spec.Start
	endTime := scaler.Spec.End
	replicas := scaler.Spec.Replicas

	// log.Log.Info(
	// 	fmt.Sprintf("Current hour: %d", time.Now().UTC().Hour()),
	// 	fmt.Sprintf("Start time: %d", startTime),
	// 	fmt.Sprintf("End time: %d", endTime),
	// 	fmt.Sprintf("Replicas: %d", replicas),
	// )

	currentHour := time.Now().UTC().Hour() // get current hour in UTC

	if currentHour >= startTime && currentHour < endTime {
		for _, deployment := range scaler.Spec.Deployments {
			// get deployment
			deploy := &v1App.Deployment{}
			err := r.Get(ctx, client.ObjectKey{
				Namespace: deployment.NameSpace,
				Name:      deployment.Name,
			}, deploy)
			if err != nil {
				return ctrl.Result{}, err
			}

			// update deployment
			if *deploy.Spec.Replicas != replicas {
				deploy.Spec.Replicas = &replicas
				err = r.Update(ctx, deploy)
				if err != nil {
					return ctrl.Result{}, err
				}

			}
			// deploy.Spec.Replicas = &replicas
			// err = r.Update(ctx, deploy)
			// if err != nil {
			// 	return ctrl.Result{}, err
			// }
		}
	}
	// else {
	// 	for _, deployment := range scaler.Spec.Deployments {
	// 		// get deployment
	// 		deploy := &v1App.Deployment{}
	// 		err := r.Get(ctx, client.ObjectKey{
	// 			Namespace: deployment.NameSpace,
	// 			Name:      deployment.Name,
	// 		}, deploy)
	// 		if err != nil {
	// 			return ctrl.Result{}, err
	// 		}

	// 		// update deployment
	// 		if *deploy.Spec.Replicas != 0 {
	// 			deploy.Spec.Replicas = &replicas
	// 			err = r.Update(ctx, deploy)
	// 			if err != nil {
	// 				return ctrl.Result{}, err
	// 			}
	// 		}
	// 		// deploy.Spec.Replicas = &replicas
	// 		// err = r.Update(ctx, deploy)
	// 		// if err != nil {
	// 		// 	return ctrl.Result{}, err
	// 		// }
	// 	}
	// }

	return ctrl.Result{
		RequeueAfter: time.Duration(30 * time.Second),
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Scaler{}).
		Complete(r)
}
