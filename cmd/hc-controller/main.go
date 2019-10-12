package main

import (
	"flag"
	"time"

	healthcontroller "github.com/mbellgb/healthcheck-controller/internal/pkg/controller"
	"github.com/mbellgb/healthcheck-controller/internal/pkg/signals"
	clientset "github.com/mbellgb/healthcheck-controller/pkg/generated/clientset/versioned"
	healthinformers "github.com/mbellgb/healthcheck-controller/pkg/generated/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubeconfig string
	masterURL  string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	healthClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building health client: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	healthInformerFactory := healthinformers.NewSharedInformerFactory(healthClient, time.Second*30)

	controller := healthcontroller.NewController(
		kubeClient,
		healthClient,
		kubeInformerFactory.Batch().V1beta1().CronJobs(),
		healthInformerFactory.Health().V1alpha1().HealthChecks(),
	)

	kubeInformerFactory.Start(stopCh)
	healthInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error starting controller: %s\n", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig if out of cluster. Ignore to use in-cluster-config.")
	flag.StringVar(&masterURL, "master", "", "Address of k8s API if out of cluster. Ignore to use in-cluster-config.")
}
