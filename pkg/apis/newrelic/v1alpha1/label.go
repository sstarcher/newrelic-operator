package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LabelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Label `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Label struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              LabelSpec `json:"spec"`
	Status            Status    `json:"status,omitempty"`
}

type LabelSpec struct {
	Key                           *string             `json:"key,omitempty"`
	Category                      *string             `json:"category,omitempty"`
	Name                          *string             `json:"name,omitempty"`
	LabelsApplicationHealthStatus *LabelsHealthStatus `json:"application_health_status,omitempty"`
	LabelsServerHealthStatus      *LabelsHealthStatus `json:"server_health_status,omitempty"`
	LabelLinks                    *LabelLinks         `json:"links,omitempty"`
}

type LabelsHealthStatus struct {
	Green  []*int64 `json:"green,omitempty"`
	Orange []*int64 `json:"orange,omitempty"`
	Red    []*int64 `json:"red,omitempty"`
	Gray   []*int64 `json:"gray,omitempty"`
}

type LabelLinks struct {
	Applications []*int64 `json:"applications,omitempty"`
	Servers      []*int64 `json:"servers,omitempty"`
}
