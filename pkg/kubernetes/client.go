/*
Copyright 2018 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"fmt"

	"github.com/jenkins-x/sso-operator/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientset creates a new k8s client
func GetClientset() (kubernetes.Interface, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "getting client config for kubernetes client")
	}
	return kubernetes.NewForConfig(config)
}

// GetClientConfig return the k8s configuration
func GetClientConfig() (*restclient.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %s", err)
	}
	return clientConfig, nil
}

// GetAPIExtensionsClient returns the k8s api extensions client
func GetAPIExtensionsClient() (apiextensionsclientset.Interface, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "getting client config for api extensions client")
	}

	return apiextensionsclientset.NewForConfig(config)
}

// GetJenkinsClient returns the Jenkins CRDs client
func GetJenkinsClient() (versioned.Interface, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "getting client config for jenkins client")
	}

	return versioned.NewForConfig(config)
}
