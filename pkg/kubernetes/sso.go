package kubernetes

import (
	v1 "github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsSSOInitialized checks if the SSO is initialized by retrieving the current state from k8s
func IsSSOInitialized(sso *v1.SSO) (bool, error) {
	client, err := GetJenkinsClient()
	if err != nil {
		return false, errors.Wrap(err, "getting Jenkins client")
	}

	ssos, err := client.JenkinsV1().SSOs(sso.GetNamespace()).List(metav1.ListOptions{})
	if err != nil {
		return false, errors.Wrap(err, "listing SSO resources")
	}

	for _, s := range ssos.Items {
		if s.GetName() == sso.GetName() {
			return s.Status.Initialized, nil
		}
	}

	return false, nil
}
