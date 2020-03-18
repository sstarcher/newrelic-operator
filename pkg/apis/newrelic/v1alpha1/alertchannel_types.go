package v1alpha1

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/newrelic/newrelic-client-go/pkg/alerts"
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

// IsCreated specifies if the object has been created in new relic yet
func (s *AlertChannel) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *AlertChannel) toNewRelic() (*alerts.Channel, error) {
	data := alerts.Channel{
		Name: s.GetObjectMeta().GetName(),
		Type: alerts.ChannelType(s.Spec.Type),
	}

	bytes, err := json.Marshal(s.Spec.Configuration)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &data.Configuration)
	if err != nil {
		return nil, err
	}

	if s.Status.ID != nil {
		data.ID = int(*s.Status.GetID())
	}

	switch data.Type {
	// TODO more validation
	case "":
		return nil, errors.New("no valid type specified")
	case alerts.ChannelTypes.Slack:
		if _, ok := s.Spec.Configuration["channel"]; !ok {
			return nil, errors.New("slack notifications require channel configuration")
		}
		if _, ok := s.Spec.Configuration["url"]; !ok {
			return nil, errors.New("slack notifications require url configuration")
		}
	}

	return &data, nil
}

// Create in newrelic
func (s *AlertChannel) Create(ctx context.Context) bool {
	input, err := s.toNewRelic()
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	data, err := client.Alerts.CreateChannel(*input)
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	s.Status.Info = "Created"
	s.Status.SetID(data.ID)
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

	_, err := client.Alerts.DeleteChannel(int(*id))
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	return false
}

// Update object in newrelic
func (s *AlertChannel) Update(ctx context.Context) bool {
	// API does not list this as being updatable
	return false
}
