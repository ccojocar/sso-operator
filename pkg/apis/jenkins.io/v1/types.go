package v1

import (
	"k8s.io/api/core/v1"
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
	// OIDCIssuerURL URL of dex IdP
	OIDCIssuerURL string `json:"oidcIssuerUrl,omitempty"`
	// Name of the upstream service for which the SSO is created
	UpstreamService string `json:"upstreamService,omitempty"`
	// Domain name under which the SSO service will be exposed
	Domain string `json:"domain,omitempty"`
	// cert-manager issuer name
	CertIssuerName string `json:"certIssuerName,omitempty"`
	// Docker image for oauth2_proxy
	ProxyImage string `json:"proxyImage,omitempty"`
	// Docker image tag for oauth2_proxy
	ProxyImageTag string `json:"proxyImageTag,omitempty"`
	// Resource requirements for oauth2_proxy pod
	ProxyResources v1.ResourceRequirements `json:"proxyResources,omitempty"`
	// Indicate if the access token should be forwarded to the upstream service
	ForwardToken bool `json:"forwardToken,omitempty"`
	// CookieSpec cookie specifications
	CookieSpec CookieSpec `json:"cookieSpec,omitempty"`
}

// CookieSpec is the specification of a cookie for a Single Sign-On resource
type CookieSpec struct {
	// Cookie name
	Name string `json:"name,omitempty"`
	// Expiration time of the cookie
	Expire string `json:"expire,omitempty"`
	// Refresh time of the cookie
	Refresh string `json:"refresh,omitempty"`
	// Cookie is only send over a HTTPS connection
	Secure bool `json:"secure,omitempty"`
	// Cookie is not readable from JavaScript
	HTTPOnly bool `json:"httpOnly,omitempty"`
}

// SSOStatus is the status of an Single Sign-On resource
type SSOStatus struct {
	// OIDC client ID created in dex
	ClientID string `json:"clientId,omitempty" protobuf:"bytes,2,opt,name=clientId"`
	// Initialized indicated if the SSO was configured in dex and oauth2_proxy
	Initialized bool `json:"initialized,omitempty" protobuf:"bytes,2,opt,name=initialized"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SSOList represents a list of Single Sign-On Kubernetes objects
type SSOList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SSO `json:"items"`
}
