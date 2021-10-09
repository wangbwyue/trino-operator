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

	//"trino-operator/pkg/apis/clientset/versioned"
	"trino-operator/pkg/common/config"
)

var (
	INSERT = "INSERT"
	UPDATE = "UPDATE"

	COORDINATOR = "coordinator"
	WORKER      = "worker"
)

type TrinoContriller struct {
	KubeClient  *kubernetes.Clientset
	TrinoClinet *versioned.Clientset
}

func (t *TrinoContriller) OnAdd(obj interface{}) {
	trino, err := t.ToTrino(obj, INSERT)
	if err != nil {
		return
	}
	t.InsertOrUpdate(trino, INSERT)

}

func (t *TrinoContriller) OnDelete(obj interface{}) {
}

func (t *TrinoContriller) OnUpdate(oldObj, newObj interface{}) {
	controller, ok := newObj.(*v12.Trino)
	if !ok {
		klog.Info("error to process Trino OnUpdate", newObj)
		return
	}
	klog.Info("OnUpdate  Trino: ", controller.GetName())
	trinoNew, err := t.ToTrino(newObj, "OnDelete")
	if err != nil {
		klog.Info("error to process Trino OnUpdate", err)
		return
	}
	trinoOld, err := t.ToTrino(oldObj, "OnDelete")
	if err != nil {
		klog.Info("error to process Trino OnUpdate", err)
		return
	}
	if reflect.DeepEqual(trinoNew.Spec, trinoOld.Spec) {
		klog.Info("Skip OnUpdate  Trino: version: ", trinoNew.ObjectMeta.ResourceVersion)
		return
	}
	t.InsertOrUpdate(trinoNew, UPDATE)
}

func (t *TrinoContriller) InsertOrUpdate(trino *v12.Trino, status string) {
	reference := t.GetReference(trino)

	err, needRestart := t.InsertOrUpdateCatalogConfig(trino, reference, status)
	klog.Infof("InsertOrUpdateCatalogConfig err: %v, need update %t", err, needRestart)

	err, needRestartCoordinator := t.InsertOrUpdateCoordinatorConfig(trino, reference, status)
	klog.Infof("InsertOrUpdateCoordinatorConfig err: %v, need update %t", err, needRestart)

	err, needRestartWorker := t.InsertOrUpdateWorkConfig(trino, reference, status)
	klog.Infof("InsertOrUpdateWorkConfig err: %v, need update %t", err, needRestart)

	err = t.InsertService(trino, reference, status)
	klog.Infof("InsertService err: %v", err)

	err = t.InsertOrUpdateDeploy(trino, reference, status, COORDINATOR, needRestart || needRestartCoordinator)
	klog.Infof("InsertCoordinatorDeploy err: %v", err)

	err = t.InsertOrUpdateDeploy(trino, reference, status, WORKER, needRestart || needRestartWorker)
	klog.Infof("InsertWorkerDeploy err: %v", err)

	err = t.InsertNodePort(trino, reference, status)
	klog.Infof("InsertNodePort err: %v", err)

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

	newHash := t.GeneratorMapHash(configMap.Data)
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
	newHash := t.GeneratorMapHash(configMap.Data)
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
	configMap.Data["node.properties"] = trino.Spec.WorkerConfig.NodeProperties
	configMap.Data["jvm.config"] = trino.Spec.WorkerConfig.JvmConfig
	configMap.Data["config.properties"] = trino.Spec.WorkerConfig.ConfigProperties
	configMap.Data["log.properties"] = trino.Spec.WorkerConfig.LogProperties

	newHash := t.GeneratorMapHash(configMap.Data)

	if status == UPDATE && bytes.Compare(newHash, configMap.BinaryData["hash"]) != 0 {
		needRestart = true
	}
	configMap.BinaryData["hash"] = newHash

	if status == UPDATE {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Update(context.Background(), configMap, metav1.UpdateOptions{})
		return err, needRestart
	} else {
		_, err := t.KubeClient.CoreV1().ConfigMaps(trino.Namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
		return err, needRestart
	}

}

func (t *TrinoContriller) InsertOrUpdateDeploy(trino *v12.Trino, reference metav1.OwnerReference, status, typ string, restart bool) error {

	label := t.GetLabelMap(trino, typ)
	name := GetName(trino, typ)
	runAsGroup, runAsUser := t.GetUser()
	replicas := t.GetReplicas(trino, typ)
	quantityCpu, quantityMem, err := t.GetQuantity(trino, typ)
	if err != nil {
		return err
	}

	spec := v13.DeploymentSpec{
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
									Name: t.GetConfigMapName(trino, typ),
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
						Name:  "trino",
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
						Env: []v1.EnvVar{},
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

	//update
	if status == UPDATE {
		deploy, err := t.KubeClient.AppsV1().Deployments(trino.Namespace).Get(context.Background(), name, metav1.GetOptions{})

		if err != nil {
			return err
		}
		baseEnv := deploy.Spec.Template.Spec.Containers[0].Env
		if restart {
			// set a new start_time env to restart
			deploy.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{{Name: "start_time", Value: time.Now().String()}}
		} else {
			deploy.Spec.Template.Spec.Containers[0].Env = baseEnv
		}
		deploy.Spec = spec
		_, err = t.KubeClient.AppsV1().Deployments(trino.Namespace).Update(context.Background(), deploy, metav1.UpdateOptions{})
		return err
	}

	//create
	deploy := &v13.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       trino.Namespace,
			Labels:          label,
			OwnerReferences: []metav1.OwnerReference{reference},
		},
		Spec: spec,
	}
	_, err2 := t.KubeClient.AppsV1().Deployments(trino.Namespace).Create(context.Background(), deploy, metav1.CreateOptions{})
	return err2

}

func (t *TrinoContriller) GetUser() (int64, int64) {
	runAsGroup := int64(1000)
	runAsUser := int64(1000)
	return runAsGroup, runAsUser
}

func (t *TrinoContriller) GetReplicas(trino *v12.Trino, typ string) int32 {
	replicas := int32(0)
	//set num
	if trino.Spec.Pause {
		return replicas
	}
	if typ == COORDINATOR {
		replicas = trino.Spec.CoordinatorConfig.Num
	}
	if typ == WORKER {
		replicas = trino.Spec.WorkerConfig.Num
	}
	return replicas
}

func (t *TrinoContriller) GetQuantity(trino *v12.Trino, typ string) (resource.Quantity, resource.Quantity, error) {
	//COORDINATOR
	if typ == COORDINATOR {
		quantityCpu, err2 := resource.ParseQuantity(fmt.Sprintf("%dm", trino.Spec.CoordinatorConfig.CpuRequest*1000))
		if err2 != nil {
			return resource.Quantity{}, resource.Quantity{}, err2
		}
		quantityMem, err3 := resource.ParseQuantity(fmt.Sprintf("%dm", trino.Spec.CoordinatorConfig.MemoryRequest))
		if err3 != nil {
			return resource.Quantity{}, resource.Quantity{}, err3
		}
		return quantityCpu, quantityMem, nil
	}
	//Worker
	quantityCpu, err2 := resource.ParseQuantity(fmt.Sprintf("%dm", trino.Spec.WorkerConfig.CpuRequest*1000))
	if err2 != nil {
		return resource.Quantity{}, resource.Quantity{}, err2
	}
	quantityMem, err3 := resource.ParseQuantity(fmt.Sprintf("%dm", trino.Spec.WorkerConfig.MemoryRequest))
	if err3 != nil {
		return resource.Quantity{}, resource.Quantity{}, err3
	}
	return quantityCpu, quantityMem, nil

}

func (t *TrinoContriller) InsertService(trino *v12.Trino, reference metav1.OwnerReference, status string) error {
	if status == UPDATE {
		return nil
	}
	service := &v1.Service{}
	service.SetName(trino.Name + "-trino")
	service.SetNamespace(trino.Namespace)
	service.Spec.Type = v1.ServiceTypeClusterIP
	service.Spec.Ports = []v1.ServicePort{
		{
			Name:       "http",
			Protocol:   "TCP",
			Port:       8080,
			TargetPort: intstr.Parse("http"),
		},
	}
	service.Spec.Selector = t.GetLabelMap(trino, COORDINATOR)
	service.OwnerReferences = []metav1.OwnerReference{reference}
	_, err := t.KubeClient.CoreV1().Services(trino.Namespace).Create(context.Background(), service, metav1.CreateOptions{})
	return err
}

func (t *TrinoContriller) InsertNodePort(trino *v12.Trino, reference metav1.OwnerReference, status string) error {
	if !trino.Spec.NodePort {
		return nil
	}
	label := t.GetLabelMap(trino, COORDINATOR)
	service := v1.Service{}
	service.Labels = label
	service.Name = trino.Name + "-nodeservice"
	service.Namespace = trino.Namespace
	service.Spec.Type = "NodePort"

	service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
		Name:       "trino",
		Port:       8080,
		Protocol:   v1.ProtocolTCP,
		TargetPort: intstr.FromInt(8080),
	})
	service.Spec.Selector = t.GetLabelMap(trino, COORDINATOR)
	service.OwnerReferences = append(service.OwnerReferences, reference)

	_, err := t.KubeClient.CoreV1().Services(trino.Namespace).Create(context.Background(), &service, metav1.CreateOptions{})
	return err

}

func (t *TrinoContriller) GetReference(trino *v12.Trino) metav1.OwnerReference {
	reference := metav1.OwnerReference{}
	reference.Name = trino.Name
	reference.Kind = v12.Kind
	reference.APIVersion = v12.GroupVersion.String()
	reference.UID = trino.UID
	return reference
}

func (t *TrinoContriller) GeneratorMapHash(m map[string]string) []byte {
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

func (t *TrinoContriller) ToTrino(obj interface{}, status string) (*v12.Trino, error) {
	trino, ok := obj.(*v12.Trino)
	if !ok {
		klog.Info(status+" Trino error to process: ", obj)
		return nil, errors.New(status + " Trino error to process")
	}
	klog.Info(status, "  Trino: ", trino.GetName())
	return trino, nil
}

func GetName(trino *v12.Trino, typ string) string {
	if typ == COORDINATOR {
		return trino.GetName() + "-coordinator"
	}
	return trino.GetName() + "-worker"
}

func (t *TrinoContriller) GetServiceName(trino v12.Trino) string {
	return trino.GetName() + "-trino"
}

func (t *TrinoContriller) GetLabelMap(trino *v12.Trino, typ string) map[string]string {
	label := make(map[string]string)
	label["app"] = trino.GetName()
	label["component"] = typ
	for name, value := range trino.Spec.Labels {
		// make sure app and component is right
		if _, ok := label[name]; !ok {
			label[name] = value
		}
	}
	return label
}

func (t *TrinoContriller) GetConfigMapName(trino *v12.Trino, typ string) string {
	return trino.Name + "-trino-" + typ
}
