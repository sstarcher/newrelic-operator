package v1alpha1

import (
	"context"
	"fmt"
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
	Create(context.Context) bool
	Update(context.Context) bool
	Delete(context.Context) bool
	IsCreated() bool
	GetDeletionTimestamp() *metav1.Time
	SetFinalizers([]string)
}

type SpecInterface interface {
}

type StatusInterface interface {
	HandleOnErrorMessage(context.Context, error, string) bool
	HandleOnError(context.Context, error) bool
}

type Spec struct {
	Data string `json:"data,omitempty"`
}

type Status struct {
	ID   *string `json:"id,omitempty"`
	Info string  `json:"info,omitempty"`
	Hash []byte  `json:"hash,omitempty"`
}

// IsCreated let us know if the dashboard exists
func (s *Status) IsCreated() bool {
	return s.ID != nil
}

func (s *Status) GetID() *int {
	if s.ID == nil {
		return nil
	}
	id, _ := strconv.Atoi(*s.ID)
	return &id
}

func (s *Status) SetID(id int) {
	str := string(id)
	s.ID = &str
}

// HandleOnErrorMessage returns true if an error had occured
func (s *Status) HandleOnErrorMessage(ctx context.Context, err error, msg string) bool {
	if msg != "" {
		err = fmt.Errorf("%s %v", msg, err)
	}

	return s.HandleOnError(ctx, err)
}

// HandleOnError returns true if an error had occured
func (s *Status) HandleOnError(ctx context.Context, err error) bool {
	logger := GetLogger(ctx)

	if err != nil {
		s.Info = err.Error()
		logger.Info(s.Info)
		return true
	}
	return false
}
