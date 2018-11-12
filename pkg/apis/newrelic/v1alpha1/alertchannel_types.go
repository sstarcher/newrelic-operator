package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ CRD = &AlertChannel{}

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
	Spec              AlertChannelSpec `json:"spec"`
	Status            Status           `json:"status,omitempty"`
}

type data map[string]string

type AlertChannelSpec struct {
	Type          string   `json:"type,omitempty"`
	Configuration data     `json:"configuration,omitempty"`
	Policies      []string `json:"policies,omitempty"`
}

func (s AlertChannelSpec) GetSum() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		log.Error("unable to marshal Alert Channel")
		return nil
	}
	return sum(b)
}

// IsCreated specifies if the object has been created in new relic yet
func (s *AlertChannel) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *AlertChannel) HasChanged() bool {
	return false
}

// Create in newrelic
func (s *AlertChannel) Create(ctx context.Context) error {
	data := &newrelic.AlertsChannelEntity{
		AlertsChannel: &newrelic.AlertsChannel{
			Name:          &s.ObjectMeta.Name,
			Type:          newrelic.AlertsChannelType(s.Spec.Type),
			Configuration: s.Spec.Configuration,
		},
	}

	channels, rsp, err := client.AlertsChannels.Create(ctx, data)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return err
	}

	createdInt(*channels.AlertsChannels[0].ID, &s.Status, &s.Spec)
	return nil
}

// Delete in newrelic
func (s *AlertChannel) Delete(ctx context.Context) error {
	id := s.Status.GetID()
	if id == nil {
		return fmt.Errorf("alert channel object has not been created %s", s.ObjectMeta.Name)
	}
	rsp, err := client.AlertsChannels.DeleteByID(ctx, *id)
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetID for the new relic object
func (s *AlertChannel) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return "nil"
}

// Signature for the CRD
func (s *AlertChannel) Signature() string {
	return fmt.Sprintf("%s %s/%s", s.TypeMeta.Kind, s.Namespace, s.Name)
}

// Update object in newrelic
func (s *AlertChannel) Update(ctx context.Context) error {
	// API does not list this as being updatable
	return nil
}

func init() {
	SchemeBuilder.Register(&AlertChannel{}, &AlertChannelList{})
}
