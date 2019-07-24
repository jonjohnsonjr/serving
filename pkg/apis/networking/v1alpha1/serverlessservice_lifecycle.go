/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

var sksActivatorCondSet = apis.NewLivingConditionSet(
	ServerlessServiceConditionActivatorEndpointsPopulated,
)
var sksEndpointsCondSet = apis.NewLivingConditionSet(
	ServerlessServiceConditionEndpointsPopulated,
)

// GetGroupVersionKind returns the GVK for the ServerlessService.
func (ss *ServerlessService) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServerlessService")
}

// GetCondition returns the value of the condition `t`.
func (sss *ServerlessServiceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return sss.condSet().Manage(sss).GetCondition(t)
}

// InitializeConditions initializes the conditions.
func (sss *ServerlessServiceStatus) InitializeConditions() {
	sss.condSet().Manage(sss).InitializeConditions()
}

// MarkEndpointsReady marks the ServerlessServiceStatus endpoints populated condition to true.
func (sss *ServerlessServiceStatus) MarkEndpointsReady() {
	sss.condSet().Manage(sss).MarkTrue(ServerlessServiceConditionEndpointsPopulated)
}

// MarkEndpointsNotOwned marks that we don't own K8s service.
func (sss *ServerlessServiceStatus) MarkEndpointsNotOwned(kind, name string) {
	sss.condSet().Manage(sss).MarkFalse(
		ServerlessServiceConditionEndpointsPopulated, "NotOwned",
		"Resource %s of type %s is not owned by SKS", name, kind)
}

// MarkActivatorEndpointsPopulated is setting the ActivatorEndpointsPopulated to True.
func (sss *ServerlessServiceStatus) MarkActivatorEndpointsPopulated() {
	sss.condSet().Manage(sss).SetCondition(apis.Condition{
		Type:    ServerlessServiceConditionActivatorEndpointsPopulated,
		Status:  corev1.ConditionTrue,
		Reason:  "ActivatorEndpointsPopulated",
		Message: "Revision is backed by Activator",
	})
}

// MarkActivatorEndpointsRemoved is setting the ActivatorEndpointsPopulated to False.
func (sss *ServerlessServiceStatus) MarkActivatorEndpointsRemoved() {
	sss.condSet().Manage(sss).SetCondition(apis.Condition{
		Type:    ServerlessServiceConditionActivatorEndpointsPopulated,
		Status:  corev1.ConditionFalse,
		Reason:  "ActivatorEndpointsPopulated",
		Message: "Revision is backed by Activator",
	})
}

// MarkEndpointsNotReady marks the ServerlessServiceStatus endpoints populated condition to unknown.
func (sss *ServerlessServiceStatus) MarkEndpointsNotReady(reason string) {
	sss.condSet().Manage(sss).MarkUnknown(
		ServerlessServiceConditionEndpointsPopulated, reason,
		"K8s Service is not ready")
}

// IsReady returns true if ServerlessService is ready.
func (sss *ServerlessServiceStatus) IsReady() bool {
	return sss.condSet().Manage(sss).IsHappy()
}

// EndpointsPopulated returns true when EndpointsPopulated is true.
func (sss *ServerlessServiceStatus) EndpointsPopulated() bool {
	cond := sss.GetCondition(ServerlessServiceConditionEndpointsPopulated)
	return cond != nil && cond.Status == corev1.ConditionTrue
}

func (sss *ServerlessServiceStatus) duck() *duckv1beta1.Status {
	return &sss.Status
}

func (sss *ServerlessServiceStatus) condSet() apis.ConditionSet {
	if sss.Mode == SKSOperationModeProxy {
		return sksActivatorCondSet
	}
	return sksEndpointsCondSet
}
