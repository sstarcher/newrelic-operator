package dashboard

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	newrelicv1alpha1 "github.com/sstarcher/newrelic-operator/pkg/apis/newrelic/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var defaultRequeue = reconcile.Result{
	Requeue:      true,
	RequeueAfter: time.Minute * 5,
}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Dashboard Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDashboard{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dashboard-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Dashboard
	err = c.Watch(&source.Kind{Type: &newrelicv1alpha1.Dashboard{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDashboard{}

// ReconcileDashboard reconciles a Dashboard object
type ReconcileDashboard struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Dashboard object and makes changes based on the state read
// and what is in the Dashboard.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDashboard) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reconcileResult := reconcile.Result{}

	// Fetch the Dashboard instance
	instance := &newrelicv1alpha1.Dashboard{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcileResult, nil
		}
		// Error reading the object - requeue the request.
		return reconcileResult, err
	}

	logger := log.WithFields(log.Fields{"type": "dashboard", "name": request.Name, "namespace": request.Namespace})
	ctx := newrelicv1alpha1.WithLogger(context.TODO(), logger)

	if instance.GetDeletionTimestamp() != nil {
		logger.Infof("delete")
		err = instance.Delete(ctx)
		if err != nil {
			logger.Error(err)
			reconcileResult = defaultRequeue
		} else {
			instance.SetFinalizers(nil)
		}
	} else if instance.IsCreated() {
		if instance.HasChanged() {
			logger.Infof("update %s", instance.GetID())
			err = instance.Update(ctx)
			if err != nil {
				logger.Error(err)
				reconcileResult = defaultRequeue
			}
		}
	} else {
		logger.Info("create")
		err := instance.Create(ctx)
		if err != nil {
			logger.Error(err)
			reconcileResult = defaultRequeue
		}
	}

	return reconcileResult, r.client.Update(ctx, instance)
}
