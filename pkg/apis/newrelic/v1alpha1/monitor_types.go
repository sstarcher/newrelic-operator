package v1alpha1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/apex/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MonitorSpec defines the desired state of Monitor
type MonitorSpec struct {
	Type          *string              `json:"type,omitempty"`
	Frequency     *int64               `json:"frequency,omitempty"`
	URI           *string              `json:"uri,omitempty"`
	Locations     []*string            `json:"locations,omitempty"`
	Status        *MonitorStatusString `json:"status,omitempty"`
	SLAThreshold  *float64             `json:"slaThreshold,omitempty"`
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

func (s *Monitor) toNewRelic() (*newrelic.Monitor, error) {
	data := &newrelic.Monitor{
		Name:      &s.ObjectMeta.Name,
		Type:      s.Spec.Type,
		Frequency: s.Spec.Frequency,
		URI:       s.Spec.URI,
		Locations: s.Spec.Locations,
		Script:    &newrelic.Script{},
		Options: newrelic.MonitorOptions{
			ValidationString:       s.Spec.Options.ValidationString,
			VerifySSL:              s.Spec.Options.VerifySSL,
			BypassHEADRequest:      s.Spec.Options.BypassHEADRequest,
			TreatRedirectAsFailure: s.Spec.Options.TreatRedirectAsFailure,
		},
	}

	// slaThreshold, err := strconv.ParseFloat(s.Spec.SLAThreshold, 64)
	// if err != nil {
	// 	return nil, err
	// }
	data.SLAThreshold = s.Spec.SLAThreshold

	if s.Spec.Status != nil {
		status := s.Spec.Status.String()
		data.Status = &status
	}

	if data.Type == nil {
		val := string(typePing)
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

	return data, nil
}

// Create in newrelic
func (s *Monitor) Create(ctx context.Context) bool {
	logger := GetLogger(ctx)

	input, err := s.toNewRelic()
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	data, rsp, err := clientSythetics.SyntheticsMonitors.Create(ctx, input)
	if rsp.StatusCode == 400 {
		// TODO consider checking to improve error message
		logger.Info("this may already exist")
	}

	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	s.Status.ID = data.ID
	s.Status.Info = "Created"

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

	id := s.Status.ID
	if id == nil {
		logger.Info("object does not exist")
		return false
	}

	rsp, err := clientSythetics.SyntheticsMonitors.DeleteByID(ctx, id)
	if rsp.StatusCode == 404 {
		logger.Info("unable to find so skipping deletion", "id", id)
		return false
	}
	err = handleError(rsp, err)
	if err != nil {
		return true
	}

	return false
}

// GetID for the new relic object
func (s *Monitor) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return ""
}

// Update object in newrelic
func (s *Monitor) Update(ctx context.Context) bool {
	logger := GetLogger(ctx)

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
	rsp, err := clientSythetics.SyntheticsMonitors.Update(ctx, monitor, s.Status.ID)
	if rsp.StatusCode == 404 {
		s.Status.ID = nil
		logger.Info("id is missing recreating", "name", s.ObjectMeta.Name)
		return false
	}

	if s.Status.Handle(ctx, handleError(rsp, err), "failed") {
		logger.Info("duh fuck")
		logger.Info(s.Status.Info)
		return true
	}

	err = s.updateScript(ctx)
	if s.Status.Handle(ctx, err, "failed on script") {
		return true
	}

	err = s.updateCondition(ctx)
	if s.Status.Handle(ctx, err, "failed on condition") {
		return true
	}

	return false
}

func (s *Monitor) updateScript(ctx context.Context) error {
	if s.Spec.Type != nil && strings.ToUpper(*s.Spec.Type) == typeAPI && s.Spec.Script != nil && s.Spec.Script.ScriptText != nil {
		encoded := base64.StdEncoding.EncodeToString([]byte(*s.Spec.Script.ScriptText))
		rsp, err := clientSythetics.SyntheticsScript.UpdateByID(ctx, &newrelic.Script{ScriptText: &encoded}, *s.Status.ID)
		return handleError(rsp, err)
	}
	return nil
}

func (s *Monitor) updateCondition(ctx context.Context) error {
	if s.Spec.Conditions != nil {
		for _, item := range s.Spec.Conditions {
			var failureName = "Check Failure"
			var enabled = true
			cond := &newrelic.AlertsConditionEntity{
				AlertsSyntheticsConditionEntity: &newrelic.AlertsSyntheticsConditionEntity{
					AlertsSyntheticsCondition: &newrelic.AlertsSyntheticsCondition{
						Name:       &failureName,
						MonitorID:  s.Status.ID,
						RunbookURL: item.RunbookURL,
						Enabled:    &enabled,
					},
				},
			}

			policyID, err := s.findPolicyID(ctx, item.PolicyName)
			if err != nil {
				s.Status.Info = err.Error()
				return err
			}

			result, rsp, err := client.AlertsConditions.List(ctx, &newrelic.AlertsConditionsOptions{PolicyIDOptions: strconv.FormatInt(*policyID, 10)}, newrelic.ConditionSynthetics)
			err = handleError(rsp, err)
			if err != nil {
				s.Status.Info = err.Error()
				return err
			}

			exists := false
			if result != nil && result.AlertsSyntheticsConditionList != nil && result.AlertsSyntheticsConditionList.AlertsSyntheticsConditions != nil {
				for _, key := range result.AlertsSyntheticsConditionList.AlertsSyntheticsConditions {
					if *key.MonitorID == s.GetID() {
						exists = true
					}
				}
			}

			if exists {
				continue
			}

			_, rsp, err = client.AlertsConditions.Create(ctx, newrelic.ConditionSynthetics, cond, *policyID)
			err = handleError(rsp, err)
			if err != nil {
				s.Status.Info = err.Error()
				return err
			}
		}
	}

	return nil
}

func (s *Monitor) getCurrent(ctx context.Context) (*newrelic.Monitor, error) {
	data, rsp, err := clientSythetics.SyntheticsMonitors.GetByID(ctx, s.GetID())
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return nil, err
	}
	return data, nil
}

func (s *Monitor) findPolicyID(ctx context.Context, name string) (*int64, error) {
	policies, rsp, err := client.AlertsPolicies.ListAll(ctx, &newrelic.AlertsPolicyListOptions{
		NameOptions: name,
	})
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return nil, err
	}

	var id *int64
	for _, item := range policies.AlertsPolicies {
		if *item.Name == name {
			if id != nil {
				err = fmt.Errorf("expected a policy search by name to only return 1 result, but found multiple for %s", name)
				s.Status.Info = err.Error()
				return nil, err
			}
			id = item.ID
		}
	}

	return id, nil
}
