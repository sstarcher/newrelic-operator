package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
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
		sdk.Update(s)
		return err
	}

	created(*data.ID, &s.Status, &s.Spec)
	return sdk.Update(s)
}

// Delete in newrelic
func (s *Monitor) Delete(ctx context.Context) error {
	if s.Status.ID == nil {
		return fmt.Errorf("alert Policy object has not been created %s", s.ObjectMeta.Name)
	}

	rsp, err := clientSythetics.SyntheticsMonitors.DeleteByID(ctx, s.Status.ID)
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
		sdk.Update(s)
		return err
	}

	update(&s.Spec, &s.Status)
	return sdk.Update(s)
}
