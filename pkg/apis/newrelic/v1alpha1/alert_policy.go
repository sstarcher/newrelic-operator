package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	IncidentPreference string `json:"incident_preference,omitempty"`
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
		sdk.Update(s)
		return err
	}

	created(*data.AlertsPolicy.ID, &s.Status, &s.Spec)
	return sdk.Update(s)
}

// Delete in newrelic
func (s *AlertPolicy) Delete(ctx context.Context) error {
	if s.Status.ID == nil {
		return fmt.Errorf("alert Policy object has not been created %s", s.ObjectMeta.Name)
	}
	rsp, err := client.AlertsPolicies.DeleteByID(ctx, *s.Status.ID)
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetID for the new relic object
func (s *AlertPolicy) GetID() string {
	if s.Status.ID != nil {
		return strconv.FormatInt(*s.Status.ID, 10)
	}
	return "nil"
}

// Signature for the CRD
func (s *AlertPolicy) Signature() string {
	return fmt.Sprintf("%s %s/%s", s.TypeMeta.Kind, s.Namespace, s.Name)
}

// Update object in newrelic
func (s *AlertPolicy) Update(ctx context.Context) error {
	data := &newrelic.AlertsPolicyEntity{
		AlertsPolicy: &newrelic.AlertsPolicy{
			Name:               &s.ObjectMeta.Name,
			IncidentPreference: newrelic.IncidentPerPolicy,
		},
	}
	data, rsp, err := client.AlertsPolicies.Update(ctx, data, *s.Status.ID)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		sdk.Update(s)
		return err
	}

	update(&s.Spec, &s.Status)
	return sdk.Update(s)
}
