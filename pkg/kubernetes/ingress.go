package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FindIngressHosts searches an ingress resource by name and retrieves its hosts
func FindIngressHosts(name string, namespace string) ([]string, error) {
	k8sClient, err := GetClientset()
	if err != nil {
		return nil, errors.Wrap(err, "getting k8s client")
	}

	ingresses, err := k8sClient.Extensions().Ingresses(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "listing ingress resources")
	}

	for _, ingress := range ingresses.Items {
		if ingress.GetName() == name {
			hosts := []string{}
			rules := ingress.Spec.Rules
			for _, rule := range rules {
				hosts = append(hosts, rule.Host)
			}
			return hosts, nil
		}
	}
	return nil, fmt.Errorf("ingress '%s' not found", name)
}
