package stub

import (
	"context"
	"os"

	"github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"
	"github.com/sstarcher/newrelic-operator/pkg/stub/newrelic"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func NewHandler(m *Metrics) sdk.Handler {
	client, err := newrelic.NewClient()
	if err != nil {
		panic(err)
	}

	clean := os.Getenv("NEW_RELIC_OPERATOR_CLEANUP")
	if clean != "" {
		list, err := client.ListDashboards(context.Background())
		if err != nil {
			panic(err)
		}
		for _, item := range list {
			if item.OwnerEmail != nil && *item.OwnerEmail == clean {
				err := client.DeleteDashboard(context.Background(), *item.ID)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return &Handler{
		metrics: m,
		client:  client,
	}
}

type Metrics struct {
	operatorErrors prometheus.Counter
}

type Handler struct {
	// Metrics example
	metrics *Metrics
	client  *newrelic.Client
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.AlertChannel:
		log.Printf("channel", o)
	case *v1alpha1.AlertPolicy:
		log.Printf("policy", o)
	case *v1alpha1.Dashboard:
		if event.Deleted {
			log.Printf("deleting dashboard %s/%s", o.Namespace, o.Name)
			if o.Status.ID != nil {
				return h.client.DeleteDashboard(ctx, *o.Status.ID)
			}
		} else if o.Status.IsCreated() {
			if o.HasChanged() {
				log.Printf("update dashboard %s/%s %d", o.Namespace, o.Name, o.Status.ID)
				err := h.client.UpdateDashboard(ctx, o.Spec.Data, *o.Status.ID)
				if err != nil {
					o.Status.Info = err.Error()
					return sdk.Update(o)
				}
				o.Update()
				sdk.Update(o)
			}
		} else {
			log.Printf("creating dashboard %s/%s", o.Namespace, o.Name)
			id, err := h.client.CreateDashboard(ctx, o.Spec.Data)
			if err != nil {
				o.Status.Info = err.Error()
				return sdk.Update(o)
			}

			o.Created(*id)
			sdk.Update(o)
		}
	case *v1alpha1.Label:
		log.Printf("label", o)
	case *v1alpha1.Monitor:
		log.Printf("monitor", o)
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
