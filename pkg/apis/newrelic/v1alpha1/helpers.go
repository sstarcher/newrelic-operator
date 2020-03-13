package v1alpha1

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	// L is an alias for the the standard logger.
	L = logf.Log.Logger
)

type (
	loggerKey struct{}
)

// WithLogger returns a new context with the provided logger.
func WithLogger(ctx context.Context, logger *logr.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger returns the logger from the context
func GetLogger(ctx context.Context) logr.Logger {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return L
	}

	return *(logger.(*logr.Logger))
}

// DefaultRequeue is the timing we by default to requeue
var DefaultRequeue = reconcile.Result{
	Requeue:      true,
	RequeueAfter: time.Minute * 5,
}

// DoReconcile generic processing loop
func DoReconcile(log logr.Logger, instance CRD) reconcile.Result {
	reconcileResult := reconcile.Result{}
	ctx := WithLogger(context.TODO(), &log)

	if instance.GetDeletionTimestamp() != nil {
		log.WithValues("action", "delete")
		log.Info("")
		if instance.Delete(ctx) {
			reconcileResult = DefaultRequeue
		} else {
			instance.SetFinalizers(nil)
		}
	} else if instance.IsCreated() {
		log.WithValues("action", "update")
		log.Info("")
		if instance.Update(ctx) {
			reconcileResult = DefaultRequeue
		}
	} else {
		log.WithValues("action", "create")
		log.Info("")
		if instance.Create(ctx) {
			reconcileResult = DefaultRequeue
		}
	}
	return reconcileResult
}
