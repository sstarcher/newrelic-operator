package main

import (
	"context"
	"runtime"
	"time"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	stub "github.com/sstarcher/newrelic-operator/pkg/stub"

	"github.com/sirupsen/logrus"
	_ "github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	sdk.ExposeMetricsPort()
	metrics, err := stub.RegisterOperatorMetrics()
	if err != nil {
		logrus.Errorf("failed to register operator specific metrics: %v", err)
	}
	h := stub.NewHandler(metrics)

	resource := "newrelic.shanestarcher.com/v1alpha1"
	resyncPeriod := time.Duration(5) * time.Second
	logrus.Infof("Watching %s, %d", resource, resyncPeriod)
	sdk.Watch(resource, "AlertChannel", "", resyncPeriod)
	sdk.Watch(resource, "AlertPolicy", "", resyncPeriod)
	sdk.Watch(resource, "Dashboard", "", resyncPeriod)
	sdk.Watch(resource, "Label", "", resyncPeriod)
	sdk.Watch(resource, "Monitor", "", resyncPeriod)
	sdk.Handle(h)
	sdk.Run(context.TODO())
}
