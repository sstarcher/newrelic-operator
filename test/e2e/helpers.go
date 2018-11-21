package e2e

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewObjectMeta(name string, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
}

func NewTypeMeta(kind string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: "newrelic.shanestarcher.com/v1alpha1",
	}
}

func WaitForStatusPolicy(item *v1alpha1.AlertPolicy) (err error) {
	operation := func() (err error) {
		namespaced := types.NamespacedName{Namespace: item.ObjectMeta.Namespace, Name: item.ObjectMeta.Name}
		err = framework.Global.Client.Get(context.TODO(), namespaced, item)
		if err != nil {
			return
		} else if item.Status.GetID() == nil {
			return errors.New(item.Status.Info)
		}
		return
	}

	err = backoff.Retry(operation, ShortBackoff())
	if err != nil {
		// Handle error.
		return
	}
	return
}

func WaitForStatusMonitor(item *v1alpha1.Monitor) (err error) {
	operation := func() (err error) {
		namespaced := types.NamespacedName{Namespace: item.ObjectMeta.Namespace, Name: item.ObjectMeta.Name}
		err = framework.Global.Client.Get(context.TODO(), namespaced, item)
		if err != nil {
			return
		} else if item.Status.GetID() == nil {
			return errors.New(item.Status.Info)
		}
		return
	}

	err = backoff.Retry(operation, ShortBackoff())
	if err != nil {
		// Handle error.
		return
	}
	return
}

func WaitForStatusChannel(item *v1alpha1.AlertChannel) (err error) {
	operation := func() (err error) {
		namespaced := types.NamespacedName{Namespace: item.ObjectMeta.Namespace, Name: item.ObjectMeta.Name}
		err = framework.Global.Client.Get(context.TODO(), namespaced, item)
		if err != nil {
			return
		} else if item.Status.GetID() == nil {
			return errors.New(item.Status.Info)
		}
		return
	}

	err = backoff.Retry(operation, ShortBackoff())
	if err != nil {
		// Handle error.
		return
	}
	return
}

func ShortBackoff() *backoff.ExponentialBackOff {
	back := backoff.NewExponentialBackOff()
	back.MaxElapsedTime = 10 * time.Second
	return back
}

func Create(ctx *framework.TestCtx, obj runtime.Object) error {
	return framework.Global.Client.Create(context.TODO(), obj, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
}
