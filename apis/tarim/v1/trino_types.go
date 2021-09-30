/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TrinoSpec defines the desired state of Trino
type TrinoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//catalog config
	CataLogConfig map[string]string `json:"cataLogConfig"`

	//coordiator config
	CoordinatorConfig WorkloadConfig `json:"coordinatorConfig"`

	//work config
	WorkConfig WorkloadConfig `json:"workConfig"`

	//projectSpeaceId
	ProjectSpeaceId int `json:"projectSpeaceId"`

	//Pause
	Pause bool `json:"pause"`

	//metastore
	Metastore string `json:"metastore,omitempty"`

	//TenantId
	TenantId string `json:"tenantId,omitempty"`

	CpuTotal int64 `json:"cpuTotal,omitempty"`

	MemoryTotal int64 `json:"memoryTotal,omitempty"`

	CreateTime metav1.Time `json:"createTime,omitempty"`

	//Coordinator cpu
	CoordinatorCpu int `json:"coordinatorCpu"`
	//Coordinator mem
	CoordinatorMemory int `json:"coordinatorMemory"`

	//WorkNum
	WorkNum int `json:"workNum"`
	// per work cpu
	WorkCpu int `json:"workCpu"`
	//per work memory
	WorkMemory int `json:"workMemory"`
}

type WorkloadConfig struct {
	//node prpperties
	NodeProperties string `json:"nodeProperties"`

	//jvm config
	JvmConfig string `json:"jvmConfig"`

	//config properties
	ConfigProperties string `json:"configProperties"`

	//log properties
	LogProperties string `json:"logProperties"`
}

// TrinoStatus defines the observed state of Trino
type TrinoStatus struct {
	//status
	Status string `json:"status"`

	//Coordinator
	CoordinatorPod []PodStatus `json:"coordinatorPod"`

	//Worker
	WorkerPod []PodStatus `json:"workerPod"`
}

type PodStatus struct {
	Name      string `json:"name"`
	Cpu       string `json:"cpu"`
	Memory    string `json:"memory"`
	PodStatus string `json:"podStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,path=trinos
// +genclient

// Trino is the Schema for the trinoes API
type Trino struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrinoSpec   `json:"spec,omitempty"`
	Status TrinoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TrinoList contains a list of Trino
type TrinoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trino `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Trino{}, &TrinoList{})
}
