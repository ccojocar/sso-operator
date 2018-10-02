package proxy

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/dex/api"
	apiv1 "github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/sso-operator/pkg/kubernetes"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	configPath          = "/config/oauth2_proxy.cfg"
	configVolumeName    = "proxy-config"
	configSecretName    = "proxy-secret" // #nosec
	secretVersionEnv    = "SECRET_VERSION"
	portName            = "proxy-port"
	port                = 4180
	healthPath          = "/ping"
	replicas            = 1
	publicPort          = 80
	cookieSecretLen     = 32
	fakeURL             = "https://fake-oauth2-proxy"
	createTimeout       = time.Duration(60 * time.Second)
	createIntervalCheck = time.Duration(10 * time.Second)
	readyTimeout        = time.Duration(5 * time.Minute)
	appLabel            = "app"
	releaseLabel        = "release"

	exposeAnnotation        = "fabric8.io/expose"
	exposeIngressAnnotation = "fabric8.io/ingress.annotations"
	ingressNameAnnotation   = "fabric8.io/ingress.name"
	ingressClassAnnotations = "kubernetes.io/ingress.class"
	certManagerAnnotation   = "certmanager.k8s.io/issuer"
	ingressClass            = "nginx"
)

// Proxy keeps the k8s resources created for a proxy
type Proxy struct {
	AppName    string
	Secret     *v1.Secret
	Deployment *appsv1.Deployment
	Service    *v1.Service
}

// FakeRedirectURL builds a fake redirect URL for oauth2 proxy
func FakeRedirectURL() string {
	return RedirectURL(fakeURL)
}

// RedirectURL build the redirect URL for oauth2 proxy
func RedirectURL(URL string) string {
	return fmt.Sprintf("%s/oauth2/callback", URL)
}

// ConvertHostsToRedirectURLs converts a list of host to proxy redirect URLs
func ConvertHostsToRedirectURLs(hosts []string, sso *apiv1.SSO) []string {
	redirectURLs := []string{}
	for _, host := range hosts {
		redirectURL := RedirectURL(fmt.Sprintf("https://%s", host))
		redirectURLs = append(redirectURLs, redirectURL)
	}
	return redirectURLs
}

func buildName(name string, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, name)
}

func labels(sso *apiv1.SSO, appName string) map[string]string {
	return map[string]string{"app": appName, "sso": sso.GetName()}
}

func serviceAnnotations(sso *apiv1.SSO, appName string) map[string]string {
	return map[string]string{
		exposeAnnotation:        "true",
		ingressNameAnnotation:   appName,
		exposeIngressAnnotation: ingressClassAnnotations + ": " + ingressClass + "\n" + certManagerAnnotation + ": " + sso.Spec.CertIssuerName,
	}
}

// Deploy deploys the oauth2 proxy
func Deploy(sso *apiv1.SSO, oidcClient *api.Client, cookieSecret string) (*Proxy, error) {
	appName, err := getAppName(sso.Spec.UpstreamService, sso.GetNamespace())
	if err != nil {
		return nil, errors.Wrap(err, "gettting the app name from upstream service labels")
	}
	secret, err := proxySecret(sso, oidcClient, cookieSecret, labels(sso, appName))
	if err != nil {
		return nil, errors.Wrap(err, "creating oauth2_proxy config")
	}
	secret.SetOwnerReferences(append(secret.GetOwnerReferences(), ownerRef(sso)))
	err = sdk.Create(secret)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "creating oauth2_proxy secret")
	}

	ns := sso.GetNamespace()
	secretVersion := computeSecretVersion(secret)
	podTempl := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName(sso.GetName(), sso.GetNamespace()),
			Namespace: ns,
			Labels:    labels(sso, appName),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{proxyContainer(sso, secretVersion)},
			Volumes: []v1.Volume{{
				Name: configVolumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: configSecretName,
					},
				},
			}},
		},
	}

	deployment := buildName(sso.GetName(), sso.GetNamespace())
	var replicas int32 = replicas
	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment,
			Namespace: ns,
			Labels:    labels(sso, appName),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels(sso, appName)},
			Template: podTempl,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				},
			},
		},
	}

	d.SetOwnerReferences(append(d.GetOwnerReferences(), ownerRef(sso)))

	err = sdk.Create(d)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "creating oauth2_proxy deployment")
	}

	service := sso.GetName()
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        service,
			Namespace:   ns,
			Labels:      labels(sso, appName),
			Annotations: serviceAnnotations(sso, appName),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:       portName,
				Protocol:   v1.ProtocolTCP,
				Port:       publicPort,
				TargetPort: intstr.FromInt(port),
			}},
			Selector: labels(sso, appName),
		},
	}

	svc.SetOwnerReferences(append(svc.GetOwnerReferences(), ownerRef(sso)))

	err = sdk.Create(svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "creating oauth2_proxy service")
	}

	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return nil, errors.Wrap(err, "getting k8s client")
	}

	exists := true
	err = kubernetes.WaitForService(k8sClient, ns, service, exists, createIntervalCheck, createTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "wait for service")
	}

	label := k8slabels.SelectorFromSet(k8slabels.Set(map[string]string{"sso": sso.GetName()}))
	err = kubernetes.WaitForPodsWithLabelRunning(k8sClient, ns, label, readyTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "waiting for SSO proxy")
	}

	return &Proxy{
		AppName:    appName,
		Secret:     secret,
		Deployment: d,
		Service:    svc,
	}, nil
}

// Update updates the oauth2_proxy secret and deployment
func Update(proxy *Proxy, sso *apiv1.SSO, client *api.Client, cookieSecret string) error {
	err := updateProxySecret(proxy.Secret, sso, client, cookieSecret)
	if err != nil {
		return errors.Wrap(err, "updating oauth2_proxy secret")
	}

	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	namespace := sso.GetNamespace()
	deploymentList, err := k8sClient.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "listing deployments")
	}
	secretVersion := computeSecretVersion(proxy.Secret)
	deploymentName := proxy.Deployment.GetName()
	for _, deployment := range deploymentList.Items {
		if deployment.GetName() == deploymentName {
			containers := deployment.Spec.Template.Spec.Containers
			for _, container := range containers {
				updateContainer(&container, secretVersion)
			}
			deployment.TypeMeta = metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			}
			err = sdk.Update(&deployment)
			if err != nil {
				return errors.Wrap(err, "updating oauth2_proxy deployment")
			}
		}
	}

	label := k8slabels.SelectorFromSet(k8slabels.Set(map[string]string{"sso": sso.GetName()}))
	err = kubernetes.WaitForPodsWithLabelRunning(k8sClient, sso.GetNamespace(), label, readyTimeout)
	if err != nil {
		return errors.Wrap(err, "waiting for SSO proxy")
	}
	return nil
}

func updateContainer(container *v1.Container, secretVersion string) {
	for i, env := range container.Env {
		if env.Name == secretVersionEnv {
			container.Env[i].Value = secretVersion
		}
	}
}

func computeSecretVersion(secret *v1.Secret) string {
	secretData := ""
	for k, v := range secret.StringData {
		secretData += k + v
	}
	hash := sha256.Sum256([]byte(secretData))
	secretVersion := base64.URLEncoding.EncodeToString(hash[:])
	return secretVersion
}

func ownerRef(sso *apiv1.SSO) metav1.OwnerReference {
	controller := true
	return metav1.OwnerReference{
		APIVersion: apiv1.SchemeGroupVersion.String(),
		Kind:       apiv1.SSOKind,
		Name:       sso.Name,
		UID:        sso.UID,
		Controller: &controller,
	}
}

func proxyContainer(sso *apiv1.SSO, secretVersion string) v1.Container {
	return v1.Container{
		Name:            sso.GetName(),
		Image:           fmt.Sprintf("%s:%s", sso.Spec.ProxyImage, sso.Spec.ProxyImageTag),
		ImagePullPolicy: v1.PullIfNotPresent,
		Args:            []string{fmt.Sprintf("--config=%s", configPath)},
		Ports: []v1.ContainerPort{{
			Name:          portName,
			ContainerPort: int32(port),
			Protocol:      v1.ProtocolTCP,
		}},
		Resources: sso.Spec.ProxyResources,
		VolumeMounts: []v1.VolumeMount{{
			Name:      configVolumeName,
			ReadOnly:  true,
			MountPath: filepath.Dir(configPath),
		}},
		Env: []v1.EnvVar{{
			Name:  secretVersionEnv,
			Value: secretVersion,
		}},
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   healthPath,
					Port:   intstr.FromInt(port),
					Scheme: v1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 60,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			FailureThreshold:    3,
		},
		ReadinessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   healthPath,
					Port:   intstr.FromInt(port),
					Scheme: v1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 30,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			FailureThreshold:    3,
		},
	}
}

func proxyConfig(sso *apiv1.SSO, client *api.Client, cookieSecret string) (string, error) {
	upstreamURL, err := getUpstreamURL(sso.Spec.UpstreamService, sso.Namespace)
	if err != nil {
		return "", errors.Wrap(err, "getting the upstream service URL")
	}
	redirectURLs := client.RedirectUris
	if len(redirectURLs) == 0 {
		return "", errors.New("no redirect URL provided")
	}
	issuerURL := sso.Spec.OIDCIssuerURL
	if !strings.HasPrefix(issuerURL, "https://") {
		return "", errors.New("issuer URL must used HTTPS")
	}
	c := &Config{
		Port:          port,
		ClientID:      client.GetId(),
		ClientSecret:  client.GetSecret(),
		OIDCIssuerURL: sso.Spec.OIDCIssuerURL,
		RedirectURL:   redirectURLs[0],
		LoginURL:      fmt.Sprintf("%s/auth", issuerURL),
		RedeemURL:     fmt.Sprintf("%s/token", issuerURL),
		Upstream:      upstreamURL,
		ForwardToken:  sso.Spec.ForwardToken,
		Cookie: Cookie{
			Name:     sso.Spec.CookieSpec.Name,
			Secret:   cookieSecret,
			Domain:   sso.Spec.Domain,
			Expire:   sso.Spec.CookieSpec.Expire,
			Refresh:  sso.Spec.CookieSpec.Refresh,
			Secure:   sso.Spec.CookieSpec.Secure,
			HTTPOnly: sso.Spec.CookieSpec.HTTPOnly,
		},
	}

	config, err := renderConfig(c)
	if err != nil {
		return "", errors.Wrap(err, "rendering oauth2_proxy config")
	}
	return config, nil
}

func updateProxySecret(secret *v1.Secret, sso *apiv1.SSO, client *api.Client, cookieSecret string) error {
	config, err := proxyConfig(sso, client, cookieSecret)
	if err != nil {
		return errors.Wrap(err, "creating oauth2_proxy config")
	}

	secret.StringData[filepath.Base(configPath)] = config

	err = sdk.Update(secret)
	if err != nil {
		return errors.Wrap(err, "updating oauth2_proxy secret")
	}
	return nil
}

func proxySecret(sso *apiv1.SSO, client *api.Client, cookieSecret string, labels map[string]string) (*v1.Secret, error) {
	config, err := proxyConfig(sso, client, cookieSecret)
	if err != nil {
		return nil, errors.Wrap(err, "creating oauth2_proxy config")
	}
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configSecretName,
			Namespace: sso.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			filepath.Base(configPath): config,
		},
		Type: v1.SecretTypeOpaque,
	}
	return secret, nil
}

// GenerateCookieKey generates a random key which used to sign the SSO cookie
func GenerateCookieKey() (string, error) {
	return generateSecret(cookieSecretLen)
}

func generateSecret(size int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes, err := generateRandomBytes(size)
	if err != nil {
		return "", errors.Wrap(err, "generating secret")
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

func generateRandomBytes(len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "generating random")
	}
	return b, nil
}

func getUpstreamURL(upstreamService string, namespace string) (string, error) {
	kubeClient, err := kubernetes.GetClientset()
	if err != nil {
		return "", errors.Wrap(err, "creating k8s client")
	}

	serviceList, err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "listing services in namespace '%s'", namespace)
	}
	for _, service := range serviceList.Items {
		if service.GetName() == upstreamService {
			port := service.Spec.Ports[0].Port
			return fmt.Sprintf("http://%s:%d", service.Name, port), nil
		}
	}
	return "", fmt.Errorf("no service '%s' found in namespace '%s'", upstreamService, namespace)
}

func getAppName(upstreamService string, namespace string) (string, error) {
	kubeClient, err := kubernetes.GetClientset()
	if err != nil {
		return "", errors.Wrap(err, "creating k8s client")
	}
	serviceList, err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "listing services in namespace '%s'", namespace)
	}
	for _, service := range serviceList.Items {
		if service.GetName() == upstreamService {
			labels := service.ObjectMeta.GetLabels()
			appLabelValue := ""
			releaseLabelValue := ""
			for name, value := range labels {
				if name == appLabel {
					appLabelValue = value
					break
				}
				if name == releaseLabel {
					releaseLabelValue = value
				}
			}
			if appLabelValue != "" {
				return appLabelValue, nil
			}
			if releaseLabelValue != "" {
				return strings.Replace(service.Name, releaseLabelValue+"-", "", 1), nil
			}
			return service.Name, nil
		}
	}
	return "", fmt.Errorf("no service '%s' found in namespace '%s'", upstreamService, namespace)
}
