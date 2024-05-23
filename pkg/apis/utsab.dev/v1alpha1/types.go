package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterID",type=string,JSONPath=`.status.klusterID`
// +kubebuilder:printcolumn:name="Progress",type=string,JSONPath=`.status.progress`

// Kluster should contain typemeta, objectmeta and the spec
type Kluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KlusterSpec   `json:"spec,omitempty"`
	Status KlusterStatus `json:"status,omitempty"`
}

type KlusterStatus struct {
	KlusterID  string `json:"klusterID,omitempty"`
	Progress   string `json:"progress,omitempty"`
	KubeConfig string `json:"kubeConfig,omitempty"`
}

// this specifies all the fields required as input to your operator
// for the digitalocean to create k8s cluster, required fields must be specified as per documentation.
type KlusterSpec struct {
	Name        string `json:"name,omitempty"`
	Region      string `json:"region,omitempty"`
	Version     string `json:"version,omitempty"`
	TokenSecret string `json:"tokenSecret,omitempty"`

	NodePools []NodePool `json:"nodePools,omitempty"`
}

// since NodePools is slice of NodePool, we need to specify the required section for the NodePool struct.
type NodePool struct {
	Size  string `json:"size,omitempty"`
	Name  string `json:"name,omitempty"`
	Count int    `json:"count,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// To list the Kluster
// same as get pods which list all the pods
// this lists all the Kluster

// This should also be registered as addknownTypes in register.go file
type KlusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Kluster `json:"items,omitempty"`
}

// func (kl *KlusterList) DeepCopyObject() runtime.Object {
//     if kl == nil {
//         return nil
//     }
//     out := new(KlusterList)
//     kl.DeepCopyInto(out)
//     return out
// }
