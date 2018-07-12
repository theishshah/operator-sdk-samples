package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IpfsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Ipfs `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Ipfs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              IpfsSpec   `json:"spec"`
	Status            IpfsStatus `json:"status,omitempty"`
}

type IpfsSpec struct {
	Size int32 `json:"size"`
}
type IpfsStatus struct {
	Nodes []string `json:"nodes"`
}
