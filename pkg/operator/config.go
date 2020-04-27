package operator

import (
	"github.com/jenkins-x/sso-operator/pkg/kubernetes"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	operatorSecretName = "operator-secret"
	ssoCookieKey       = "ssoCookieKey"
)

type operatorConfig struct {
	ssoCookieKey string
}

func getOperatorConfigFromSecret(namespace string) (*operatorConfig, error) {
	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return nil, errors.Wrap(err, "getting k8s client")
	}

	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(operatorSecretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.New("operator config secret not found")
		}
		delerr := deleteOperatorConfigSecret(namespace)
		if delerr != nil {
			return nil, errors.Wrap(delerr, "cleaning up the operator config secert")
		}
		return nil, errors.Wrap(err, "cleaning up the operator config secret due to error")
	}

	cookieKey, ok := secret.Data[ssoCookieKey]
	if !ok {
		delerr := deleteOperatorConfigSecret(namespace)
		if delerr != nil {
			return nil, errors.Wrap(delerr, "cleaning up the operator config secert because cookie key is missing")
		}
		return nil, errors.New("sso cookie key not found in operator secret")
	}

	return &operatorConfig{
		ssoCookieKey: string(cookieKey),
	}, nil
}

func deleteOperatorConfigSecret(namespace string) error {
	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}
	return k8sClient.CoreV1().Secrets(namespace).Delete(operatorSecretName, &metav1.DeleteOptions{})
}

func storeOperatorConfigInSecret(namespace string, config *operatorConfig) error {
	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			ssoCookieKey: config.ssoCookieKey,
		},
		Type: v1.SecretTypeOpaque,
	}

	_, err = k8sClient.CoreV1().Secrets(namespace).Create(secret)
	return err
}
