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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigurationSpec defines the desired state of the Misconfiguration to be applied to deployments
type ConfigurationSpec struct {

	// Set Container Imagetag
	ImageTag string `json:"imageTag,omitempty"`

	// Set ContainerPort
	ContainerPort int32 `json:"containerPort,omitempty"`

	// Set allowPrivilegeEscalation
	AllowPrivilegeEscalation bool `json:"allowPrivilegeEscalation,omitempty"`

	// Set readOnlyRootFilesystem
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`

	// Set runAsNonRoot
	RunAsNonRoot bool `json:"runAsNonRoot,omitempty"`

	// CPU limits
	CPULimits resource.Quantity `json:"limits,omitempty"`

	// Memory limits
	MemoryLimits resource.Quantity `json:"memorylimits,omitempty"`

	//CPU requests
	CPURequests resource.Quantity `json:"requests,omitempty"`

	// Memory requests
	MemoryRequests resource.Quantity `json:"memoryrequests,omitempty"`
}

// ConfigurationStatus defines the observed state of Configuration
type ConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Configuration is the Schema for the configurations API
type Configuration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigurationSpec   `json:"spec,omitempty"`
	Status ConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigurationList contains a list of Configuration
type ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Configuration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Configuration{}, &ConfigurationList{})
}
