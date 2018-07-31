package kubernetes

import (
	jenkinsio "github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegisterSSOCRD ensures that the CRD is registered for SSO
func RegisterSSOCRD(apiClient apiextensionsclientset.Interface) error {
	name := "ssos." + jenkinsio.GroupName
	names := &v1beta1.CustomResourceDefinitionNames{
		Kind:       "SSO",
		ListKind:   "SSOList",
		Plural:     "ssos",
		Singular:   "sso",
		ShortNames: []string{"sso"},
	}

	return registerCRD(apiClient, name, names)
}

func registerCRD(apiClient apiextensionsclientset.Interface, name string, names *v1beta1.CustomResourceDefinitionNames) error {
	_, err := apiClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	crd := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   jenkinsio.GroupName,
			Version: jenkinsio.Version,
			Scope:   v1beta1.NamespaceScoped,
			Names:   *names,
		},
	}
	_, err = apiClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	return err
}
