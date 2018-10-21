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
	Spec              Spec   `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

type MonitorSpec struct {
	ID           *string        `json:"id,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Type         *string        `json:"type,omitempty"`
	Frequency    *int64         `json:"frequency,omitempty"`
	URI          *string        `json:"uri,omitempty"`
	Locations    []*string      `json:"locations,omitempty"`
	Status       *string        `json:"status,omitempty"`
	SLAThreshold *float64       `json:"slaThreshold,omitempty"`
	UserID       *int64         `json:"userId,omitempty"`
	APIVersion   *string        `json:"apiVersion,omitempty"`
	CreatedAt    *string        `json:"createdAt,omitempty"`
	UpdatedAt    *string        `json:"modifiedAt,omitempty"`
	Options      MonitorOptions `json:"options,omitempty"`
	Script       *Script        `json:"script,omitempty"`
	Labels       []*string      `json:"labels,omitempty"`
}

type MonitorOptions struct {
	ValidationString       *string `json:"validationString,omitempty"`
	VerifySSL              bool    `json:"verifySSL,omitempty"`
	BypassHEADRequest      bool    `json:"bypassHEADRequest,omitempty"`
	TreatRedirectAsFailure bool    `json:"treatRedirectAsFailure,omitempty"`
}

type Script struct {
	ScriptText *string `json:"scriptText,omitempty"`
}
