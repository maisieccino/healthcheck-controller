package controller

import (
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	healthv1alpha1 "github.com/mbellgb/healthcheck-controller/pkg/apis/health/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	defaultCronPattern = "*/1 * * * *"
)

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key '%s'", key))
		return nil
	}

	healthcheck, err := c.healthchecksLister.HealthChecks(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("HealthCheck '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	cronjobName := healthcheck.Status.CronJobName
	if cronjobName == "" {
		cronjobName = healthcheck.GetName()
	}

	// Get specified CronJob.
	cronjob, err := c.cronjobsLister.CronJobs(healthcheck.GetNamespace()).Get(cronjobName)
	// If not found, create a new one.
	if errors.IsNotFound(err) {
		cronjob, err = c.kubeclientset.BatchV1beta1().CronJobs(healthcheck.GetNamespace()).Create(newCronJob(healthcheck, cronjobName))
	}

	// Throw error so the work item can be retried.
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(cronjob, healthcheck) {
		msg := fmt.Sprintf(MessageResourceExists, cronjob.GetName())
		c.recorder.Event(healthcheck, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	newCronjob := newCronJob(healthcheck, cronjobName)
	if !reflect.DeepEqual(cronjob.Spec, newCronjob.Spec) {
		klog.V(4).Infof("Updating CronJob '%s' to reflect changes from HealthCheck '%s'", cronjob.GetName(), healthcheck.GetName())
		cronjob, err = c.kubeclientset.BatchV1beta1().CronJobs(healthcheck.GetNamespace()).Update(newCronjob)
	}

	// Throw error so the work item can be retried.
	if err != nil {
		return err
	}

	err = c.updateHealthCheckStatus(healthcheck, cronjob)
	if err != nil {
		return err
	}

	c.recorder.Event(healthcheck, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateHealthCheckStatus(hc *healthv1alpha1.HealthCheck, cronjob *batchv1beta1.CronJob) error {
	healthcheckCopy := hc.DeepCopy()
	healthcheckCopy.Status.CronJobName = cronjob.GetName()
	_, err := c.healthclientset.HealthV1alpha1().HealthChecks(hc.GetNamespace()).Update(healthcheckCopy)
	return err
}

func newCronJob(hc *healthv1alpha1.HealthCheck, name string) *batchv1beta1.CronJob {
	labels := map[string]string{
		"controller": hc.GetName(),
	}

	schedule := defaultCronPattern
	if len(hc.Spec.CronPattern) > 0 {
		schedule = hc.Spec.CronPattern
	}

	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: hc.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(hc, healthv1alpha1.SchemeGroupVersion.WithKind("HealthCheck")),
			},
		},
		Spec: batchv1beta1.CronJobSpec{
			FailedJobsHistoryLimit:     int32Ptr(10),
			SuccessfulJobsHistoryLimit: int32Ptr(10),
			ConcurrencyPolicy:          batchv1beta1.ForbidConcurrent,
			StartingDeadlineSeconds:    int64Ptr(10),
			Schedule:                   schedule,
			Suspend:                    boolPtr(false),
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					BackoffLimit: int32Ptr(0),
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: labels,
						},
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "healthcheck",
									Image: hc.Spec.Image,
									Args:  hc.Spec.Args,
								},
							},
						},
					},
				},
			},
		},
	}
}

func boolPtr(b bool) *bool    { return &b }
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
