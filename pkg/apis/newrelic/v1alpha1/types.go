package v1alpha1

import (
	"context"
	"crypto/sha256"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Data struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec   `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

type CRD interface {
	HasChanged() bool
	// UpdateStatus()
	// Created(int64)
	Create(context.Context) error
	Update(context.Context) error
	Delete(context.Context) error
	Signature() string
	GetID() string
	IsCreated() bool
}

type SpecInterface interface {
	GetSum() []byte
}

type StatusInterface interface {
	GetSum() []byte
	SetSum([]byte)
}

type Spec struct {
	Data string
}

func (s *Spec) GetSum() []byte {
	return sha256.New().Sum([]byte(s.Data))
}

type Status struct {
	ID   *int64 `json:"id,omitempty"`
	Info string `json:"info,omitempty"`
	Hash []byte `json:"hash,omitempty"`
}

// IsCreated let us know if the dashboard exists
func (s *Status) IsCreated() bool {
	return s.ID != nil
}

func (s *Status) GetSum() []byte {
	return s.Hash
}

func (s *Status) SetSum(data []byte) {
	s.Hash = data
}

// HasChanged detects if the data is out of sync with the hash
func hasChanged(spec SpecInterface, status SpecInterface) bool {
	return !reflect.DeepEqual(spec.GetSum(), status.GetSum())
}

// Update the hash
func update(spec SpecInterface, status StatusInterface) {
	status.SetSum(spec.GetSum())
}

// Update the hash
func created(id int64, status *Status, spec SpecInterface) {
	status.ID = &id
	status.Info = "Created"
	update(spec, status)
}

func sum(data []byte) []byte {
	return sha256.New().Sum(data)
}
