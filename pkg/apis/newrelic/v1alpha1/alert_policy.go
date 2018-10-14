package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlertPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertPolicy `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlertPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertPolicySpec   `json:"spec"`
	Status            AlertPolicyStatus `json:"status,omitempty"`
}

type AlertPolicySpec struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type AlertPolicyStatus struct {
	Status string
}
