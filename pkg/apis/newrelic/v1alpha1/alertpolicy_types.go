package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ CRD = &AlertPolicy{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlertPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AlertPolicy `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AlertPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AlertPolicySpec `json:"spec"`
	Status            Status          `json:"status,omitempty"`
}

type AlertPolicySpec struct {
	IncidentPreference string   `json:"incident_preference,omitempty"`
	Channels           []string `json:"channels,omitempty"`
}

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
func (s *AlertPolicy) Create(ctx context.Context) error {
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
		return err
	}

	createdInt(*data.AlertsPolicy.ID, &s.Status, &s.Spec)
	s.SetFinalizers([]string{finalizer})

	return s.addChannels(ctx)
}

// Delete in newrelic
func (s *AlertPolicy) Delete(ctx context.Context) error {
	id := s.Status.GetID()
	if id == nil {
		return fmt.Errorf("alert Policy object has not been created %s", s.ObjectMeta.Name)
	}

	rsp, err := client.AlertsPolicies.DeleteByID(ctx, *id)
	if rsp.StatusCode == 404 {
		log.Warn(responseBodyToString(rsp))
		return nil
	}
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetID for the new relic object
func (s *AlertPolicy) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return "nil"
}

// Update object in newrelic
func (s *AlertPolicy) Update(ctx context.Context) error {
	// TODO update is creating extra objects
	data := &newrelic.AlertsPolicyEntity{
		AlertsPolicy: &newrelic.AlertsPolicy{
			Name:               &s.ObjectMeta.Name,
			IncidentPreference: newrelic.IncidentPreferenceOption(s.Spec.IncidentPreference),
		},
	}

	id := s.Status.GetID()
	if id == nil {
		return fmt.Errorf("alert Policy object has not been created %s", s.ObjectMeta.Name)
	}

	data, rsp, err := client.AlertsPolicies.Update(ctx, data, *id)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return err
	}

	err = s.addChannels(ctx)
	if err != nil {
		s.Status.Info = err.Error()
		return err
	}

	update(&s.Spec, &s.Status)
	return nil
}

func (s *AlertPolicy) addChannels(ctx context.Context) error {
	log := GetLogger(ctx)

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
				log.Warnf("unable to find the %s channel", channel)
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

func init() {
	SchemeBuilder.Register(&AlertPolicy{}, &AlertPolicyList{})
}
