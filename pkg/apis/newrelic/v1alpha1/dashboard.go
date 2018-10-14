package v1alpha1

import (
	"crypto/sha256"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Dashboard `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DashboardSpec   `json:"spec"`
	Status            DashboardStatus `json:"status,omitempty"`
}

type DashboardSpec struct {
	Data string
}

type DashboardStatus struct {
	ID   *int64
	Info string
	Hash []byte
}

// IsCreated let us know if the dashboard exists
func (s *DashboardStatus) IsCreated() bool {
	return s.ID != nil
}

// HasChanged detects if the data is out of sync with the hash
func (s *Dashboard) HasChanged() bool {
	hash := sha256.New().Sum([]byte(s.Spec.Data))
	return !reflect.DeepEqual(hash, s.Status.Hash)
}

// Update the hash
func (s *Dashboard) Update() {
	s.Status.Hash = sha256.New().Sum([]byte(s.Spec.Data))
}

// Update the hash
func (s *Dashboard) Created(id int64) {
	s.Status.ID = &id
	s.Status.Info = "Created"
	s.Update()
}
