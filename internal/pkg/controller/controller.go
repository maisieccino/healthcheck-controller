package controller

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	healthv1alpha1 "github.com/mbellgb/healthcheck-controller/pkg/apis/health/v1alpha1"

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
	healthcheckInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueHealthCheck,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueHealthCheck(new)
		},
		DeleteFunc: controller.deleteCronJob,
	})
	// TODO: Add event handlers for cronjob resources.
	cronjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return controller
}

func (c *Controller) deleteCronJob(obj interface{}) {
	var (
		hc healthv1alpha1.HealthCheck
		ok bool
	)
	if hc, ok = obj.(healthv1alpha1.HealthCheck); !ok {
		utilruntime.HandleError(fmt.Errorf("could not decode object into HealthCheck"))
		fmt.Println(obj)
		return
	}

	// Find matching CronJob if any.
	cronjob, err := c.cronjobsLister.CronJobs(hc.GetNamespace()).Get(hc.Status.CronJobName)
	if errors.IsNotFound(err) {
		// "My work here is done."
		// "But you didn't do anything!"
		// "Oh, didn't I?"
		return
	}

	if err := c.kubeclientset.BatchV1beta1().CronJobs(cronjob.GetNamespace()).Delete(cronjob.GetName(), &metav1.DeleteOptions{}); err != nil {
		utilruntime.HandleError(err)
	}
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting HealthCheck controller")

	klog.Info("Waiting for caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cronjobsSynced, c.healthchecksSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		// run worker
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Killing workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// Work closure
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var (
			key string
			ok  bool
		)

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) enqueueHealthCheck(obj interface{}) {
	var (
		key string
		err error
	)

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
