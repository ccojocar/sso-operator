package operator

import (
	"github.com/jenkins-x/sso-operator/pkg/kubernetes"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
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
		return nil, errors.Wrap(err, "getting the operator secret")
	}

	cookieKey, ok := secret.StringData[ssoCookieKey]
	if !ok {
		return nil, errors.New("sso cookie key not found in operator secret")
	}

	return &operatorConfig{
		ssoCookieKey: cookieKey,
	}, nil
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
