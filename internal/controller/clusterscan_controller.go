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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	scanv1 "github.com/ahmedoksh/Cluster-Scan/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=scan.my.domain,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scan.my.domain,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scan.my.domain,resources=clusterscans/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

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
	log := log.FromContext(ctx)

	var clusterScan scanv1.ClusterScan
	if err := r.Get(ctx, req.NamespacedName, &clusterScan); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if clusterScan.Spec.Schedule != nil {
		// delete if Job (not Cron Job) exists
		if err := r.deleteJob(ctx, clusterScan); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.createCronJob(ctx, clusterScan); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// delete if Cron Job exists
		if err := r.deleteCronJob(ctx, clusterScan); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.createJob(ctx, clusterScan); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.updateClusterScanStatus(ctx, &clusterScan); err != nil {
		if errors.IsConflict(err) {
			log.Info("Conflict while updating ClusterScan status")
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to update ClusterScan status")
		return ctrl.Result{}, err
	} else {
		log.Info("Updated ClusterScan status")
	}

	return ctrl.Result{}, nil
}

// createJob creates a Job based on the ClusterScan specification
func (r *ClusterScanReconciler) createJob(ctx context.Context, clusterScan scanv1.ClusterScan) error {
	log := log.FromContext(ctx)

	updatedJobTemplate := clusterScan.Spec.JobTemplate.DeepCopy()
	updatedJobTemplate.ObjectMeta.Name = fmt.Sprintf("%s-job", clusterScan.Name)
	updatedJobTemplate.ObjectMeta.Namespace = clusterScan.Namespace
	if updatedJobTemplate.Spec.Template.ObjectMeta.Labels != nil {
		updatedJobTemplate.Spec.Template.ObjectMeta.Labels["ClusterScanUID"] = string(clusterScan.UID)
	} else {
		updatedJobTemplate.Spec.Template.ObjectMeta.Labels = map[string]string{"ClusterScanUID": string(clusterScan.UID)}
	}

	job := &batchv1.Job{
		ObjectMeta: *updatedJobTemplate.ObjectMeta.DeepCopy(),
		Spec:       *updatedJobTemplate.Spec.DeepCopy(),
	}

	// Set ClusterScan instance as the owner and controller
	if err := controllerutil.SetControllerReference(&clusterScan, job, r.Scheme); err != nil {
		return err
	}

	// Check if the Job already exists
	foundJob := &batchv1.Job{}
	err := r.Get(ctx, client.ObjectKey{Name: job.Name, Namespace: job.Namespace}, foundJob)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return err
		} else {
			log.Info("Created a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		}
	} else if err != nil {
		log.Error(err, "Failed to get Job")
		return err
	}

	return nil
}

// createCronJob creates a CronJob based on the ClusterScan specification
func (r *ClusterScanReconciler) createCronJob(ctx context.Context, clusterScan scanv1.ClusterScan) error {
	log := log.FromContext(ctx)

	updatedJobTemplate := clusterScan.Spec.JobTemplate
	updatedJobTemplate.ObjectMeta.Name = fmt.Sprintf("%s-scheduled-job", clusterScan.Name)
	updatedJobTemplate.ObjectMeta.Namespace = clusterScan.Namespace
	if updatedJobTemplate.Spec.Template.ObjectMeta.Labels != nil {
		updatedJobTemplate.Spec.Template.ObjectMeta.Labels["ClusterScanUID"] = string(clusterScan.UID)
	} else {
		updatedJobTemplate.Spec.Template.ObjectMeta.Labels = map[string]string{"ClusterScanUID": string(clusterScan.UID)}
	}

	// Define a cronjob
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-cronjob", clusterScan.Name),
			Namespace: clusterScan.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:    *clusterScan.Spec.Schedule,
			JobTemplate: *updatedJobTemplate.DeepCopy(),
		},
	}

	// Set ClusterScan instance as the owner and controller
	if err := controllerutil.SetControllerReference(&clusterScan, cronJob, r.Scheme); err != nil {
		return err
	}

	// Check if the CronJob already exists
	foundCronJob := &batchv1.CronJob{}
	err := r.Get(ctx, client.ObjectKey{Name: cronJob.Name, Namespace: cronJob.Namespace}, foundCronJob)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
		err = r.Create(ctx, cronJob)
		if err != nil {
			log.Error(err, "Failed to create new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
			return err
		} else {
			log.Info("Created a new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
		}
	} else if err != nil {
		log.Error(err, "Failed to get CronJob")
		return err
	}

	return nil
}

// This functions deletes the CronJob based on the ClusterScan specification if it is present
func (r *ClusterScanReconciler) deleteCronJob(ctx context.Context, clusterScan scanv1.ClusterScan) error {
	log := log.FromContext(ctx)

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-cronjob", clusterScan.Name),
			Namespace: clusterScan.Namespace,
		},
	}
	err := r.Get(ctx, client.ObjectKey{Name: cronJob.Name, Namespace: cronJob.Namespace}, cronJob)
	if err == nil {
		if err := r.Delete(ctx, cronJob); err != nil {
			log.Error(err, "Failed to delete CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
			return err
		}
		log.Info("Deleted CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
	} else if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get CronJob")
		return err
	}
	return nil
}

// This functions deletes the Job based on the ClusterScan specification if it is present
func (r *ClusterScanReconciler) deleteJob(ctx context.Context, clusterScan scanv1.ClusterScan) error {
	log := log.FromContext(ctx)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-job", clusterScan.Name),
			Namespace: clusterScan.Namespace,
		},
	}

	err := r.Get(ctx, client.ObjectKey{Name: job.Name, Namespace: job.Namespace}, job)
	if err == nil {
		if err := r.Delete(ctx, job); err != nil {
			log.Error(err, "Failed to delete job", "job.Namespace", job.Namespace, "job.Name", job.Name)
			return err
		}
		log.Info("Deleted job", "job.Namespace", job.Namespace, "job.Name", job.Name)
	} else if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get CronJob")
		return err
	}
	return nil
}

func (r *ClusterScanReconciler) updateClusterScanStatus(ctx context.Context, clusterScan *scanv1.ClusterScan) error {
	log := log.FromContext(ctx)
	log = log.WithValues("ClusterScan", clusterScan.Name)

	labelSelector, _ := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"ClusterScanUID": string(clusterScan.UID)},
	})
	childJobs := &batchv1.JobList{}
	if err := r.List(ctx, childJobs, &client.ListOptions{Namespace: clusterScan.Namespace, LabelSelector: labelSelector}); err != nil {
		log.Error(err, "Failed to list jobs")
		return err
	}
	if len(childJobs.Items) == 0 {
		log.Info("No jobs found")
		return nil
	}

	var successfulJobsCount, failedJobsCount, activeJobsCount int
	var ExecutedJobsResults []scanv1.JobInfo

	for _, job := range childJobs.Items {
		jobInfo := scanv1.JobInfo{
			Name:         job.Name,
			CreationTime: job.CreationTimestamp.Time.String(),
			CompletedAt:  "",
			Duration:     "",
		}

		if job.Status.CompletionTime != nil {
			jobInfo.CompletedAt = job.Status.CompletionTime.Time.String()
			jobInfo.Duration = job.Status.CompletionTime.Time.Sub(job.CreationTimestamp.Time).String()
		}

		if job.Status.Succeeded > 0 {
			successfulJobsCount++
			jobInfo.State = "Successful"
		} else if job.Status.Failed > 0 {
			failedJobsCount++
			jobInfo.State = "Failed"
		} else {
			activeJobsCount++
			jobInfo.State = "Active"
		}
		ExecutedJobsResults = append(ExecutedJobsResults, jobInfo)
	}

	clusterScan.Status = scanv1.ClusterScanStatus{
		SuccessfulJobsCount: successfulJobsCount,
		FailedJobsCount:     failedJobsCount,
		ActiveJobsCount:     activeJobsCount,
		TotalJobsCount:      len(childJobs.Items),
		ExecutedJobsResults: ExecutedJobsResults,
	}

	return r.Status().Update(ctx, clusterScan)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scanv1.ClusterScan{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
