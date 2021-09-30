package trino

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"reflect"
	"sort"
	"time"
	v12 "trino-operator/apis/tarim/v1"
	"trino-operator/pkg/apis/clientset/versioned"
	"trino-operator/pkg/common/config"
)

var (
	INSERT = "INSERT"
	UPDATE = "UPDATE"

	STOPPED       = "STOPPED"
	RUNNING       = "RUNNING"
	TRANSITIONING = "TRANSITIONING"
)

type TrinoContriller struct {
	KubeClient  *kubernetes.Clientset
	TrinoClinet *versioned.Clientset
}

func (t *TrinoContriller) OnAdd(obj interface{}) {
	trino, err := ToTrino(obj, "OnAdd")
	if err != nil {
		return
	}
	t.InsertOrUpdate(trino, INSERT)

}

func (t *TrinoContriller) OnUpdate(oldObj, newObj interface{}) {
	controller, ok := newObj.(*v12.Trino)
	if !ok {
		klog.Info("error to process Trino OnUpdate", newObj)
		return
	}
	klog.Info("OnUpdate  Trino: ", controller.GetName())
	trinoNew, err := ToTrino(newObj, "OnDelete")
	if err != nil {
		return
	}
	trinoOld, err := ToTrino(oldObj, "OnDelete")
	if err != nil {
		return
	}
	if reflect.DeepEqual(trinoNew.Spec, trinoOld.Spec) {
		klog.Info("Skip OnUpdate  Trino: version: ", trinoNew.ObjectMeta.ResourceVersion)
		return
	}
	t.Update(trinoOld, trinoNew)

}

func (t *TrinoContriller) Update(old *v12.Trino, trinoNew *v12.Trino) {
	t.InsertOrUpdate(trinoNew, UPDATE)
}

func (t *TrinoContriller) OnDelete(obj interface{}) {
	//trino, err := ToTrino(obj, "OnDelete")
	//if err != nil {
	//	return
	//}

}

func (t *TrinoContriller) InsertOrUpdate(trino *v12.Trino, status string) {
	reference := metav1.OwnerReference{}
	reference.Name = trino.Name
	reference.Kind = v12.Kind
	reference.APIVersion = v12.GroupVersion.String()
	reference.UID = trino.UID

	err, needRestart := t.InsertOrUpdateCatalogConfig(trino, reference, status)
	klog.Infof("InsertOrUpdateCatalogConfig trino err: %v, need update %t", err, needRestart)

	err, needRestartCoordinator := t.InsertOrUpdateCoordinatorConfig(trino, reference, status)
	klog.Infof("InsertOrUpdateCoordinatorConfig trino err: %v, need update %t", err, needRestart)

	err, needRestartWorker := t.InsertOrUpdateWorkConfig(trino, reference, status)
	klog.Infof("InsertWorkConfig trino err: %v, need update %t", err, needRestart)

	err = t.InsertService(trino, reference, status)
	klog.InfoS("InsertService", "trino", err)

	err = t.InsertOrUpdateCoordinatorDeploy(trino, reference, status, needRestart || needRestartCoordinator)
	klog.InfoS("InsertCoordinatorDeploy", "trino", err)

	err = t.InsertOrupdateWorkerDeploy(trino, reference, status, needRestart || needRestartWorker)
	klog.InfoS("InsertWorkerDeploy", "trino", err)

	err = t.InsertNodePort(trino, reference, status)
	klog.InfoS("InsertNodePort", "trino", err)

}

func (t *TrinoContriller) InsertOrUpdateCatalogConfig(trino *v12.Trino, reference metav1.OwnerReference, status string) (error, bool) {

	name := trino.Name + "-trino-catalog"
	var needRestart = false
	var configMap *v1.ConfigMap

	if status == UPDATE {
		configMapGet, err2 := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err2 != nil {
			return err2, needRestart
		}
		configMap = configMapGet

	} else {
		configMap = &v1.ConfigMap{}
		configMap.SetName(name)
		configMap.SetNamespace(trino.Namespace)
		configMap.OwnerReferences = append(configMap.OwnerReferences, reference)
		configMap.BinaryData = make(map[string][]byte)

	}
	configMap.Data = trino.Spec.CataLogConfig

	newHash := generatorMapHash(configMap.Data)
	fmt.Printf(" old hash : %x\n", configMap.BinaryData["hash"])
	if status == UPDATE && bytes.Compare(newHash, configMap.BinaryData["hash"]) != 0 {

		needRestart = true
	}
	fmt.Printf(" new hash : %x\n", newHash)

	configMap.BinaryData["hash"] = newHash

	if status == UPDATE {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Update(context.Background(), configMap, metav1.UpdateOptions{})
		return err, needRestart

	} else {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
		return err, needRestart
	}
}

func (t *TrinoContriller) InsertOrUpdateCoordinatorConfig(trino *v12.Trino, reference metav1.OwnerReference, status string) (error, bool) {
	name := trino.Name + "-trino-coordinator"
	var needRestart = false
	var configMap *v1.ConfigMap

	if status == UPDATE {
		configMapGet, err2 := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err2 != nil {
			return err2, needRestart
		}
		configMap = configMapGet
	} else {
		configMap = &v1.ConfigMap{}
		configMap.SetName(name)
		configMap.SetNamespace(trino.Namespace)
		configMap.OwnerReferences = append(configMap.OwnerReferences, reference)
		configMap.BinaryData = make(map[string][]byte)
	}
	//set data
	configMap.Data = make(map[string]string)
	configMap.Data["node.properties"] = trino.Spec.CoordinatorConfig.NodeProperties
	configMap.Data["jvm.config"] = trino.Spec.CoordinatorConfig.JvmConfig
	configMap.Data["config.properties"] = trino.Spec.CoordinatorConfig.ConfigProperties
	configMap.Data["log.properties"] = trino.Spec.CoordinatorConfig.LogProperties
	fmt.Printf(" old hash : %x\n", configMap.BinaryData["hash"])
	newHash := generatorMapHash(configMap.Data)
	if status == UPDATE && bytes.Compare(newHash, configMap.BinaryData["hash"]) != 0 {
		needRestart = true
	}

	configMap.BinaryData["hash"] = newHash
	fmt.Printf(" new hash : %x\n", newHash)

	if status == UPDATE {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Update(context.Background(), configMap, metav1.UpdateOptions{})
		return err, needRestart
	} else {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
		return err, needRestart
	}
}

func (t *TrinoContriller) InsertOrUpdateWorkConfig(trino *v12.Trino, reference metav1.OwnerReference, status string) (error, bool) {

	name := trino.Name + "-trino-worker"
	var needRestart = false
	var configMap *v1.ConfigMap

	if status == UPDATE {
		configMapGet, err2 := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err2 != nil {
			return err2, needRestart
		}
		configMap = configMapGet
	} else {
		configMap = &v1.ConfigMap{}
		configMap.SetName(name)
		configMap.SetNamespace(trino.Namespace)
		configMap.OwnerReferences = append(configMap.OwnerReferences, reference)
		configMap.BinaryData = make(map[string][]byte)

	}

	configMap.Data = make(map[string]string)
	configMap.Data["node.properties"] = trino.Spec.WorkConfig.NodeProperties
	configMap.Data["jvm.config"] = trino.Spec.WorkConfig.JvmConfig
	configMap.Data["config.properties"] = trino.Spec.WorkConfig.ConfigProperties
	configMap.Data["log.properties"] = trino.Spec.WorkConfig.LogProperties

	fmt.Printf(" old hash : %x\n", configMap.BinaryData["hash"])
	newHash := generatorMapHash(configMap.Data)

	if status == UPDATE && bytes.Compare(newHash, configMap.BinaryData["hash"]) != 0 {
		needRestart = true
	}
	configMap.BinaryData["hash"] = newHash
	fmt.Printf(" new hash : %x\n", newHash)

	if status == UPDATE {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Update(context.Background(), configMap, metav1.UpdateOptions{})
		return err, needRestart
	} else {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
		return err, needRestart
	}

}

func (t *TrinoContriller) InsertOrUpdateCoordinatorDeploy(trino *v12.Trino, reference metav1.OwnerReference, status string, restart bool) error {

	label := GetLabelMap(trino.Name, "coordinator")
	name := GetCoordinatorName(trino.Name)
	runAsGroup := int64(1000)
	runAsUser := int64(1000)
	var replicas int32
	if trino.Spec.Pause {
		replicas = 0
	} else {
		replicas = 1
	}

	quantityCpu, err2 := resource.ParseQuantity(fmt.Sprintf("%d%s", trino.Spec.CoordinatorCpu*1000, "m"))
	if err2 != nil {
		return err2
	}
	quantityMem, err2 := resource.ParseQuantity(fmt.Sprintf("%d", trino.Spec.CoordinatorMemory))
	if err2 != nil {
		return err2
	}
	var deploy *v13.Deployment

	baseEnv := []v1.EnvVar{}
	if status != UPDATE {
		//create
		deploy = &v13.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: trino.Namespace,
				Labels:    label,
			},
		}
		deploy.OwnerReferences = append(deploy.OwnerReferences, reference)
	} else {
		get, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		//update
		if err != nil {
			return err
		}
		baseEnv = get.Spec.Template.Spec.Containers[0].Env

		deploy = get
	}
	deploy.Spec = v13.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: label,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: label,
			},
			Spec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsUser:  &runAsUser,
					RunAsGroup: &runAsGroup,
				},
				Volumes: []v1.Volume{
					{
						Name: "config-volume",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: trino.Name + "-trino-coordinator",
								},
							},
						},
					},
					{
						Name: "catalog-volume",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: trino.Name + "-trino-catalog",
								},
							},
						},
					},
				},
				Containers: []v1.Container{
					{
						Name:  "trino-coordinator",
						Image: config.Image,
						Ports: []v1.ContainerPort{
							{
								Name:          "http",
								Protocol:      v1.Protocol("TCP"),
								ContainerPort: 8080,
							},
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config-volume",
								MountPath: "/etc/trino",
							}, {
								Name:      "catalog-volume",
								MountPath: "/etc/trino/catalog",
							},
						},
						ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/v1/info",
									Port: intstr.Parse("http"),
								},
							},
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/v1/info",
									Port: intstr.Parse("http"),
								},
							},
							InitialDelaySeconds: 20,
						},
						Resources: v1.ResourceRequirements{
							Requests: map[v1.ResourceName]resource.Quantity{
								v1.ResourceCPU:    quantityCpu,
								v1.ResourceMemory: quantityMem,
							},
						},
						Env: baseEnv,
					},
				},
			},
		},
		Strategy:                v13.DeploymentStrategy{},
		MinReadySeconds:         10,
		RevisionHistoryLimit:    nil,
		Paused:                  false,
		ProgressDeadlineSeconds: nil,
	}

	if restart {
		deploy.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{{Name: "start_time", Value: time.Now().String()}}
	}
	if status != UPDATE {
		//create
		_, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Create(context.Background(), deploy, metav1.CreateOptions{})
		return err
	} else {
		//update
		_, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Update(context.Background(), deploy, metav1.UpdateOptions{})
		return err
	}

}

func (t *TrinoContriller) InsertOrupdateWorkerDeploy(trino *v12.Trino, reference metav1.OwnerReference, status string, restart bool) error {
	label := GetLabelMap(trino.Name, "worker")
	name := GetWorkerName(trino.Name)
	runAsGroup := int64(1000)
	runAsUser := int64(1000)
	var replicas int32
	if trino.Spec.Pause {
		replicas = 0
	} else {
		replicas = int32(trino.Spec.WorkNum)

	}

	quantityCpu, err2 := resource.ParseQuantity(fmt.Sprintf("%d%s", trino.Spec.WorkCpu*1000, "m"))
	if err2 != nil {
		return err2
	}
	quantityMem, err2 := resource.ParseQuantity(fmt.Sprintf("%d", trino.Spec.CoordinatorMemory))

	var deploy *v13.Deployment
	baseEnv := []v1.EnvVar{}
	if status != UPDATE {
		//create
		deploy = &v13.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: trino.Namespace,
				Labels:    label,
			},
		}
		deploy.OwnerReferences = append(deploy.OwnerReferences, reference)
	} else {
		//update
		get, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err

		}
		deploy = get
		baseEnv = get.Spec.Template.Spec.Containers[0].Env
	}

	deploy.Spec = v13.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: label,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: label,
			},
			Spec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsUser:  &runAsUser,
					RunAsGroup: &runAsGroup,
				},
				Volumes: []v1.Volume{
					{
						Name: "config-volume",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: trino.Name + "-trino-worker",
								},
							},
						},
					},
					{
						Name: "catalog-volume",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: trino.Name + "-trino-catalog",
								},
							},
						},
					},
				},
				Containers: []v1.Container{
					{
						Name:  "trino-coordinator",
						Image: config.Image,
						Ports: []v1.ContainerPort{
							{
								Name:          "http",
								Protocol:      v1.Protocol("TCP"),
								ContainerPort: 8080,
							},
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "config-volume",
								MountPath: "/etc/trino",
							}, {
								Name:      "catalog-volume",
								MountPath: "/etc/trino/catalog",
							},
						},
						ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/v1/info",
									Port: intstr.Parse("http"),
								},
							},
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/v1/info",
									Port: intstr.Parse("http"),
								},
							},
							InitialDelaySeconds: 20,
						},
						Resources: v1.ResourceRequirements{
							Requests: map[v1.ResourceName]resource.Quantity{
								v1.ResourceCPU:    quantityCpu,
								v1.ResourceMemory: quantityMem,
							},
						},
						Env: baseEnv,
					},
				},
			},
		},
		Strategy:                v13.DeploymentStrategy{},
		MinReadySeconds:         0,
		RevisionHistoryLimit:    nil,
		Paused:                  false,
		ProgressDeadlineSeconds: nil,
	}
	if restart {
		deploy.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{{Name: "start_time", Value: time.Now().String()}}
	}

	if status != UPDATE {
		//create
		_, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Create(context.Background(), deploy, metav1.CreateOptions{})
		return err
	} else {
		//update
		_, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Update(context.Background(), deploy, metav1.UpdateOptions{})
		return err
	}
}

func (t *TrinoContriller) InsertService(trino *v12.Trino, reference metav1.OwnerReference, status string) error {
	if status == UPDATE {
		return nil
	}
	service := v1.Service{}
	service.SetName(trino.Name + "-trino")
	service.SetNamespace(trino.Namespace)
	service.Spec.Type = v1.ServiceTypeClusterIP
	port := v1.ServicePort{}
	port.Name = "http"
	port.Protocol = "TCP"
	port.Port = 8080
	port.TargetPort = intstr.Parse("http")
	service.Spec.Ports = append(service.Spec.Ports, port)
	service.Spec.Selector = make(map[string]string)
	service.Spec.Selector["app"] = trino.Name
	service.Spec.Selector["component"] = "coordinator"
	service.OwnerReferences = append(service.OwnerReferences, reference)
	_, err := t.KubeClient.CoreV1().Services(trino.Namespace).Create(context.Background(), &service, metav1.CreateOptions{})
	return err
}

func (t *TrinoContriller) InsertNodePort(trino *v12.Trino, reference metav1.OwnerReference, status string) error {
	if status == UPDATE {
		return nil
	}
	label := make(map[string]string)
	label["app"] = trino.Name
	label["component"] = "coordinator"
	service := v1.Service{}
	service.Labels = label
	service.Name = trino.Name + "-nodeservice"
	service.Namespace = trino.Namespace
	service.Spec.Type = v1.ServiceType("NodePort")

	service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
		Name:       "trino",
		Port:       8080,
		Protocol:   v1.ProtocolTCP,
		TargetPort: intstr.FromInt(8080),
	})
	service.Spec.Selector = make(map[string]string)
	service.Spec.Selector["app"] = trino.Name
	service.Spec.Selector["component"] = "coordinator"
	service.OwnerReferences = append(service.OwnerReferences, reference)

	_, err := t.KubeClient.CoreV1().Services(trino.Namespace).Create(context.Background(), &service, metav1.CreateOptions{})
	return err

}

func generatorMapHash(m map[string]string) []byte {
	var str string
	strIndex := []string{}
	for key, _ := range m {
		strIndex = append(strIndex, key)
	}
	sort.Strings(strIndex)
	for _, key := range strIndex {
		str += key
		str += m[key]
	}
	res := md5.Sum([]byte(str))
	return res[:]
}

func ToTrino(obj interface{}, status string) (*v12.Trino, error) {
	trino, ok := obj.(*v12.Trino)
	if !ok {
		klog.Info(status+" Trino error to process: ", obj)
		return nil, errors.New(status + " Trino error to process")
	}
	klog.Info(status, "  Trino: ", trino.GetName())
	return trino, nil
}

func GetWorkerName(name string) string {
	return name + "-worker"
}

func GetCoordinatorName(name string) string {
	return name + "-coordinator"
}

func GetLabelMap(name, typ string) map[string]string {
	label := make(map[string]string)
	label["app"] = name
	label["component"] = typ
	return label
}

func GetLabelString(name, typ string) string {
	return fmt.Sprintf("%s:%s,%s:%s", "app", name, "component", typ)
}
