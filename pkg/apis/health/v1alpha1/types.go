package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HealthCheck defines the healthcheck resource.
type HealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HealthCheckSpec   `json:"spec"`
	Status HealthCheckStatus `json:"status"`
}

// HealthCheckSpec defines the specification of a HealthCheck resource.
type HealthCheckSpec struct {
	Image string `json:"image"`
}

// HealthCheckStatus defines the status object of a HealthCheck resource.
type HealthCheckStatus struct {
	Healthy bool `json:"healthy"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HealthCheckList is a list of HealthCheck resources.
type HealthCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []HealthCheck `json:"items"`
}
