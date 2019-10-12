package controller

import (
	clientset "github.com/mbellgb/healthcheck-controller/pkg/generated/clientset/versioned"
	informers "github.com/mbellgb/healthcheck-controller/pkg/generated/informers/externalversions/health/v1alpha1"
	listers "github.com/mbellgb/healthcheck-controller/pkg/generated/listers/health/v1alpha1"
	batchinformers "k8s.io/client-go/informers/batch/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	batchlisters "k8s.io/client-go/listers/batch/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"

	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	healthscheme "github.com/mbellgb/healthcheck-controller/pkg/generated/clientset/versioned/scheme"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const controllerAgentName = "healthcheck-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a HealthCheck is
	// synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a
	// HealthCheck fails to sync due to a Deployment of the same name already
	// existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by HealthCheck"
	// MessageResourceSynced is the message used for an Event fired when a
	// HealthCheck is synced successfully
	MessageResourceSynced = "HealthCheck synced successfully"
)

// Controller manages HealthCheck resources.
type Controller struct {
	kubeclientset   kubernetes.Interface
	healthclientset clientset.Interface

	cronjobsLister     batchlisters.CronJobLister
	cronjobsSynced     cache.InformerSynced
	healthchecksLister listers.HealthCheckLister
	healthchecksSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController creates a new healthcheck controller.
func NewController(
	kubeclientset kubernetes.Interface,
	healthclientset clientset.Interface,
	cronjobInformer batchinformers.CronJobInformer,
	healthcheckInformer informers.HealthCheckInformer,
) *Controller {
	utilruntime.Must(healthscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	controller := &Controller{
		kubeclientset:      kubeclientset,
		healthclientset:    healthclientset,
		cronjobsLister:     cronjobInformer.Lister(),
		cronjobsSynced:     cronjobInformer.Informer().HasSynced,
		healthchecksLister: healthcheckInformer.Lister(),
		healthchecksSynced: healthcheckInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "HealthChecks"),
		recorder:           recorder,
	}

	klog.Info("Setting up event handlers")
	// TODO: Add event handlers for healthcheck objects
	healthcheckInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	cronjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	return nil
}
