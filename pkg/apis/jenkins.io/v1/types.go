package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// SSO represent Single Sign-On required to create a OIDC client in dex
type SSO struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   SSOSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status SSOStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`
}

// SSOSpec is the specification of a Single Sing-On resource
type SSOSpec struct {
	ID string `json:"id,omitempty"`

	SecretName   string   `json:"secret,omitempty"`
	RedirectURIs []string `json:"redirectURIs,omitempty"`
	TrustedPeers []string `json:"trustedPeers,omitempty"`

	Public bool `json:"public"`

	Name    string `json:"name,omitempty"`
	LogoURL string `json:"logoURL,omitempty"`
}

// SSOStatus is the status of an Single Sign-On resource
type SSOStatus struct {
	Status string `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SSOList represents a list of Single Sign-On Kubernetes objects
type SSOList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SSO `json:"items"`
}
