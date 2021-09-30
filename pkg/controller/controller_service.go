package controller

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
	"trino-operator/pkg/apis/clientset/versioned"
	"trino-operator/pkg/apis/informers/externalversions"
	"trino-operator/pkg/controller/trino"
)

type ControllerService struct {
	KubeConfig    *rest.Config
	TrinoInformer *cache.SharedIndexInformer
	PodInformer   *cache.SharedIndexInformer
	StopSignalCh  <-chan struct{}
}

func (s *ControllerService) InitConfig() {
	s.KubeConfig = ctrl.GetConfigOrDie()
}

func (s *ControllerService) StartInformer() {
	s.InitTrinoInformer()
	s.InitDeploymentInformer()

}

func (s *ControllerService) InitDeploymentInformer() {
	klog.Info("start deployment informer")
	clientset, err := kubernetes.NewForConfig(s.KubeConfig)
	if err != nil {
		panic("deployment informer init faild")
		klog.Errorln("deployment informer init faild")
	}
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)
	informer := sharedInformerFactory.Apps().V1().Deployments().Informer()

	handler := &trino.DeploymentController{}
	handler.KubeClient = clientset
	handler.TrinoClinet = versioned.NewForConfigOrDie(s.KubeConfig)
	informer.AddEventHandler(handler)

	s.PodInformer = &informer
	sharedInformerFactory.Start(s.StopSignalCh)
	cache.WaitForCacheSync(s.StopSignalCh, informer.HasSynced)
	klog.Info("start deployment informer cached")
}

func (s *ControllerService) InitTrinoInformer() {
	klog.Info("start trino informer")
	clientset, err := versioned.NewForConfig(s.KubeConfig)
	if err != nil {
		klog.Errorln("trino informer init faild")
	}
	sharedInformerFactory := externalversions.NewSharedInformerFactory(clientset, time.Minute)
	informer := sharedInformerFactory.Tarim().V1().Trinos().Informer()
	handler := &trino.TrinoContriller{}
	handler.KubeClient = kubernetes.NewForConfigOrDie(s.KubeConfig)
	handler.TrinoClinet = clientset

	informer.AddEventHandler(handler)
	s.TrinoInformer = &informer
	sharedInformerFactory.Start(s.StopSignalCh)
	cache.WaitForCacheSync(s.StopSignalCh, informer.HasSynced)
	klog.Info("start trino informer cached")
}
