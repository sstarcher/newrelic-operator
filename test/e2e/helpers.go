package e2e

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func WaitForStatusPolicy(policy *v1alpha1.AlertPolicy) (err error) {
	operation := func() (err error) {
		namespaced := types.NamespacedName{Namespace: policy.ObjectMeta.Namespace, Name: policy.ObjectMeta.Name}
		err = framework.Global.Client.Get(context.TODO(), namespaced, policy)
		if err != nil {
			return
		} else if policy.Status.GetID() == nil {
			return errors.New("no id")
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
