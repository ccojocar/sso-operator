package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	subjectKind        = "ServiceAccount"
	roleKind           = "ClusterRole"
	serviceAccountName = "sso-operator-sa"
)

// EnsureClusterRoleBinding ensures for the given cluster role name that there is a binding to a service account
// in the given namespace
func EnsureClusterRoleBinding(clusterRoleName string, namespace string) (string, error) {
	k8sClient, err := GetClientset()
	if err != nil {
		return "", errors.Wrap(err, "getting k8s client")
	}

	roleBindingList, err := k8sClient.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
	for _, roleBinding := range roleBindingList.Items {
		roleRef := roleBinding.RoleRef
		if roleRef.Kind == roleKind && roleRef.Name == clusterRoleName {
			// Check if there is already a service account assigned to the operator cluster role in the given namespace
			for _, subj := range roleBinding.Subjects {
				if subj.Kind == subjectKind && subj.Namespace == namespace {
					return subj.Name, nil
				}
			}

			saName, err := CreateServiceAccount(serviceAccountName, namespace)
			if err != nil {
				return "", errors.Wrap(err, "role binding creating service account")
			}

			subj := rbacv1.Subject{
				Kind:      subjectKind,
				Name:      saName,
				Namespace: namespace,
			}
			roleBinding.Subjects = append(roleBinding.Subjects, subj)

			_, err = k8sClient.RbacV1().ClusterRoleBindings().Update(&roleBinding)
			if err != nil {
				return "", errors.Wrapf(err, "adding service account '%s' to cluster role binding '%s'", saName, roleBinding.Name)
			}
			return saName, nil
		}
	}

	return "", fmt.Errorf("no cluster role binding found for cluster role '%s', make sure a binding existing when deploying the operator", clusterRoleName)
}

// CreateServiceAccount creates a new service account in the given namespace and returns the service account name
func CreateServiceAccount(name string, namespace string) (string, error) {
	k8sClient, err := GetClientset()
	if err != nil {
		return "", errors.Wrap(err, "getting k8s client")
	}

	sa, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
	// If a service account already exists just re-use it
	if err == nil {
		return sa.Name, nil
	}

	sa = &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       subjectKind,
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	sa, err = k8sClient.CoreV1().ServiceAccounts(namespace).Create(sa)
	if err != nil {
		return "", errors.Wrapf(err, "creating service account '%s'", sa)
	}

	return sa.Name, nil
}
