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
)

// Expose executes the exposecontroller as a Job in order publicly expose the SSO service
func Expose(sso *apiv1.SSO, serviceName string) error {
	configMap, err := exposeConfigMap(sso, serviceName)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "building expose config map")
	}
	configMap.SetOwnerReferences(append(configMap.GetOwnerReferences(), ownerRef(sso)))
	err = sdk.Create(configMap)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "creating expose config map")
	}

	job := exposeJob(sso)
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

	err = sdk.Delete(job)
	if err != nil {
		return errors.Wrap(err, "cleaning up the expose job")
	}

	err = sdk.Delete(configMap)
	if err != nil {
		return errors.Wrap(err, "cleaning up the expose config map")
	}

	return nil
}

func exposeJob(sso *apiv1.SSO) *batchv1.Job {
	ns := sso.GetNamespace()

	podTempl := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName(sso.GetName(), ns),
			Namespace: ns,
			Labels:    labels(sso),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{exposeContainer(sso)},
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

	jobName := buildName(sso.GetName(), ns)
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ns,
			Labels:    labels(sso),
		},
		Spec: batchv1.JobSpec{
			Template: podTempl,
		},
	}
}

func exposeContainer(sso *apiv1.SSO) v1.Container {
	return v1.Container{
		Name:            sso.GetName(),
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

func exposeConfigMap(sso *apiv1.SSO, serviceName string) (*v1.ConfigMap, error) {
	exposeConfig := &ExposeConfig{
		Domain:   sso.Spec.Domain,
		Exposer:  exposer,
		PathMode: "",
		HTTP:     !sso.Spec.TLS,
		TLSAcme:  sso.Spec.TLS,
		Services: []string{serviceName},
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
			Labels:    labels(sso),
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
}

func renderExposeConfig(config *ExposeConfig) (string, error) {
	b, err := yaml.Marshal(config)
	if err != nil {
		return "", errors.Wrap(err, "marshaling expose config to YAML")
	}
	return string(b), nil
}
