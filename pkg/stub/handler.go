package stub

import (
	"context"

	"github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func NewHandler(m *Metrics) sdk.Handler {
	return &Handler{
		metrics: m,
	}
}

type Metrics struct {
	operatorErrors prometheus.Counter
}

type Handler struct {
	// Metrics example
	metrics *Metrics
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case v1alpha1.CRD:
		logger := log.WithFields(log.Fields{"signature": o.Signature()})
		if event.Deleted {
			logger.Infof("deleting")
			o.Delete(ctx)
		} else if o.IsCreated() {
			if o.HasChanged() {
				logger.Infof("update %s", o.GetID())
				o.Update(ctx)
			}
		} else {
			logger.Info("creating")
			err := o.Create(ctx)
			if err != nil {
				logger.Error(err)
			}
		}
	default:
		log.Warnf("recieved a event we don't have a CRD for %s", o)
	}
	return nil
}

func RegisterOperatorMetrics() (*Metrics, error) {
	operatorErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "memcached_operator_reconcile_errors_total",
		Help: "Number of errors that occurred while reconciling the memcached deployment",
	})
	err := prometheus.Register(operatorErrors)
	if err != nil {
		return nil, err
	}
	return &Metrics{operatorErrors: operatorErrors}, nil
}
