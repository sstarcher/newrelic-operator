package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sstarcher/newrelic-operator/pkg/apis"
	v1alpha1 "github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"

	"github.com/IBM/newrelic-cli/newrelic"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestMonitor(t *testing.T) {

	alertPolicyList := &v1alpha1.AlertPolicyList{
		TypeMeta: NewTypeMeta("AlertPolicy"),
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, alertPolicyList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	alertChannelList := &v1alpha1.AlertChannelList{
		TypeMeta: NewTypeMeta("AlertChannel"),
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, alertChannelList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	dashboardList := &v1alpha1.DashboardList{
		TypeMeta: NewTypeMeta("Dashboard"),
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, dashboardList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	monitorList := &v1alpha1.MonitorList{
		TypeMeta: NewTypeMeta("Monitor"),
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, monitorList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// run subtests
	t.Run("test-group", func(t *testing.T) {
		t.Run("Monitor", Monitor)
	})

}

func Monitor(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for newrelic-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "newrelic-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// create memcached custom resource
	policy := &v1alpha1.AlertPolicy{
		TypeMeta:   NewTypeMeta("AlertPolicy"),
		ObjectMeta: NewObjectMeta("policy", namespace),
		Spec: v1alpha1.AlertPolicySpec{
			IncidentPreference: string(newrelic.IncidentPerCondition),
		},
	}
	err = f.Client.Create(context.TODO(), policy, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForStatusPolicy(policy)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(policy.Status)

}
