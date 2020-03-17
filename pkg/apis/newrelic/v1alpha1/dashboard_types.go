package v1alpha1

import (
	"context"

	"github.com/newrelic/newrelic-client-go/pkg/dashboards"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dashboard is the Schema for the dashboards API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=dashboards,scope=Namespaced
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DashboardSpec `json:"spec"`
	Status            Status        `json:"status,omitempty"`
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

// DashboardSpec defines the structure of the dashboard for new relic
type DashboardSpec struct {
	Title string `json:"title,omitempty"`
	Icon  string `json:"icon,omitempty"`
	// TODO
	// Widgets    []dashboards.DashboardWidget `json:"widgets,omitempty"`
	Visibility string `json:"visibility,omitempty"`
	Editable   string `json:"editable,omitempty"`
	// Filter      `json:"filter,omitempty"`
}

var _ CRD = &Dashboard{}

// IsCreated specifies if the object has been created in new relic yet
func (s *Dashboard) IsCreated() bool {
	return s.Status.IsCreated()
}

func (s *Dashboard) toNewRelic() (*dashboards.Dashboard, error) {
	data := &dashboards.Dashboard{
		Title:      s.Spec.Title,
		Icon:       s.Spec.Icon,
		Visibility: dashboards.VisibilityType(s.Spec.Visibility),
		Editable:   dashboards.EditableType(s.Spec.Editable),
	}

	if s.Status.ID != nil {
		data.ID = int(*s.Status.GetID())
	}

	return data, nil
}

// Create in newrelic
func (s *Dashboard) Create(ctx context.Context) bool {
	input, err := s.toNewRelic()
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	rsp, err := client.Dashboards.CreateDashboard(*input)
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	s.Status.Info = "Created"
	s.Status.SetID(rsp.ID)
	s.SetFinalizers([]string{finalizer})
	return false
}

// Delete in newrelic
func (s *Dashboard) Delete(ctx context.Context) bool {
	logger := GetLogger(ctx)

	id := s.Status.GetID()
	if id == nil {
		logger.Info("skipping deletion ID is missing from object")
		return false
	}

	_, err := client.Dashboards.DeleteDashboard(int(*id))
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	return false
}

// Update object in newrelic
func (s *Dashboard) Update(ctx context.Context) bool {
	input, err := s.toNewRelic()
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	_, err = client.Dashboards.UpdateDashboard(*input)
	if s.Status.HandleOnError(ctx, err) {
		return true
	}

	return false
}
