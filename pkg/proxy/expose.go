package proxy

import (
	"fmt"
	"path/filepath"
	"time"

	apiv1 "github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/sso-operator/pkg/kubernetes"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	exposeImage            = "jenkinsxio/exposecontroller"
	exposeImageTag         = "latest"
	exposeCmd              = "/exposecontroller"
	exposeConfigPath       = "/etc/exposecontoller/config.yml"
	exposeConfigVolumeName = "expose-config"
	exposeConfigMapName    = "expose-configmap"
	exposeEnv              = "KUBERNETES_NAMESPACE"
	exposer                = "Ingress"
	exposeTimeout          = time.Duration(5 * time.Minute)
	exposeCheckInterval    = time.Duration(10 * time.Second)
	cleanupeTimeout        = time.Duration(2 * time.Minute)
	cleanupCheckInterval   = time.Duration(10 * time.Second)
)

// Expose executes the exposecontroller as a Job in order publicly expose the SSO service
func Expose(sso *apiv1.SSO, serviceName string, serviceAccount string) error {
	configMap, err := exposeConfigMap(sso, serviceName)
	if err != nil {
		return errors.Wrap(err, "building expose config map")
	}
	configMap.SetOwnerReferences(append(configMap.GetOwnerReferences(), ownerRef(sso)))
	err = sdk.Create(configMap)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "creating expose config map")
	}

	job := createJob("expose", sso, serviceAccount, exposeContainer(sso))
	job.SetOwnerReferences(append(job.GetOwnerReferences(), ownerRef(sso)))
	err = sdk.Create(job)
	if err != nil {
		msg := "creating expose job"
		if apierrors.IsAlreadyExists(err) {
			errdel := sdk.Delete(job)
			if errdel != nil {
				return errors.Wrapf(errdel, "%s: deleting existing expose job", msg)
			}
		}
		return errors.Wrap(err, msg)
	}

	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}
	err = kubernetes.WaitForJobComplete(k8sClient, sso.GetNamespace(), job.GetName(), exposeCheckInterval, exposeTimeout)
	if err != nil {
		return errors.Wrap(err, "waiting for SSO to be exposed")
	}

	deletePropagation := metav1.DeletePropagationBackground
	deleteOption := &metav1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	}
	err = sdk.Delete(job, sdk.WithDeleteOptions(deleteOption))
	if err != nil {
		return errors.Wrap(err, "cleaning up the expose job")
	}

	err = sdk.Delete(configMap)
	if err != nil {
		return errors.Wrap(err, "cleaning up the expose config map")
	}

	return nil
}

// Cleanup executes the exposecontroller as a job to cleanup the ingress resources
func Cleanup(sso *apiv1.SSO, serviceName string, serviceAccount string) error {
	configMap, err := exposeConfigMap(sso, serviceName)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "building cleanup config map")
	}
	err = sdk.Create(configMap)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "creating cleanup config map")
	}

	job := createJob("cleanup", sso, serviceAccount, cleanupContainer(sso, serviceName))
	err = sdk.Create(job)
	if err != nil {
		msg := "creating cleanup job"
		if apierrors.IsAlreadyExists(err) {
			errdel := sdk.Delete(job)
			if errdel != nil {
				return errors.Wrapf(errdel, "%s: deleting existing cleanup job", msg)
			}
		}
		return errors.Wrap(err, msg)
	}

	k8sClient, err := kubernetes.GetClientset()
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}
	err = kubernetes.WaitForJobComplete(k8sClient, sso.GetNamespace(), job.GetName(), cleanupCheckInterval, cleanupeTimeout)
	if err != nil {
		return errors.Wrap(err, "waiting for SSO to be cleaned up")
	}

	deletePropagation := metav1.DeletePropagationBackground
	deleteOption := &metav1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	}
	err = sdk.Delete(job, sdk.WithDeleteOptions(deleteOption))
	if err != nil {
		return errors.Wrap(err, "deleting the cleanup job")
	}

	err = sdk.Delete(configMap)
	if err != nil {
		return errors.Wrap(err, "deleting the cleanup config map")
	}

	return nil
}

func createJob(name string, sso *apiv1.SSO, serviceAccount string, container *v1.Container) *batchv1.Job {
	ns := sso.GetNamespace()
	name = fmt.Sprintf("%s-%s", buildName(sso.GetName(), ns), name)

	podTempl := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: serviceAccount,
			Containers:         []v1.Container{*container},
			Volumes: []v1.Volume{{
				Name: exposeConfigVolumeName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: exposeConfigMapName,
						},
					},
				},
			}},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: batchv1.JobSpec{
			Template: podTempl,
		},
	}
}

func exposeContainer(sso *apiv1.SSO) *v1.Container {
	return &v1.Container{
		Name:            fmt.Sprintf("%s-expose", sso.GetName()),
		Image:           fmt.Sprintf("%s:%s", exposeImage, exposeImageTag),
		ImagePullPolicy: v1.PullIfNotPresent,
		Command:         []string{exposeCmd},
		Args:            []string{fmt.Sprintf("--config=%s", exposeConfigPath), "--v", "4"},
		VolumeMounts: []v1.VolumeMount{{
			Name:      exposeConfigVolumeName,
			ReadOnly:  true,
			MountPath: filepath.Dir(exposeConfigPath),
		}},
		Env: []v1.EnvVar{{
			Name:  exposeEnv,
			Value: sso.GetNamespace(),
		}},
	}
}

func cleanupContainer(sso *apiv1.SSO, filter string) *v1.Container {
	return &v1.Container{
		Name:            fmt.Sprintf("%s-cleanup", sso.GetName()),
		Image:           fmt.Sprintf("%s:%s", exposeImage, exposeImageTag),
		ImagePullPolicy: v1.PullIfNotPresent,
		Command:         []string{exposeCmd},
		Args:            []string{fmt.Sprintf("--config=%s", exposeConfigPath), "--cleanup", fmt.Sprintf("--filter=%s", filter)},
		VolumeMounts: []v1.VolumeMount{{
			Name:      exposeConfigVolumeName,
			ReadOnly:  true,
			MountPath: filepath.Dir(exposeConfigPath),
		}},
		Env: []v1.EnvVar{{
			Name:  exposeEnv,
			Value: sso.GetNamespace(),
		}},
	}
}

func exposeConfigMap(sso *apiv1.SSO, serviceName string) (*v1.ConfigMap, error) {
	exposeConfig := &ExposeConfig{
		Domain:   sso.Spec.Domain,
		Exposer:  exposer,
		PathMode: "",
		HTTP:     false,
		TLSAcme:  true,
		Services: []string{serviceName},
		UrlTemplate: sso.Spec.UrlTemplate,
	}

	config, err := renderExposeConfig(exposeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "rendering expose config")
	}

	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      exposeConfigMapName,
			Namespace: sso.GetNamespace(),
		},
		Data: map[string]string{
			filepath.Base(exposeConfigPath): config,
		},
	}, nil
}

// ExposeConfig holds the configuration for exposecontroller
type ExposeConfig struct {
	Domain   string   `yaml:"domain,omitempty" json:"domain"`
	Exposer  string   `yaml:"exposer" json:"exposer"`
	PathMode string   `yaml:"path-mode" json:"path_mode"`
	HTTP     bool     `yaml:"http" json:"http"`
	TLSAcme  bool     `yaml:"tls-acme" json:"tls_acme"`
	Services []string `yaml:"services,omitempty" json:"services"`
	UrlTemplate string `yaml:"urltemplate,omitempty" json:"urltemplate"`
}

func renderExposeConfig(config *ExposeConfig) (string, error) {
	b, err := yaml.Marshal(config)
	if err != nil {
		return "", errors.Wrap(err, "marshaling expose config to YAML")
	}
	return string(b), nil
}
