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

package v1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterScanSpec defines the desired state of ClusterScan
type ClusterScanSpec struct {
	// +kubebuilder:validation:Optional
	// Schedule if ClusterScan is doing CronJob
	Schedule *string `json:"schedule,omitempty"`
	// +kubebuilder:validation:Required
	// Template for the job that will be created
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate"`
}

// ClusterScanStatus defines the observed state of ClusterScan
type JobInfo struct {
	// Name of the job
	Name string `json:"name"`
	// Time when the job was created
	CreationTime string `json:"creationTime"`
	// Time when the job was completed (if it is completed)
	CompletedAt string `json:"completedAt"`
	// Duration of the job (if it is completed)
	Duration string `json:"duration"`
	// State of the job: "Succeeded", "Failed", "Running"
	State string `json:"state"`
}

type ClusterScanStatus struct {
	// Total number of jobs that have been executed
	TotalJobsCount int `json:"totalJobsCount"`
	// Number of jobs that have been executed successfully
	SuccessfulJobsCount int `json:"successfulJobsCount"`
	// Number of jobs that have failed
	FailedJobsCount int `json:"failedJobsCount"`
	// Number of jobs that are currently running
	ActiveJobsCount int `json:"activeJobsCount"`
	// List of executed jobs with their results
	ExecutedJobsResults []JobInfo `json:"executedJobsResults"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterScan is the Schema for the clusterscans API
type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterScanList contains a list of ClusterScan
type ClusterScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterScan{}, &ClusterScanList{})
}
