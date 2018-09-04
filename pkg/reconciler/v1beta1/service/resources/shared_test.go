/*
Copyright 2018 The Knative Authors

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

package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/knative/serving/pkg/apis/serving/v1beta1"
)

const (
	testServiceName            = "test-service"
	testServiceNamespace       = "test-service-namespace"
	testRevisionName           = "test-revision-name"
	testContainerNameRunLatest = "test-container-run-latest"
	testContainerNamePinned    = "test-container-pinned"
	testLabelKey               = "test-label-key"
	testLabelValuePinned       = "test-label-value-pinned"
	testLabelValueRunLatest    = "test-label-value-run-latest"
)

func expectOwnerReferencesSetCorrectly(t *testing.T, ownerRefs []metav1.OwnerReference) {
	if got, want := len(ownerRefs), 1; got != want {
		t.Errorf("expected %d owner refs got %d", want, got)
		return
	}

	expectedRefs := []metav1.OwnerReference{{
		APIVersion: "serving.knative.dev/v1beta1",
		Kind:       "Service",
		Name:       testServiceName,
	}}
	if diff := cmp.Diff(expectedRefs, ownerRefs, cmpopts.IgnoreFields(expectedRefs[0], "Controller", "BlockOwnerDeletion")); diff != "" {
		t.Errorf("Unexpected service owner refs diff (-want +got): %v", diff)
	}
}

func createConfiguration(containerName string) v1beta1.ConfigurationSpec {
	return v1beta1.ConfigurationSpec{
		RevisionTemplate: v1beta1.RevisionTemplateSpec{
			Spec: v1beta1.RevisionSpec{
				Container: corev1.Container{
					Name: containerName,
				},
			},
		},
	}
}

func createServiceMeta() *v1beta1.Service {
	return &v1beta1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testServiceName,
			Namespace: testServiceNamespace,
		},
	}
}
func createServiceWithRunLatest() *v1beta1.Service {
	s := createServiceMeta()
	s.Spec = v1beta1.ServiceSpec{
		RunLatest: &v1beta1.RunLatestType{
			Configuration: createConfiguration(testContainerNameRunLatest),
		},
	}
	s.Labels = make(map[string]string, 2)
	s.Labels[testLabelKey] = testLabelValueRunLatest
	return s
}

func createServiceWithPinned() *v1beta1.Service {
	s := createServiceMeta()
	s.Spec = v1beta1.ServiceSpec{
		Pinned: &v1beta1.PinnedType{
			RevisionName:  testRevisionName,
			Configuration: createConfiguration(testContainerNamePinned),
		},
	}
	s.Labels = make(map[string]string, 2)
	s.Labels[testLabelKey] = testLabelValuePinned
	return s
}
