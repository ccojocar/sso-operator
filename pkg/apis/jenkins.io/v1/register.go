package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	jenkinsio "github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io"
	sdkK8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// SchemeBuilder for building the schema
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme helper
	AddToScheme = SchemeBuilder.AddToScheme

	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: jenkinsio.GroupName, Version: jenkinsio.Version}
)

func init() {
	sdkK8sutil.AddToSDKScheme(AddToScheme)
}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSO{},
		&SSOList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
