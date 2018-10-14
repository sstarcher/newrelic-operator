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
	Spec              LabelSpec   `json:"spec"`
	Status            LabelStatus `json:"status,omitempty"`
}

type LabelSpec struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type LabelStatus struct {
	Status string
}
