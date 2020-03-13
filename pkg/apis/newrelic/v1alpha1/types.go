package v1alpha1

import (
	"context"
	"crypto/sha256"
	"reflect"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
var finalizer = "needs-cleanup.newrelic.shanestarcher.com"

type Data struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec   `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

type CRD interface {
	HasChanged() bool
	Create(context.Context) bool
	Update(context.Context) bool
	Delete(context.Context) bool
	GetID() string
	IsCreated() bool
	GetDeletionTimestamp() *metav1.Time
	SetFinalizers([]string)
}

type SpecInterface interface {
	GetSum() []byte
}

type StatusInterface interface {
	GetSum() []byte
	SetSum([]byte)
	Handle(context.Context, error, string) bool
}

type Spec struct {
	Data string `json:"data,omitempty"`
}

func (s *Spec) GetSum() []byte {
	return sha256.New().Sum([]byte(s.Data))
}

type Status struct {
	ID   *string `json:"id,omitempty"`
	Info string  `json:"info,omitempty"`
	Hash []byte  `json:"hash,omitempty"`
}

// IsCreated let us know if the dashboard exists
func (s Status) IsCreated() bool {
	return s.ID != nil
}

func (s Status) GetSum() []byte {
	return s.Hash
}

func (s Status) GetID() *int64 {
	if s.ID == nil {
		return nil
	}
	id, _ := strconv.ParseInt(*s.ID, 10, 64)
	return &id
}

func (s Status) SetID(id int64) {
	str := strconv.FormatInt(id, 10)
	s.ID = &str
}

// Handle returns true if we should requeue
func (s Status) Handle(ctx context.Context, err error, msg string) bool {
	logger := GetLogger(ctx)

	if err != nil {
		s.Info = msg
		if err != nil {
			s.Info = s.Info + " " + err.Error()
		}
		logger.Info(s.Info)
		logger.Info("we set stuff")
		return true
	}
	return false
}

func (s Status) SetSum(data []byte) {
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
func created(id string, status *Status, spec SpecInterface) {
	status.ID = &id
	status.Info = "Created"
	update(spec, status)
}

func createdInt(id int64, status *Status, spec SpecInterface) {
	created(strconv.FormatInt(id, 10), status, spec)
}

func sum(data []byte) []byte {
	return sha256.New().Sum(data)
}
