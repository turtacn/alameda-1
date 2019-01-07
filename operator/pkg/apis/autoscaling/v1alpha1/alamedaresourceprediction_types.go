/*
Copyright 2019 The Alameda Authors.

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

package v1alpha1

import (
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlamedaResourcePredictionSpec defines the desired state of AlamedaResourcePrediction
type AlamedaResourcePredictionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,1,opt,name=selector"`
}

// AlamedaResourcePredictionStatus defines the observed state of AlamedaResourcePrediction
type AlamedaResourcePredictionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Prediction AlamedaPrediction `json:"prediction,omitempty" protobuf:"bytes,1,opt,name=prediction,casttype=AlamedaPrediction"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlamedaResourcePrediction is the Schema for the alamedaresourcepredictions API
// +k8s:openapi-gen=true
type AlamedaResourcePrediction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlamedaResourcePredictionSpec   `json:"spec,omitempty"`
	Status AlamedaResourcePredictionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlamedaResourcePredictionList contains a list of AlamedaResourcePrediction
type AlamedaResourcePredictionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlamedaResourcePrediction `json:"items"`
}
type PodUID string
type ContainerName string
type DeploymentUID string
type ResourceType string
type Recommendation struct {
	Time      int64
	Date      string
	Resources apicorev1.ResourceRequirements
}
type PredictData struct {
	Time  int64
	Date  string
	Value string
}
type TimeSeriesData struct {
	PredictData []PredictData
}
type PredictContainer struct {
	Name            string
	RawPredict      map[ResourceType]TimeSeriesData
	Recommendations []Recommendation
	InitialResource apicorev1.ResourceRequirements
}
type PredictPod struct {
	Name       string
	Containers map[ContainerName]PredictContainer
}
type PredictDeployment struct {
	UID       string
	Namespace string
	Name      string
	Pods      map[PodUID]PredictPod
}
type AlamedaPrediction struct {
	Deployments map[DeploymentUID]PredictDeployment
}

func init() {
	SchemeBuilder.Register(&AlamedaResourcePrediction{}, &AlamedaResourcePredictionList{})
}
