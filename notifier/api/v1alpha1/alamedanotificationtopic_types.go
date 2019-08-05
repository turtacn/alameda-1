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
type alamedaSubject struct {
	Kind       string `json:"kind,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type alamedaTopic struct {
	Type    []string          `json:"type,omitempty"`
	Subject []*alamedaSubject `json:"subject,omitempty"`
	Level   []string          `json:"level,omitempty"`
	Source  []*alamedaSource  `json:"source,omitempty"`
}

type alamedaSource struct {
	Host      string `json:"host,omitempty"`
	Component string `json:"component,omitempty"`
}

type AlamedaChannel struct {
	Emails []*AlamedaEmailChannel `json:"emails,omitempty"`
}

type AlamedaEmailChannel struct {
	Name string   `json:"name,"`
	To   []string `json:"to,"`
	Cc   []string `json:"cc,omitempty"`
}

// AlamedaNotificationTopicSpec defines the desired state of AlamedaNotificationTopic
type AlamedaNotificationTopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Disabled bool            `json:"disabled,omitempty"`
	Topics   []*alamedaTopic `json:"topics,"`
	Channel  *AlamedaChannel `json:"channel,"`
}

// AlamedaNotificationTopicStatus defines the observed state of AlamedaNotificationTopic
type AlamedaNotificationTopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=alamedanotificationtopics,scope=Cluster
// AlamedaNotificationTopic is the Schema for the alamedanotificationtopics API
type AlamedaNotificationTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlamedaNotificationTopicSpec   `json:"spec,omitempty"`
	Status AlamedaNotificationTopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlamedaNotificationTopicList contains a list of AlamedaNotificationTopic
type AlamedaNotificationTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlamedaNotificationTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlamedaNotificationTopic{}, &AlamedaNotificationTopicList{})
}
