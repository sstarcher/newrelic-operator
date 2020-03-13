package v1alpha1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/apex/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlertChannelSpec defines the desired state of AlertChannel
type AlertChannelSpec struct {
	// TODO don't require setting of the type
	Type          string   `json:"type,omitempty"`
	Configuration data     `json:"configuration,omitempty"`
	Policies      []string `json:"policies,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertChannel is the Schema for the alertchannels API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=alertchannels,scope=Namespaced
type AlertChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertChannelSpec `json:"spec"`
	Status            Status           `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertChannelList contains a list of AlertChannel
type AlertChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertChannel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlertChannel{}, &AlertChannelList{})
}

// Additional code

var _ CRD = &AlertChannel{}

type data map[string]string

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

func (s *AlertChannel) validate() error {
	slackType := newrelic.AlertsChannelType(s.Spec.Type)
	switch slackType {
	case newrelic.ChannelBySlack:
		if _, ok := s.Spec.Configuration["channel"]; !ok {
			return errors.New("slack notifications require channel configuration")
		}
		if _, ok := s.Spec.Configuration["url"]; !ok {
			return errors.New("slack notifications require url configuration")
		}
	}
	return nil
}

// Create in newrelic
func (s *AlertChannel) Create(ctx context.Context) bool {
	if err := s.validate(); err != nil {
		return true
	}

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
		return true
	}

	createdInt(*channels.AlertsChannels[0].ID, &s.Status, &s.Spec)
	s.SetFinalizers([]string{finalizer})
	return false
}

// Delete in newrelic
func (s *AlertChannel) Delete(ctx context.Context) bool {
	logger := GetLogger(ctx)
	id := s.Status.GetID()
	if id == nil {
		logger.Info("object does not exist")
		return false
	}
	rsp, err := client.AlertsChannels.DeleteByID(ctx, *id)
	if rsp.StatusCode == 404 {
		return false
	}
	err = handleError(rsp, err)
	if err != nil {
		return true
	}

	return false
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
func (s *AlertChannel) Update(ctx context.Context) bool {
	// API does not list this as being updatable
	return false
}
