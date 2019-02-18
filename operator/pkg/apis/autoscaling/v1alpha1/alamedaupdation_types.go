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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlamedaUpdationSpec defines the desired state of AlamedaUpdation
type AlamedaUpdationSpec struct {
	AlamedaRecommendationSpec AlamedaRecommendationSpec `json:"alamedaRecommendationSpec,omitempty" protobuf:"bytes,1,opt,name=alameda_recommendation_spec"`
}

// AlamedaUpdationStatus defines the observed state of AlamedaUpdation
type AlamedaUpdationStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlamedaUpdation is the Schema for the alamedaupdations API
// +k8s:openapi-gen=true
type AlamedaUpdation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlamedaUpdationSpec   `json:"spec,omitempty"`
	Status AlamedaUpdationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlamedaUpdationList contains a list of AlamedaUpdation
type AlamedaUpdationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlamedaUpdation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlamedaUpdation{}, &AlamedaUpdationList{})
}
