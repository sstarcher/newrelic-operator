package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/IBM/newrelic-cli/newrelic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dashboard is the Schema for the dashboards API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=dashboards,scope=Namespaced
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec   `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DashboardList contains a list of Dashboard
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Dashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dashboard{}, &DashboardList{})
}

// Additional Code

var _ CRD = &Dashboard{}

// IsCreated specifies if the object has been created in new relic yet
func (s *Dashboard) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *Dashboard) HasChanged() bool {
	return hasChanged(&s.Spec, &s.Status)
}

// Create in newrelic
func (s *Dashboard) Create(ctx context.Context) bool {
	rsp, data, err := client.Dashboards.Create(ctx, s.Spec.Data)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	var result newrelic.CreateDashboardResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return true
	}

	createdInt(*result.Dashboard.ID, &s.Status, &s.Spec)
	s.SetFinalizers([]string{finalizer})
	return false
}

// Delete in newrelic
func (s *Dashboard) Delete(ctx context.Context) bool {
	logger := GetLogger(ctx)

	id := s.Status.GetID()
	if id == nil {
		logger.Info("object does not exist")
		return false
	}

	rsp, _, err := client.Dashboards.DeleteByID(ctx, *id)
	if rsp.StatusCode == 404 {
		return false
	}
	err = handleError(rsp, err)
	if err != nil {
		return true
	}

	return false
}

// GetID for the new relic object
func (s *Dashboard) GetID() string {
	if s.Status.ID != nil {
		return *s.Status.ID
	}
	return "nil"
}

// Update object in newrelic
func (s *Dashboard) Update(ctx context.Context) bool {
	logger := GetLogger(ctx)
	id := s.Status.GetID()
	if id == nil {
		logger.Info("object has already been deleted")
		return false
	}

	rsp, _, err := client.Dashboards.Update(ctx, s.Spec.Data, *id)
	err = handleError(rsp, err)
	if err != nil {
		s.Status.Info = err.Error()
		return true
	}

	update(&s.Spec, &s.Status)
	return false
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
