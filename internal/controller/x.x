/*
Copyright 2024.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	scanv1 "github.com/ahmedoksh/Cluster-Scan/api/v1"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=scan.my.domain,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scan.my.domain,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scan.my.domain,resources=clusterscans/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterScan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// c.Watch(&source.Kind{Type: &scanv1.ClusterScan{}}, &handler.EnqueueRequestForObject{})

	// ctrl.Builder.Watches("ClusterScan", &scanv1.ClusterScan{}, &ctrl.WatchOptions{Namespace: "default"}) // TODO(user): Modify as needed
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scanv1.ClusterScan{}).
		Complete(r)
}

// func (r *ReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
// ctrl.Builder.Watches()
// return ctrl.NewControllerManagedBy(mgr).
// 	For(&appsv1.ReplicaSet{}).
// 	Watches(
// 		&source.Kind{Type: &corev1.Pod{}},
// 		&handler.EnqueueRequestForOwner{
// 			IsController: true,
// 			OwnerType:    &appsv1.ReplicaSet{},
// 		},
// 	).
// 	Complete(r)
// }
