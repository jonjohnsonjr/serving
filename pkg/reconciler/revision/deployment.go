/*
Copyright 2018 The Knative Authors.

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

package revision

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

var deploymentCondSet = apis.NewLivingConditionSet(
	DeploymentConditionPodsReady,
	DeploymentConditionReplicaSetReady,
	DeploymentConditionProgressing,
)

var podCondSet = apis.NewLivingConditionSet(
	PodConditionContainerReady,
	PodConditionResourcesAvailable,
	PodConditionPodScheduled,
)

// ConditionType represents a Deployment condition value.
const (
	// DeploymentConditionReady is set when
	DeploymentConditionReady = apis.ConditionReady
	// DeploymentConditionPodsReady is set when
	DeploymentConditionPodsReady apis.ConditionType = "PodsReady"
	// DeploymentConditionReplicaSetReady is set when
	DeploymentConditionReplicaSetReady apis.ConditionType = "ReplicaSetReady"
	// DeploymentConditionProgressing is set when
	// TODO: Test that this == appsv1.DeploymentProgressing
	DeploymentConditionProgressing apis.ConditionType = "Progressing"
)

type DeploymentStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// IsReady returns if the deployment is ready to serve the requested configuration.
func (ds *DeploymentStatus) IsReady() bool {
	return deploymentCondSet.Manage(ds).IsHappy()
}

// GetCondition returns the condition by name.
func (ds *DeploymentStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return deploymentCondSet.Manage(ds).GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (ds *DeploymentStatus) InitializeConditions() {
	deploymentCondSet.Manage(ds).InitializeConditions()
}

// ConditionType represents a Pod condition value.
const (
	// PodConditionReady is set when
	PodConditionReady = apis.ConditionReady
	// PodConditionContainerReady is set when
	PodConditionContainerReady apis.ConditionType = "ContainerReady"
	// PodConditionResourcesAvailable is set when
	PodConditionResourcesAvailable apis.ConditionType = "ResourcesAvailable"
	// PodConditionResourcesAvailable is set when
	// TODO: test this == corev1.PodScheduled
	PodConditionPodScheduled apis.ConditionType = "PodScheduled"
)

type PodStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// IsReady returns if the pod is ready to serve the requested configuration.
func (ps *PodStatus) IsReady() bool {
	return podCondSet.Manage(ps).IsHappy()
}

// GetCondition returns the condition by name.
func (ps *PodStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return podCondSet.Manage(ps).GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (ps *PodStatus) InitializeConditions() {
	podCondSet.Manage(ps).InitializeConditions()
}

func DiagnoseDeployment(deployment *appsv1.Deployment, kubeclient kubernetes.Interface, container string) (*DeploymentStatus, error) {
	ns := deployment.Namespace
	ds := DeploymentStatus{}

	ds.PropagateDeploymentStatus(&deployment.Status)

	// If a container keeps crashing (no active pods in the deployment although we want some)
	if *deployment.Spec.Replicas > 0 && deployment.Status.AvailableReplicas == 0 {
		pods, err := kubeclient.CoreV1().Pods(ns).List(metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector)})
		if err != nil {
			return nil, fmt.Errorf("error getting pods: %v", err)
		} else if len(pods.Items) > 0 {
			// Arbitrarily grab the very first pod, as they all should be crashing
			pod := pods.Items[0]
			ps := DiagnosePod(&pod, container)
			ds.PropagatePodStatus(ps)
		}
	}

	return &ds, nil
}

func DiagnosePod(pod *corev1.Pod, container string) *PodStatus {
	ps := &PodStatus{}

	// Update the revision status if pod cannot be scheduled(possibly resource constraints)
	// If pod cannot be scheduled then we expect the container status to be empty.
	ps.PropagatePodStatus(&pod.Status)

	// Find the user's container.
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == container {
			ps.PropagateContainerStatus(&status)
			break
		}
	}

	return ps
}

func (s *PodStatus) PropagatePodStatus(ps *corev1.PodStatus) {
	for _, cond := range ps.Conditions {
		switch cond.Type {
		case corev1.PodScheduled:
			podCondSet.Manage(s).SetCondition(fromPodCondition(cond, PodConditionPodScheduled))
		}
	}
}

func (s *PodStatus) PropagateContainerStatus(cs *corev1.ContainerStatus) {
	if t := cs.LastTerminationState.Terminated; t != nil {
		podCondSet.Manage(s).MarkFalse(PodConditionContainerReady, t.Reason, "Container terminated (%d): %s", t.ExitCode, t.Message)
	} else if w := cs.State.Waiting; w != nil {
		podCondSet.Manage(s).MarkFalse(PodConditionContainerReady, w.Reason, w.Message)
	}
}

func (s *DeploymentStatus) PropagateDeploymentStatus(ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {
		switch cond.Type {
		case appsv1.DeploymentProgressing:
			// Pass through Progressing.
			deploymentCondSet.Manage(s).SetCondition(fromDeploymentCondition(cond, DeploymentConditionProgressing))
		case appsv1.DeploymentReplicaFailure:
			// Turn this into a positive condition.
			rf := invert(fromDeploymentCondition(cond, DeploymentConditionReplicaSetReady))
			deploymentCondSet.Manage(s).SetCondition(rf)

		}
	}
}

func (s *DeploymentStatus) PropagatePodStatus(ps *PodStatus) {
	pc := ps.GetCondition(PodConditionReady)
	if pc == nil {
		return
	}
	switch {
	case pc.Status == corev1.ConditionUnknown:
		deploymentCondSet.Manage(s).MarkUnknown(DeploymentConditionPodsReady, pc.Reason, pc.Message)
	case pc.Status == corev1.ConditionTrue:
		deploymentCondSet.Manage(s).MarkTrue(DeploymentConditionPodsReady)
	case pc.Status == corev1.ConditionFalse:
		deploymentCondSet.Manage(s).MarkFalse(DeploymentConditionPodsReady, pc.Reason, pc.Message)
	}
}

func invert(c apis.Condition) apis.Condition {
	switch c.Status {
	case corev1.ConditionTrue:
		c.Status = corev1.ConditionFalse
	case corev1.ConditionFalse:
		c.Status = corev1.ConditionTrue
	}
	return c
}

func fromPodCondition(c corev1.PodCondition, t apis.ConditionType) apis.Condition {
	return apis.Condition{
		Type:    t,
		Status:  c.Status,
		Reason:  c.Reason,
		Message: c.Message,
	}
}

func fromDeploymentCondition(c appsv1.DeploymentCondition, t apis.ConditionType) apis.Condition {
	return apis.Condition{
		Type:    t,
		Status:  c.Status,
		Reason:  c.Reason,
		Message: c.Message,
	}
}
