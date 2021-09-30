package trino

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	v12 "trino-operator/apis/tarim/v1"
	"trino-operator/pkg/apis/clientset/versioned"
)

type DeploymentController struct {
	KubeClient  *kubernetes.Clientset
	TrinoClinet *versioned.Clientset
}

func (t *DeploymentController) OnAdd(obj interface{}) {
	//controller, ok := obj.(*v1.Deployment)
	//if !ok {
	//	klog.Info("error to process OnAdd", controller)
	//
	//	return
	//}
	//klog.Info("OnAdd  deploy: ", controller.GetName())
}

func (t *DeploymentController) OnUpdate(oldObj, newObj interface{}) {
	deployNewObj, ok2 := newObj.(*v1.Deployment)

	if !ok2 {
		klog.Infof("error to update deploy ,name: %s ", deployNewObj.GetName())
		return
	}
	if deployNewObj.DeletionTimestamp != nil {
		klog.Infof("skip to update deploy  ,name: %s", deployNewObj.GetName())
		return
	}

	klog.Infof(" process update deploy, name: %s", deployNewObj.GetName())

	references := deployNewObj.GetOwnerReferences()
	if len(references) == 0 || references[0].APIVersion != v12.GroupVersion.String() {
		klog.Infof("process deploy, name:  %s,  skip", deployNewObj.GetName())
		return
	}
	reference := references[0]
	trino, err := t.TrinoClinet.TarimV1().Trinos(deployNewObj.Namespace).Get(context.Background(), reference.Name, v13.GetOptions{})
	if err != nil {
		klog.Infof(" process update deploy, error to get trino crd, name: %s", reference.Name)
		return
	}
	coordinatorDeploy, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Get(context.Background(), GetCoordinatorName(trino.Name), v13.GetOptions{})
	if err != nil {
		klog.Infof(" process update deploy, ,error to get trino crd, name: %s ,deploy %s not found ", reference.Name, GetCoordinatorName(trino.Name))
		return
	}
	workerDeploy, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Get(context.Background(), GetWorkerName(trino.Name), v13.GetOptions{})
	if err != nil {
		klog.Infof(" process update deploy, ,error to get trino crd, name: %s,deploy %s not found ", reference.Name, GetCoordinatorName(trino.Name))
		return
	}
	if trino.Spec.Pause {
		trino.Status.Status = STOPPED
	} else if coordinatorDeploy.Status.ReadyReplicas != coordinatorDeploy.Status.Replicas || workerDeploy.Status.ReadyReplicas != workerDeploy.Status.Replicas {
		trino.Status.Status = TRANSITIONING
	} else {
		trino.Status.Status = RUNNING
	}
	labelSetCoordinator := labels.Set{
		"app":       trino.Name,
		"component": "coordinator",
	}

	trino.Status.CoordinatorPod = []v12.PodStatus{}
	trino.Status.WorkerPod = []v12.PodStatus{}
	//pod status
	if !trino.Spec.Pause {
		// coordinator pod
		list, err := t.KubeClient.CoreV1().Pods(trino.Namespace).List(context.Background(), v13.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelSetCoordinator).String(),
		})
		if err != nil {
			return
		}
		for _, pod := range list.Items {
			trino.Status.CoordinatorPod = append(trino.Status.CoordinatorPod,
				v12.PodStatus{
					Name:      pod.GetName(),
					Cpu:       fmt.Sprintf("%d", trino.Spec.CoordinatorCpu),
					Memory:    fmt.Sprintf("%d", trino.Spec.CoordinatorMemory),
					PodStatus: string(pod.Status.Phase),
				})
		}

		// worker pod
		labelSetWorker := labels.Set{
			"app":       trino.Name,
			"component": "worker",
		}
		listWorker, err := t.KubeClient.CoreV1().Pods(trino.Namespace).List(context.Background(), v13.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelSetWorker).String(),
		})
		if err != nil {
			return
		}
		for _, pod := range listWorker.Items {
			trino.Status.WorkerPod = append(trino.Status.WorkerPod,
				v12.PodStatus{
					Name:      pod.GetName(),
					Cpu:       fmt.Sprintf("%d", trino.Spec.CoordinatorCpu),
					Memory:    fmt.Sprintf("%d", trino.Spec.CoordinatorMemory),
					PodStatus: string(pod.Status.Phase),
				})
		}
	}

	//save
	_, err = t.TrinoClinet.TarimV1().Trinos(trino.Namespace).UpdateStatus(context.Background(), trino, v13.UpdateOptions{})
	if err != nil {
		klog.Infof("process update deploy, error to get trino crd, name: %s ,error ", trino.Name, err)
		return
	}
	klog.Infof("process update deploy success, crd name: %s ", reference.Name)

}

func (t *DeploymentController) OnDelete(obj interface{}) {
	//controller, ok := obj.(*v1.Deployment)
	//if !ok {
	//	klog.Info("error to process OnDelete", controller)
	//
	//	return
	//}
	//
	//klog.Info("OnDelete  deploy: ", controller.GetName())
}
