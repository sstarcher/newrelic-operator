package v1alpha1

import (
	"context"
	"fmt"
	"strconv"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
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
	Type          string              `json:"type,omitempty"`
	Configuration data                `json:"configuration,omitempty"`
	Links         *AlertsChannelLinks `json:"links,omitempty"`
}

func (s AlertChannelSpec) GetSum() []byte {
	// b, _ := s.Configuration.MarshalJSON()
	// return sum(b)
	b := []byte{}
	return b
}

// AlertsChannelLinks holds the links to AlertsPolicies
type AlertsChannelLinks struct {
	PolicyIDs []*int64 `json:"policy_ids,omitempty"`
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
			Type:          newrelic.ChannelByEmail,
			Configuration: s.Spec.Configuration,
		},
	}

	if s.Status.ID != nil {
		data.AlertsChannel.ID = s.Status.ID
	}

	if s.Spec.Links != nil {
		data.AlertsChannel.Links = &newrelic.AlertsChannelLinks{
			s.Spec.Links.PolicyIDs,
		}
	}

	channels, rsp, err := client.AlertsChannels.Create(ctx, data)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		sdk.Update(s)
		return err
	}

	created(*channels.AlertsChannels[0].ID, &s.Status, &s.Spec)
	sdk.Update(s)
	return nil
}

// Delete in newrelic
func (s *AlertChannel) Delete(ctx context.Context) error {
	if s.Status.ID == nil {
		return fmt.Errorf("alert channel object has not been created %s", s.ObjectMeta.Name)
	}
	rsp, err := client.AlertsChannels.DeleteByID(ctx, *s.Status.ID)
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetID for the new relic object
func (s *AlertChannel) GetID() string {
	if s.Status.ID != nil {
		return strconv.FormatInt(*s.Status.ID, 10)
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
