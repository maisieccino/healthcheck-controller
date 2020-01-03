package controller

import (
	healthv1alpha1 "github.com/mbellgb/healthcheck-controller/pkg/apis/health/v1alpha1"
	"github.com/mbellgb/healthcheck-controller/pkg/generated/clientset/versioned/fake"
	informers "github.com/mbellgb/healthcheck-controller/pkg/generated/informers/externalversions"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/diff"
	"reflect"
	"testing"
	"time"
)

type testCase struct {
	t           *testing.T
	client      *fake.Clientset
	kubeclient  *k8sfake.Clientset
	hcLister    []*healthv1alpha1.HealthCheck
	cjLister    []*batchv1beta1.CronJob
	kubeActions []core.Action
	actions     []core.Action
	kubeObjects []runtime.Object
	objects     []runtime.Object
}

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

func newTestCase(t *testing.T) *testCase {
	return &testCase{
		t:           t,
		objects:     []runtime.Object{},
		kubeObjects: []runtime.Object{},
	}
}

func newHealthCheck(name, image, frequency, cronPattern string, args []string) *healthv1alpha1.HealthCheck {
	return &healthv1alpha1.HealthCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec:       healthv1alpha1.HealthCheckSpec{
			Image:       image,
			Frequency:   frequency,
			CronPattern: cronPattern,
			Args:        args,
		},
	}
}

func (tc *testCase) newController() (*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	tc.client = fake.NewSimpleClientset(tc.objects...)
	tc.kubeclient = k8sfake.NewSimpleClientset(tc.kubeObjects...)
	i := informers.NewSharedInformerFactory(tc.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(tc.kubeclient, noResyncPeriodFunc())

	c := NewController(
		tc.kubeclient,
		tc.client,
		k8sI.Batch().V1beta1().CronJobs(),
		i.Health().V1alpha1().HealthChecks(),
	)
	c.cronjobsSynced = alwaysReady
	c.healthchecksSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, hc := range tc.hcLister {
		i.Health().V1alpha1().HealthChecks().Informer().GetIndexer().Add(hc)
	}
	for _, cj := range tc.cjLister {
		k8sI.Batch().V1beta1().CronJobs().Informer().GetIndexer().Add(cj)
	}

	return c, i, k8sI
}

func (tc *testCase) run(hcName string) {
	tc.runController(hcName, true, false)
}

func (tc *testCase) runExpectError(hcName string) {
	tc.runController(hcName, true, true)
}

func (tc *testCase) runController(hcName string, startInformers, expectError bool) {
	c, i, k8sI := tc.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.syncHandler(hcName)
	if !expectError && err != nil {
		tc.t.Errorf("error syncing HealthCheck: %v", err)
	} else if expectError && err == nil {
		tc.t.Errorf("expected error syncing HealthCheck, got nil")
	}

	actions := filterInformerActions(tc.client.Actions())
	for i, action := range actions {
		if len(tc.actions) < i+1 {
			tc.t.Errorf("%d unexpected actions: %+v", len(actions)-len(tc.actions), actions[i:])
			break
		}
		expectedAction := tc.actions[i]
		checkAction(tc.t, expectedAction, action)
	}
	if len(tc.actions) > len(actions) {
		tc.t.Errorf("%d additional expected actions:%+v", len(tc.actions)-len(actions), tc.actions[len(actions):])
	}

	k8sActions := filterInformerActions(tc.kubeclient.Actions())
	for i, action := range k8sActions {
		if len(tc.kubeActions) < i+1 {
			tc.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(tc.kubeActions), k8sActions[i:])
			break
		}
		expectedAction := tc.kubeActions[i]
		checkAction(tc.t, expectedAction, action)
	}
	if len(tc.kubeActions) > len(k8sActions) {
		tc.t.Errorf("%d additional expected actions:%+v", len(tc.kubeActions)-len(k8sActions), tc.kubeActions[len(k8sActions):])
	}
}

func checkAction(t *testing.T, expected, actual core.Action) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("action has the wrong type, expected %t but got %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject,object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expObject := e.GetPatch()
		object := a.GetPatch()

		if !reflect.DeepEqual(expObject,object) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	default:
		t.Errorf("Uncaptured action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "healthchecks") ||
				action.Matches("watch", "healthchecks") ||
				action.Matches("list", "cronjobs") ||
				action.Matches("watch", "cronjobs")) {
			continue
		}
		ret = append(ret, action)
	}
	return ret
}

func (tc *testCase) expectCreateCronJobAction(cj *batchv1beta1.CronJob) {
	tc.kubeActions = append(tc.kubeActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "cronjobs"}, cj.Namespace, cj))
}

func (tc *testCase) expectUpdateCronJobAction(cj *batchv1beta1.CronJob) {
	tc.kubeActions = append(tc.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "cronjobs"}, cj.Namespace, cj))
}

func (tc *testCase) expectUpdateHealthCheckStatusAction(hc *healthv1alpha1.HealthCheck, cronJobName string) {
	hc.Status.CronJobName = cronJobName
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "healthchecks"}, hc.Namespace, hc)
	//action.Subresource = "status"
	tc.actions = append(tc.actions, action)
}

func getKey(t *testing.T, hc *healthv1alpha1.HealthCheck) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(hc)
	if err != nil {
		t.Errorf("unexpected error getting key for HealthCheck %v: %v", hc.Name, err)
		return ""
	}
	return key
}

func TestCreatesCronJob(t *testing.T) {
	tc := newTestCase(t)
	healthCheckName := "foo"
	hc := newHealthCheck(healthCheckName, "nginx", "", "* * * * *", nil)

	tc.hcLister = append(tc.hcLister, hc)
	tc.objects = append(tc.objects, hc)

	expectedCronJob := newCronJob(hc, healthCheckName)
	tc.expectCreateCronJobAction(expectedCronJob)
	tc.expectUpdateHealthCheckStatusAction(hc, healthCheckName)
	tc.run(getKey(t,hc))
}

func TestDoNothing(t *testing.T) {
	tc := newTestCase(t)
	healthCheckName := "foo"
	hc := newHealthCheck(healthCheckName, "nginx", "", "* * * * *", nil)
	cj := newCronJob(hc, healthCheckName)

	tc.hcLister = append(tc.hcLister, hc)
	tc.objects = append(tc.objects, hc)
	tc.cjLister = append(tc.cjLister, cj)
	tc.kubeObjects = append(tc.kubeObjects, cj)

	tc.expectUpdateHealthCheckStatusAction(hc, healthCheckName)
	tc.run(getKey(t,hc))
}

func TestUpdateCronJob(t *testing.T) {
	tc := newTestCase(t)
	healthCheckName := "foo"
	hc := newHealthCheck(healthCheckName, "nginx", "", "* * * * *", nil)
	cj := newCronJob(hc, healthCheckName)

	// Update HealthCheck image.
	hc.Spec.Image = "busybox"
	expectedCronJob := newCronJob(hc, healthCheckName)
	tc.hcLister = append(tc.hcLister, hc)
	tc.objects = append(tc.objects, hc)
	tc.cjLister = append(tc.cjLister, cj)
	tc.kubeObjects = append(tc.kubeObjects, cj)

	tc.expectUpdateHealthCheckStatusAction(hc, healthCheckName)
	tc.expectUpdateCronJobAction(expectedCronJob)
	tc.run(getKey(t, hc))
}

func TestNotControlledByHCController(t *testing.T) {
	tc := newTestCase(t)
	healthCheckName := "foo"
	hc := newHealthCheck(healthCheckName, "nginx", "", "* * * * *", nil)
	cj := newCronJob(hc, healthCheckName)

	// CronJob not owned by this controller.
	cj.ObjectMeta.OwnerReferences = []metav1.OwnerReference{}

	tc.hcLister = append(tc.hcLister, hc)
	tc.objects = append(tc.objects, hc)
	tc.cjLister = append(tc.cjLister, cj)
	tc.kubeObjects = append(tc.kubeObjects, cj)

	tc.runExpectError(getKey(t, hc))
}
