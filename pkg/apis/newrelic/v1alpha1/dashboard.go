package v1alpha1

import (
	"context"
	"encoding/json"

	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ CRD = &Dashboard{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Dashboard `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec   `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

// IsCreated specifies if the object has been created in new relic yet
func (s *Dashboard) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *Dashboard) HasChanged() bool {
	return hasChanged(&s.Spec, &s.Status)
}

// Create in newrelic
func (s *Dashboard) Create(ctx context.Context) error {
	rsp, data, err := client.Dashboards.Create(ctx, s.Spec.Data)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		sdk.Update(s)
		return err
	}

	var result newrelic.CreateDashboardResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	createdInt(*result.Dashboard.ID, &s.Status, &s.Spec)
	sdk.Update(s)
	return nil
}

// Delete in newrelic
func (s *Dashboard) Delete(ctx context.Context) error {
	if s.Status.ID == nil {
		return fmt.Errorf("dashboard object has not been created %s", s.ObjectMeta.Name)
	}

	rsp, _, err := client.Dashboards.DeleteByID(ctx, s.Status.GetID())
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetID for the new relic object
func (s *Dashboard) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return "nil"
}

// Signature for the CRD
func (s *Dashboard) Signature() string {
	return fmt.Sprintf("%s %s/%s", s.TypeMeta.Kind, s.Namespace, s.Name)
}

// Update object in newrelic
func (s *Dashboard) Update(ctx context.Context) error {
	rsp, _, err := client.Dashboards.Update(ctx, s.Spec.Data, s.Status.GetID())
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return err
	}

	update(&s.Spec, &s.Status)
	sdk.Update(s)
	return nil
}

func listDashboards(ctx context.Context) ([]*newrelic.Dashboard, error) {
	rsp, data, err := client.Dashboards.ListAll(ctx, nil)
	err = handleError(rsp, err)
	if err != nil {
		return nil, err
	}

	var list newrelic.DashboardList
	err = json.Unmarshal(data, &list)
	if err != nil {
		return nil, err
	}

	return list.Dashboards, nil
}
