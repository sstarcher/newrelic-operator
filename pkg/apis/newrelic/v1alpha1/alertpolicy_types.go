package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/apex/log"
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

func (s AlertPolicySpec) GetSum() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		log.Error("unable to marshal AlertPolicy")
		return nil
	}
	return sum(b)
}

// IsCreated specifies if the object has been created in new relic yet
func (s *AlertPolicy) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *AlertPolicy) HasChanged() bool {
	return hasChanged(&s.Spec, &s.Status)
}

// Create in newrelic
func (s *AlertPolicy) Create(ctx context.Context) bool {
	data := &newrelic.AlertsPolicyEntity{
		AlertsPolicy: &newrelic.AlertsPolicy{
			Name:               &s.ObjectMeta.Name,
			IncidentPreference: newrelic.IncidentPerPolicy,
		},
	}

	data, rsp, err := client.AlertsPolicies.Create(ctx, data)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	createdInt(*data.AlertsPolicy.ID, &s.Status, &s.Spec)
	s.SetFinalizers([]string{finalizer})

	err = s.addChannels(ctx)
	if err != nil {
		s.Status.Info = err.Error()
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

	rsp, err := client.AlertsPolicies.DeleteByID(ctx, *id)
	if rsp.StatusCode == 404 {
		logger.Info("unable to find id, skipping deletion", "id", id)
		return false
	}
	err = handleError(rsp, err)
	if err != nil {
		return true
	}

	return false
}

// GetID for the new relic object
func (s *AlertPolicy) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return "nil"
}

// Update object in newrelic
func (s *AlertPolicy) Update(ctx context.Context) bool {
	logger := GetLogger(ctx)
	// TODO update is creating extra objects
	data := &newrelic.AlertsPolicyEntity{
		AlertsPolicy: &newrelic.AlertsPolicy{
			Name:               &s.ObjectMeta.Name,
			IncidentPreference: newrelic.IncidentPreferenceOption(s.Spec.IncidentPreference),
		},
	}

	id := s.Status.GetID()
	if id == nil {
		s.Status.ID = nil
		logger.Info("object does not exist")
		return true
	}

	data, rsp, err := client.AlertsPolicies.Update(ctx, data, *id)
	if rsp.StatusCode == 404 {
		s.Status.ID = nil
		logger.Info("id is missing recreating")
		return false
	}

	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	err = s.addChannels(ctx)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	update(&s.Spec, &s.Status)
	return false
}

func (s *AlertPolicy) addChannels(ctx context.Context) error {
	logger := GetLogger(ctx)

	if s.Spec.Channels != nil {
		channels, rsp, err := client.AlertsChannels.ListAll(ctx, nil)
		err = handleError(rsp, err)
		if err != nil {
			s.Status.Info = err.Error()
			return err
		}

		channelIds := []*int64{}
		for _, channel := range s.Spec.Channels {
			found := false
			for _, alertChannel := range channels.AlertsChannels {
				if channel == *alertChannel.Name {
					channelIds = append(channelIds, alertChannel.ID)
					found = true
					break
				}
			}
			if !found {
				logger.Info("unable to find", "channel", channel)
			}
		}

		rsp, err = client.AlertsChannels.UpdatePolicyChannels(ctx, *s.Status.GetID(), channelIds)
		err = handleError(rsp, err)
		if err != nil {
			s.Status.Info = err.Error()
			return err
		}
	}
	return nil
}
