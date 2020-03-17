package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/newrelic/newrelic-client-go/pkg/alerts"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
)

// MonitorSpec defines the desired state of Monitor
type MonitorSpec struct {
	Type          *string              `json:"type,omitempty"`
	Frequency     *int64               `json:"frequency,omitempty"`
	URI           *string              `json:"uri,omitempty"`
	Locations     []*string            `json:"locations,omitempty"`
	Status        *MonitorStatusString `json:"status,omitempty"`
	SLAThreshold  *string              `json:"slaThreshold,omitempty"`
	ManageUpdates *bool                `json:"manageUpdates,omitempty"`
	Options       MonitorOptions       `json:"options,omitempty"`
	Script        *Script              `json:"script,omitempty"`
	Conditions    []Conditions         `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Monitor is the Schema for the monitors API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=monitors,scope=Namespaced
type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              MonitorSpec `json:"spec"`
	Status            Status      `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MonitorList contains a list of Monitor
type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Monitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Monitor{}, &MonitorList{})
}

// Additional Code

var _ CRD = &Monitor{}

// MonitorType defines the available types for monitors
type MonitorType string

const (
	typePing            MonitorType = "SIMPLE"
	typeBrowser                     = "BROWSER"
	typeScriptedBrowser             = "SCRIPT_BROWSER"
	typeAPI                         = "SCRIPT_API"
)

type MonitorStatusString string

const (
	Enabled  MonitorStatusString = "enabled"
	Disabled MonitorStatusString = "disabled"
	Muted    MonitorStatusString = "muted"
)

func (s MonitorStatusString) String() string {
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

// TODO flatten this structure out
type Script struct {
	ScriptText *string `json:"scriptText,omitempty"`
}

// IsCreated specifies if the object has been created in new relic yet
func (s *Monitor) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *Monitor) toNewRelic() (*synthetics.Monitor, error) {

	data := &synthetics.Monitor{
		Name: s.ObjectMeta.Name,
		Options: synthetics.MonitorOptions{
			VerifySSL:              s.Spec.Options.VerifySSL,
			BypassHEADRequest:      s.Spec.Options.BypassHEADRequest,
			TreatRedirectAsFailure: s.Spec.Options.TreatRedirectAsFailure,
		},
	}

	if s.Status.ID != nil {
		data.ID = *s.Status.ID
	}

	if s.Spec.Status != nil {
		data.Status = synthetics.MonitorStatusType(s.Spec.Status.String())
	}

	if s.Spec.Type == nil {
		data.Type = synthetics.MonitorType(typePing)
	} else {
		data.Type = synthetics.MonitorType(*s.Spec.Type)
	}

	if s.Spec.Frequency == nil {
		data.Frequency = 10
	} else {
		data.Frequency = uint(*s.Spec.Frequency)
	}

	if s.Spec.Locations == nil {
		data.Locations = []string{"AWS_US_WEST_1"}
	} else {
		data.Locations = []string{}
		for _, item := range s.Spec.Locations {
			data.Locations = append(data.Locations, *item)
		}
	}

	if s.Spec.SLAThreshold != nil {
		s, err := strconv.ParseFloat(*s.Spec.SLAThreshold, 64)
		if err != nil {
			return nil, err
		}
		data.SLAThreshold = s
	} else {
		data.SLAThreshold = 1.0
	}

	if s.Spec.Options.ValidationString != nil {
		data.Options.ValidationString = *s.Spec.Options.ValidationString
	}

	if s.Spec.URI != nil {
		data.URI = *s.Spec.URI
	}

	if s.Spec.Status == nil {
		data.Status = synthetics.MonitorStatusType(Enabled.String())
	}

	return data, nil
}

// Create in newrelic
func (s *Monitor) Create(ctx context.Context) bool {
	input, err := s.toNewRelic()
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	data, err := client.Synthetics.CreateMonitor(*input)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	s.Status.Info = "Created"
	s.Status.ID = &data.ID
	s.SetFinalizers([]string{finalizer})

	err = s.updateScript(ctx)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	err = s.updateCondition(ctx)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	return false
}

// Delete in newrelic
func (s *Monitor) Delete(ctx context.Context) bool {
	logger := GetLogger(ctx)

	if s.Status.ID == nil {
		logger.Info("object does not exist")
		return false
	}

	err := client.Synthetics.DeleteMonitor(*s.Status.ID)
	if err != nil {
		return true
	}

	return false
}

// Update object in newrelic
func (s *Monitor) Update(ctx context.Context) bool {
	monitor, err := s.toNewRelic()
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	if s.Spec.ManageUpdates != nil && *s.Spec.ManageUpdates {
		data, err := s.getCurrent(ctx)
		if err != nil {
			return true
		}
		monitor.Status = data.Status
	}

	s.Status.Info = "Updated"
	_, err = client.Synthetics.UpdateMonitor(*monitor)
	if s.Status.HandleOnErrorMessage(ctx, err, "failed") {
		return true
	}

	err = s.updateScript(ctx)
	if s.Status.HandleOnErrorMessage(ctx, err, "failed on script") {
		return true
	}

	err = s.updateCondition(ctx)
	if s.Status.HandleOnErrorMessage(ctx, err, "failed on condition") {
		return true
	}

	return false
}

func (s *Monitor) updateScript(ctx context.Context) error {
	if s.Spec.Type != nil && strings.ToUpper(*s.Spec.Type) == typeAPI && s.Spec.Script != nil && s.Spec.Script.ScriptText != nil {
		_, err := client.Synthetics.UpdateMonitorScript(*s.Status.ID, synthetics.MonitorScript{
			Text: *s.Spec.Script.ScriptText,
		})
		return err
	}
	return nil
}

func (s *Monitor) updateCondition(ctx context.Context) error {
	if s.Spec.Conditions != nil {
		for _, item := range s.Spec.Conditions {
			policyID, err := s.findPolicyID(ctx, item.PolicyName)
			if err != nil {
				return err
			}

			result, err := client.Alerts.ListSyntheticsConditions(*policyID)
			if err != nil {
				return err
			}

			exists := false
			if result != nil {
				for _, item := range result {
					id := s.Status.GetID()
					if id != nil && item.MonitorID == string(*id) {
						exists = true
					}
				}
			}

			if exists {
				continue
			}

			_, err = client.Alerts.CreateSyntheticsCondition(*policyID, alerts.SyntheticsCondition{
				Name:       "Check Failure",
				MonitorID:  *s.Status.ID,
				RunbookURL: *item.RunbookURL,
				Enabled:    true,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Monitor) getCurrent(ctx context.Context) (*synthetics.Monitor, error) {
	id := s.Status.GetID()
	if id == nil {
		return nil, errors.New("missing id")
	}

	data, err := client.Synthetics.GetMonitor(string(*id))
	if err != nil {
		s.Status.Info = err.Error()
		return nil, err
	}
	return data, nil
}

func (s *Monitor) findPolicyID(ctx context.Context, name string) (*int, error) {
	policies, err := client.Alerts.ListPolicies(&alerts.ListPoliciesParams{Name: name})
	if err != nil {
		s.Status.Info = err.Error()
		return nil, err
	}

	if len(policies) == 1 {
		return &policies[0].ID, nil
	}

	if len(policies) > 0 {
		err = fmt.Errorf("expected a policy search by name to only return 1 result, but found multiple for %s", name)
		return nil, err
	}

	return nil, fmt.Errorf("unable to find policy %s", name)
}
