package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ CRD = &Monitor{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Monitor `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              MonitorSpec `json:"spec"`
	Status            Status      `json:"status,omitempty"`
}

type MonitorSpec struct {
	Type         *string        `json:"type,omitempty"`
	Frequency    *int64         `json:"frequency,omitempty"`
	URI          *string        `json:"uri,omitempty"`
	Locations    []*string      `json:"locations,omitempty"`
	Status       *MonitorStatus `json:"status,omitempty"`
	SLAThreshold *float64       `json:"slaThreshold,omitempty"`
	Options      MonitorOptions `json:"options,omitempty"`
	Script       *Script        `json:"script,omitempty"`
	Conditions   []Conditions   `json:"conditions,omitempty"`
}

type MonitorStatus string

const (
	Enabled  MonitorStatus = "enabled"
	Disabled MonitorStatus = "disabled"
	Muted    MonitorStatus = "muted"
)

func (s MonitorStatus) String() string {
	return string(s)
}

type MonitorOptions struct {
	ValidationString       *string `json:"validationString,omitempty"`
	VerifySSL              bool    `json:"verifySSL,omitempty"`
	BypassHEADRequest      bool    `json:"bypassHEADRequest,omitempty"`
	TreatRedirectAsFailure bool    `json:"treatRedirectAsFailure,omitempty"`
}

type Conditions struct {
	PolicyName string  `json:"policyName,omitempty"`
	RunbookURL *string `json:"runbookURL,omitempty"`
}

type Script struct {
	ScriptText *string `json:"scriptText,omitempty"`
}

func (s MonitorSpec) GetSum() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		log.Error("unable to marshal Monitor")
		return nil
	}
	return sum(b)
}

// IsCreated specifies if the object has been created in new relic yet
func (s *Monitor) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *Monitor) HasChanged() bool {
	return hasChanged(&s.Spec, &s.Status)
}

func (s *Monitor) toNewRelic() *newrelic.Monitor {
	data := &newrelic.Monitor{
		Name:         &s.ObjectMeta.Name,
		Type:         s.Spec.Type,
		Frequency:    s.Spec.Frequency,
		URI:          s.Spec.URI,
		Locations:    s.Spec.Locations,
		SLAThreshold: s.Spec.SLAThreshold,
		Script:       &newrelic.Script{},
		Options: newrelic.MonitorOptions{
			ValidationString:       s.Spec.Options.ValidationString,
			VerifySSL:              s.Spec.Options.VerifySSL,
			BypassHEADRequest:      s.Spec.Options.BypassHEADRequest,
			TreatRedirectAsFailure: s.Spec.Options.TreatRedirectAsFailure,
		},
	}

	if s.Spec.Script != nil {
		data.Script.ScriptText = s.Spec.Script.ScriptText
	}

	if s.Spec.Status != nil {
		status := s.Spec.Status.String()
		data.Status = &status
	}

	if data.Type == nil {
		val := "simple"
		data.Type = &val
	}

	if data.Frequency == nil {
		val := int64(10)
		data.Frequency = &val
	}

	if data.Locations == nil {
		val := "AWS_US_WEST_1"
		data.Locations = []*string{
			&val,
		}
	}

	if data.Status == nil {
		val := Enabled.String()
		data.Status = &val
	}

	return data
}

// Create in newrelic
func (s *Monitor) Create(ctx context.Context) error {
	data, rsp, err := clientSythetics.SyntheticsMonitors.Create(ctx, s.toNewRelic())
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return err
	}

	// TODO HTTP 400 could be already exists
	// TODO when policy error happens ID/finalizer are not being set only the Status
	created(*data.ID, &s.Status, &s.Spec)
	s.SetFinalizers([]string{finalizer})

	if s.Spec.Conditions != nil {
		for _, item := range s.Spec.Conditions {
			cond := &newrelic.AlertsConditionEntity{
				AlertsSyntheticsConditionEntity: &newrelic.AlertsSyntheticsConditionEntity{
					AlertsSyntheticsCondition: &newrelic.AlertsSyntheticsCondition{
						Name:       &s.Name,
						MonitorID:  s.Status.ID,
						RunbookURL: item.RunbookURL,
					},
				},
			}

			policies, rsp, err := client.AlertsPolicies.ListAll(ctx, &newrelic.AlertsPolicyListOptions{
				NameOptions: item.PolicyName,
			})
			err = handleError(rsp, err)
			if err != nil {
				s.Status.Info = err.Error()
				return err
			}

			if len(policies.AlertsPolicies) != 1 {
				err = fmt.Errorf("expected a policy search by name to only return 1 result, but recieved %d", len(policies.AlertsPolicies))
				s.Status.Info = err.Error()
				return err
			}

			policyID := *policies.AlertsPolicies[0].ID
			_, rsp, err = client.AlertsConditions.Create(ctx, newrelic.ConditionSynthetics, cond, policyID)
			err = handleError(rsp, err)
			if err != nil {
				s.Status.Info = err.Error()
				return err
			}
		}
	}

	return nil
}

// Delete in newrelic
func (s *Monitor) Delete(ctx context.Context) error {
	id := s.Status.ID
	if id == nil {
		return fmt.Errorf("alert Policy object has not been created %s", s.ObjectMeta.Name)
	}

	// TODO Detach alert
	rsp, err := clientSythetics.SyntheticsMonitors.DeleteByID(ctx, id)
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetID for the new relic object
func (s *Monitor) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return ""
}

// Signature for the CRD
func (s *Monitor) Signature() string {
	return fmt.Sprintf("%s %s/%s", s.TypeMeta.Kind, s.Namespace, s.Name)
}

// Update object in newrelic
func (s *Monitor) Update(ctx context.Context) error {
	rsp, err := clientSythetics.SyntheticsMonitors.Update(ctx, s.toNewRelic(), s.Status.ID)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return err
	}

	update(&s.Spec, &s.Status)
	return nil
}

func init() {
	SchemeBuilder.Register(&Monitor{}, &MonitorList{})
}
