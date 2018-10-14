package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Monitor `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              MonitorSpec   `json:"spec"`
	Status            MonitorStatus `json:"status,omitempty"`
}

type MonitorSpec struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type MonitorStatus struct {
	Status string
}
