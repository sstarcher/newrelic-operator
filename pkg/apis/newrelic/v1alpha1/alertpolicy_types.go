package v1alpha1

import (
	"context"
	"errors"

	"github.com/newrelic/newrelic-client-go/pkg/alerts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlertPolicySpec defines the desired state of AlertPolicy
type AlertPolicySpec struct {
	IncidentPreference string   `json:"incident_preference,omitempty"`
	Channels           []string `json:"channels,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertPolicy is the Schema for the alertpolicies API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=alertpolicies,scope=Namespaced
type AlertPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertPolicySpec `json:"spec"`
	Status            Status          `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertPolicyList contains a list of AlertPolicy
type AlertPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlertPolicy{}, &AlertPolicyList{})
}

// Additional Code

var _ CRD = &AlertPolicy{}

// IsCreated specifies if the object has been created in new relic yet
func (s *AlertPolicy) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *AlertPolicy) toNewRelic() (*alerts.Policy, error) {
	data := alerts.Policy{
		Name:               s.GetObjectMeta().GetName(),
		IncidentPreference: alerts.IncidentPreferenceType(s.Spec.IncidentPreference),
	}

	if s.Status.ID != nil {
		data.ID = int(*s.Status.GetID())
	}
	return &data, nil
}

// Create in newrelic
func (s *AlertPolicy) Create(ctx context.Context) bool {
	input, err := s.toNewRelic()
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	data, err := client.Alerts.CreatePolicy(*input)
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	s.Status.SetID(data.ID)
	s.SetFinalizers([]string{finalizer})

	err = s.addChannels(ctx)
	if s.Status.HandleOnError(ctx, err) {
		return true
	}
	return false
}

// Delete in newrelic
func (s *AlertPolicy) Delete(ctx context.Context) bool {
	logger := GetLogger(ctx)

	id := s.Status.GetID()
	if id == nil {
		logger.Info("object does not exist")
		return false
	}

	_, err := client.Alerts.DeletePolicy(int(*id))
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	return false
}

// Update object in newrelic
func (s *AlertPolicy) Update(ctx context.Context) bool {
	input, err := s.toNewRelic()
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	_, err = client.Alerts.UpdatePolicy(*input)
	if s.Status.HandleOnError(ctx, err) {
		if err.Error() == "resource not found" {
			s.Status.ID = nil
		}
		return true
	}

	err = s.addChannels(ctx)
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	return false
}

func (s *AlertPolicy) addChannels(ctx context.Context) error {
	logger := GetLogger(ctx)

	if s.Spec.Channels != nil {
		channels, err := client.Alerts.ListChannels()
		if err != nil {
			return err
		}

		channelIds := []int{}
		for _, channel := range s.Spec.Channels {
			found := false
			for _, item := range channels {
				if channel == item.Name {
					channelIds = append(channelIds, item.ID)
					found = true
					break
				}
			}
			if !found {
				logger.Info("unable to find channel", "channel", channel)
			}
		}

		id := s.Status.GetID()
		if id == nil {
			return errors.New("id is nil")
		}

		_, err = client.Alerts.UpdatePolicyChannels(int(*id), channelIds)
		if err != nil {
			return err
		}
	}
	return nil
}
