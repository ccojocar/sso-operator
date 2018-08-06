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
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// WaitForPodReady waits for given POD to become ready
func WaitForPodReady(pods corev1.PodInterface, podName string) error {
	logrus.Infof("Waiting for %s to be scheduled", podName)
	err := wait.PollImmediate(time.Millisecond*500, time.Second*10, func() (bool, error) {
		_, err := pods.Get(podName, meta_v1.GetOptions{
			IncludeUninitialized: true,
		})
		if err != nil {
			logrus.Infof("Getting pod %s", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	logrus.Infof("Waiting for %s to be ready", podName)
	return wait.PollImmediate(time.Millisecond*500, time.Minute*10, func() (bool, error) {
		pod, err := pods.Get(podName, meta_v1.GetOptions{
			IncludeUninitialized: true,
		})
		if err != nil {
			return false, fmt.Errorf("not found: %s", podName)
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodSucceeded, v1.PodFailed:
			return false, fmt.Errorf("pod already in terminal phase: %s", pod.Status.Phase)
		case v1.PodUnknown, v1.PodPending:
			return false, nil
		}
		return false, fmt.Errorf("unknown phase: %s", pod.Status.Phase)
	})
}

// WaitForPodComplete wait for given POD to complete
func WaitForPodComplete(pods corev1.PodInterface, podName string, timeout time.Duration) error {
	logrus.Infof("Waiting for %s to be ready", podName)
	return wait.PollImmediate(time.Millisecond*500, timeout, func() (bool, error) {
		pod, err := pods.Get(podName, meta_v1.GetOptions{
			IncludeUninitialized: true,
		})
		if err != nil {
			logrus.Infof("Getting pod %s", err)
			return false, nil
		}
		switch pod.Status.Phase {
		case v1.PodSucceeded:
			return true, nil
		case v1.PodRunning:
			return false, nil
		case v1.PodFailed:
			return false, fmt.Errorf("pod already in terminal phase: %s", pod.Status.Phase)
		case v1.PodUnknown, v1.PodPending:
			return false, nil
		}
		return false, fmt.Errorf("unknown phase: %s", pod.Status.Phase)
	})
}

// WaitForPodsWithLabelRunning waits up to 10 minutes for all matching pods to become Running and at least one
// matching pod exists.
func WaitForPodsWithLabelRunning(c kubernetes.Interface, namespace string, label labels.Selector) error {
	lastKnownPodNumber := -1
	return wait.PollImmediate(500*time.Millisecond, time.Minute*10, func() (bool, error) {
		listOpts := meta_v1.ListOptions{LabelSelector: label.String()}
		pods, err := c.CoreV1().Pods(namespace).List(listOpts)
		if err != nil {
			logrus.Infof("error getting Pods with label selector %q [%v]\n", label.String(), err)
			return false, nil
		}

		if lastKnownPodNumber != len(pods.Items) {
			logrus.Infof("Found %d Pods for label selector %s\n", len(pods.Items), label.String())
			lastKnownPodNumber = len(pods.Items)
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != v1.PodRunning {
				return false, nil
			}
		}

		return true, nil
	})
}

// WaitForDeploymentToStabilize waits till the Deployment has a matching generation/replica count between spec and status.
func WaitForDeploymentToStabilize(c kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	options := meta_v1.ListOptions{FieldSelector: fields.Set{
		"metadata.name":      name,
		"metadata.namespace": namespace,
	}.AsSelector().String()}
	w, err := c.AppsV1().Deployments(namespace).Watch(options)
	if err != nil {
		return err
	}
	_, err = watch.Until(timeout, w, func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Deleted:
			return false, apierrs.NewNotFound(schema.GroupResource{Resource: "deployments"}, "")
		}
		switch dp := event.Object.(type) {
		case *appsv1.Deployment:
			if dp.Name == name && dp.Namespace == namespace &&
				dp.Generation <= dp.Status.ObservedGeneration &&
				*(dp.Spec.Replicas) == dp.Status.Replicas {
				logrus.Infof("Deployment %s in namespace %s ready.", name, namespace)
				return true, nil
			}
			logrus.Infof("Waiting for deployment %s to stabilize, generation %v observed generation %v spec.replicas %d status.replicas %d",
				name, dp.Generation, dp.Status.ObservedGeneration, *(dp.Spec.Replicas), dp.Status.Replicas)
		}
		return false, nil
	})
	return err
}

// WaitForService waits until the service appears (exist == true), or disappears (exist == false)
func WaitForService(c kubernetes.Interface, namespace, name string, exist bool, interval, timeout time.Duration) error {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		_, err := c.CoreV1().Services(namespace).Get(name, meta_v1.GetOptions{})
		switch {
		case err == nil:
			logrus.Infof("Service %s in namespace %s found.", name, namespace)
			return exist, nil
		case apierrs.IsNotFound(err):
			logrus.Infof("Service %s in namespace %s disappeared.", name, namespace)
			return !exist, nil
		case !IsRetryableAPIError(err):
			logrus.Infof("Non-retryable failure while getting service.")
			return false, err
		default:
			logrus.Infof("Get service %s in namespace %s failed: %v", name, namespace, err)
			return false, nil
		}
	})
	if err != nil {
		stateMsg := map[bool]string{true: "to appear", false: "to disappear"}
		return fmt.Errorf("error waiting for service %s/%s %s: %v", namespace, name, stateMsg[exist], err)
	}
	return nil
}

//WaitForServiceEndpointsNum waits until the amount of endpoints that implement service to expectNum.
func WaitForServiceEndpointsNum(c kubernetes.Interface, namespace, serviceName string, expectNum int, interval, timeout time.Duration) error {
	return wait.Poll(interval, timeout, func() (bool, error) {
		logrus.Infof("Waiting for amount of service:%s endpoints to be %d", serviceName, expectNum)
		list, err := c.CoreV1().Endpoints(namespace).List(meta_v1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, e := range list.Items {
			if e.Name == serviceName && countEndpointsNum(&e) == expectNum {
				return true, nil
			}
		}
		return false, nil
	})
}

func countEndpointsNum(e *v1.Endpoints) int {
	num := 0
	for _, sub := range e.Subsets {
		num += len(sub.Addresses)
	}
	return num
}

// IsRetryableAPIError indicates if the given error is retryable
func IsRetryableAPIError(err error) bool {
	return apierrs.IsTimeout(err) || apierrs.IsServerTimeout(err) || apierrs.IsTooManyRequests(err) || apierrs.IsInternalError(err)
}

// WaitForJobComplete wait for a job to complete
func WaitForJobComplete(c kubernetes.Interface, namespace, name string, interval, timeout time.Duration) error {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		job, err := c.BatchV1().Jobs(namespace).Get(name, meta_v1.GetOptions{})
		switch {
		case err == nil:
			conditions := job.Status.Conditions
			if len(conditions) == 0 {
				logrus.Infof("No condition found for job %s in namespace %s", name, namespace)
				return false, nil
			}
			for _, condition := range conditions {
				if condition.Type != batchv1.JobComplete {
					return false, fmt.Errorf("job failed: %s", condition.Message)
				}
			}
			return true, nil
		case apierrs.IsNotFound(err):
			logrus.Infof("job %s in namespace %s not found.", name, namespace)
			return false, nil
		case !IsRetryableAPIError(err):
			logrus.Infof("Non-retryable failure while getting job.")
			return false, err
		default:
			logrus.Infof("Get job %s in namespace %s failed: %v", name, namespace, err)
			return false, nil
		}
	})
	if err != nil {
		return fmt.Errorf("error waiting for job %s/%s: %v", namespace, name, err)
	}
	return nil
}
