package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlertChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertChannel `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlertChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertChannelSpec   `json:"spec"`
	Status            AlertChannelStatus `json:"status,omitempty"`
}

type AlertChannelSpec struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	// Type          AlertsChannelType   `json:"type,omitempty"`
	// Configuration interface{} `json:"configuration,omitempty"`
	// Links         *AlertsChannelLinks `json:"links,omitempty"`
}

type AlertChannelStatus struct {
	Status string
}
