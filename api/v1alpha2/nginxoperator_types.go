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

package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ReasonCRNotAvailable          = "OperatorResourceNotAvailable"
	ReasonDeploymentNotAvailable  = "OperandDeploymentNotAvailable"
	ReasonServiceNotAvailable     = "OperandServiceNotAvailable"
	ReasonOperandDeploymentFailed = "OperandDeploymentFailed"
	ReasonOperandServiceFailed    = "OperandServiceFailed"
	ReasonSucceeded               = "OperatorSucceeded"
)

// NginxOperatorSpec defines the desired state of NginxOperator
type NginxOperatorSpec struct {
	// Generic
	// Selectors adds selectors to service and deployment
	Selectors map[string]string `json:"selectors,omitempty"`
	// ForceRedeploy is any string, modifying this field instructs
	// the Operator to redeploy the Operand
	ForceRedeploy string `json:"forceRedeploy,omitempty"`

	// Service configurations
	// ServicePorts defines the ServicePorts exposed on the Nginx service
	ServicePorts []v1.ServicePort `json:"servicePorts,omitempty"`
	// ServiceType defines the type of Service used by the Nginx service
	ServiceType v1.ServiceType `json:"serviceType,omitempty"`

	// Deployment configurations
	// VolumeMounts defines the VolumeMounts on the Nginx Pod
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volumes defines the Volumes on the Nginx Pod
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// Ports defines the ContainerPorts exposed on the Nginx Pod
	Ports []v1.ContainerPort `json:"ports,omitempty"`
	// Replicas is the number of deployment replicas to scale
	Replicas *int32 `json:"replicas,omitempty"`
}

// NginxOperatorStatus defines the observed state of NginxOperator
type NginxOperatorStatus struct {
	// Conditions is the list of the most recent status condition updates
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// NginxOperator is the Schema for the nginxoperators API
type NginxOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxOperatorSpec   `json:"spec,omitempty"`
	Status NginxOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NginxOperatorList contains a list of NginxOperator
type NginxOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NginxOperator{}, &NginxOperatorList{})
}
