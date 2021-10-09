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

	// catalog config
	CataLogConfig map[string]string `json:"cataLogConfig"`

	// coordiator config
	CoordinatorConfig WorkloadConfig `json:"coordinatorConfig"`

	// work config
	WorkerConfig WorkloadConfig `json:"workerConfig"`

	// Pause
	Pause bool `json:"pause"`

	// labels
	Labels map[string]string `json:"labels,omitempty"`

	// annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	//nodeport
	// +kubebuilder:default=true
	NodePort bool `json:"nodePort"`
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

	// number for Workload , coordinator is 1
	// +kubebuilder:default=1
	Num int32 `json:"num,omitempty"`
	// per work cpu
	// +kubebuilder:default=1000
	CpuRequest int `json:"cpuRequest,omitempty"`
	//per work memory
	// +kubebuilder:default=2048
	MemoryRequest int `json:"memoryRequest,omitempty"`
}

// TrinoStatus defines the observed state of Trino
type TrinoStatus struct {

	// status for all trino cluster
	// 	STOPPED  when trino.spec.pause is true
	//	RUNNING      when trino.spec.pause is false and all workload is running
	//	TRANSITIONING  when trino.spec.pause is false and workload is not ready
	Status string `json:"status"`

	// TotalCpu size m
	TotalCpu int64 `json:"totalCpu"`

	// size m
	// total memory vaule for all workload
	TotalMemory int64 `json:"totalMemory"`

	// Coordinator pod status
	// when this pod is running, you can connect to trino
	CoordinatorPod []PodStatus `json:"coordinatorPod"`

	// Worker pod status
	WorkerPod []PodStatus `json:"workerPod"`
}

type PodStatus struct {
	Name      string `json:"name"`
	Cpu       string `json:"cpu"`
	Memory    string `json:"memory"`
	PodStatus string `json:"podStatus"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,path=trinos
// +kubebuilder:printcolumn:name="totalCpu",type=integer,JSONPath=`.status.totalCpu`
// +kubebuilder:printcolumn:name="totalMemory",type=integer,JSONPath=`.status.totalMemory`
// +kubebuilder:printcolumn:name="coordinatorNum",type=integer,JSONPath=`.spec.coordinatorConfig.num`
// +kubebuilder:printcolumn:name="workerNum",type=integer,JSONPath=`.spec.workerConfig.num`
// +kubebuilder:printcolumn:name="Pause",type=boolean,JSONPath=`.spec.pause`
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Trino is the Schema for the trinos API
type Trino struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrinoSpec   `json:"spec,omitempty"`
	Status TrinoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TrinoList contains a list of Trino
type TrinoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trino `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Trino{}, &TrinoList{})
}
