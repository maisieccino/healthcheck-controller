package controller

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "healthcheck-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

// Controller manages HealthCheck resources.
type Controller struct {
	kubeclientset kubernetes.Interface
	workqueue     workqueue.RateLimitingInterface
}
