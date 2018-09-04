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

package resources

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	istiov1alpha1 "github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	"github.com/knative/serving/pkg/reconciler"
	"github.com/knative/serving/pkg/reconciler/v1beta1/route/traffic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeVirtualServiceSpec_CorrectMetadata(t *testing.T) {
	targets := map[string][]traffic.RevisionTarget{}
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels:    map[string]string{"route": "test-route"},
		},
		Status: v1beta1.RouteStatus{Domain: "domain.com"},
	}
	expected := metav1.ObjectMeta{
		Name:      "test-route",
		Namespace: "test-ns",
		Labels:    map[string]string{"route": "test-route"},
		OwnerReferences: []metav1.OwnerReference{
			*reconciler.NewControllerRef(r),
		},
	}
	meta := MakeVirtualService(r, &traffic.TrafficConfig{Targets: targets}).ObjectMeta
	if diff := cmp.Diff(expected, meta); diff != "" {
		t.Errorf("Unexpected metadata (-want +got): %v", diff)
	}
}

func TestMakeVirtualServiceSpec_CorrectSpec(t *testing.T) {
	targets := map[string][]traffic.RevisionTarget{}
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels:    map[string]string{"route": "test-route"},
		},
		Status: v1beta1.RouteStatus{Domain: "domain.com"},
	}
	expected := v1alpha3.VirtualServiceSpec{
		// We want to connect to two Gateways: the Route's ingress Gateway, and the 'mesh' Gateway.  The former
		// provides access from outside of the cluster, and the latter provides access for services from inside
		// the cluster.
		Gateways: []string{
			"knative-shared-gateway.knative-serving.svc.cluster.local",
			"mesh",
		},
		Hosts: []string{
			"*.domain.com",
			"domain.com",
			"test-route.test-ns.svc.cluster.local",
		},
	}
	routes := MakeVirtualService(r, &traffic.TrafficConfig{Targets: targets}).Spec
	if diff := cmp.Diff(expected, routes); diff != "" {
		t.Errorf("Unexpected routes (-want +got): %v", diff)
	}
}

func TestMakeVirtualServiceSpec_CorrectRoutes(t *testing.T) {
	targets := map[string][]traffic.RevisionTarget{
		"": {{
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v2",
				Percent:           100,
			},
			Active: true,
		}},
		"v1": {{
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v1",
				Percent:           100,
			},
			Active: true,
		}},
	}
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels:    map[string]string{"route": "test-route"},
		},
		Status: v1beta1.RouteStatus{Domain: "domain.com"},
	}
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := []v1alpha3.HTTPRoute{{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "domain.com"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc.cluster.local"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "v2-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}, {
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "v1.domain.com"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "v1-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}}
	routes := MakeVirtualService(r, &traffic.TrafficConfig{Targets: targets}).Spec.Http
	if diff := cmp.Diff(expected, routes); diff != "" {
		fmt.Printf("%+v\n", routes)
		fmt.Printf("%+v\n", expected)
		t.Errorf("Unexpected routes (-want +got): %v", diff)
	}
}

func TestMakeVirtualServiceSpec_SingileRevisionTimeout(t *testing.T) {
	targets := map[string][]traffic.RevisionTarget{
		"": {{
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v2",
				Percent:           100,
			},
			Active: true,
			TimeoutSeconds: 10,
		}},
	}
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels:    map[string]string{"route": "test-route"},
		},
		Status: v1beta1.RouteStatus{Domain: "domain.com"},
	}
	expected := []v1alpha3.HTTPRoute{{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "domain.com"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc.cluster.local"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "v2-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		Timeout: "10s",
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: "10s",
		},
	}}
	routes := MakeVirtualService(r, &traffic.TrafficConfig{Targets: targets}).Spec.Http
	if diff := cmp.Diff(expected, routes); diff != "" {
		fmt.Printf("%+v\n", routes)
		fmt.Printf("%+v\n", expected)
		t.Errorf("Unexpected routes (-want +got): %v", diff)
	}
}

func TestMakeVirtualServiceSpec_MultipleRevisionTimeout(t *testing.T) {
	targets := map[string][]traffic.RevisionTarget{
		"": {{
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v2",
				Percent:           50,
			},
			Active: true,
			TimeoutSeconds: 10,
		}, {
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v1",
				Percent:           50,
			},
			Active: true,
			TimeoutSeconds: 20,
		}},
	}
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels:    map[string]string{"route": "test-route"},
		},
		Status: v1beta1.RouteStatus{Domain: "domain.com"},
	}
	expected := []v1alpha3.HTTPRoute{{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "domain.com"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc.cluster.local"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "v2-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 50,
		}, {
			Destination: v1alpha3.Destination{
				Host: "v1-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 50,
		}},
		// The expected timeout is 20s, which is the max value of revision "v1" timeout (20s)
		// and revision "v2" timeout (10s).
		Timeout: "20s",
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: "20s",
		},
	}}
	routes := MakeVirtualService(r, &traffic.TrafficConfig{Targets: targets}).Spec.Http
	if diff := cmp.Diff(expected, routes); diff != "" {
		fmt.Printf("%+v\n", routes)
		fmt.Printf("%+v\n", expected)
		t.Errorf("Unexpected routes (-want +got): %v", diff)
	}
}

func TestMakeVirtualServiceSpec_ZeroPercentRevisionTimeout(t *testing.T) {
	targets := map[string][]traffic.RevisionTarget{
		"": {{
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v2",
				Percent:           100,
			},
			Active: true,
			TimeoutSeconds: 10,
		}, {
			TrafficTarget: v1beta1.TrafficTarget{
				ConfigurationName: "config",
				RevisionName:      "v1",
				Percent:           0,
			},
			Active: true,
			TimeoutSeconds: 20,
		}},
	}
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels:    map[string]string{"route": "test-route"},
		},
		Status: v1beta1.RouteStatus{Domain: "domain.com"},
	}
	expected := []v1alpha3.HTTPRoute{{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "domain.com"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc.cluster.local"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns.svc"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route.test-ns"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "test-route"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "v2-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		// The expected timeout is 10s. The timeout of revision "v2" is ignored because
		// its percent is 0.
		Timeout: "10s",
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: "10s",
		},
	}}
	routes := MakeVirtualService(r, &traffic.TrafficConfig{Targets: targets}).Spec.Http
	if diff := cmp.Diff(expected, routes); diff != "" {
		fmt.Printf("%+v\n", routes)
		fmt.Printf("%+v\n", expected)
		t.Errorf("Unexpected routes (-want +got): %v", diff)
	}
}

func TestGetRouteDomains_NamelessTarget(t *testing.T) {
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels: map[string]string{
				"route": "test-route",
			},
		},
	}
	base := "domain.com"
	expected := []string{base,
		"test-route.test-ns.svc.cluster.local",
		"test-route.test-ns.svc",
		"test-route.test-ns",
		"test-route",
	}
	domains := getRouteDomains("", r, base)
	if diff := cmp.Diff(expected, domains); diff != "" {
		t.Errorf("Unexpected domains  (-want +got): %v", diff)
	}
}

func TestGetRouteDomains_NamedTarget(t *testing.T) {
	r := &v1beta1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels: map[string]string{
				"route": "test-route",
			},
		},
	}
	name := "v1"
	base := "domain.com"
	expected := []string{"v1.domain.com"}
	domains := getRouteDomains(name, r, base)
	if diff := cmp.Diff(expected, domains); diff != "" {
		t.Errorf("Unexpected domains  (-want +got): %v", diff)
	}
}

// One active target.
func TestMakeVirtualServiceRoute_Vanilla(t *testing.T) {
	targets := []traffic.RevisionTarget{{
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "config",
			RevisionName:      "revision",
			Percent:           100,
		},
		Active: true,
	}}
	domains := []string{"a.com", "b.org"}
	ns := "test-ns"
	route := makeVirtualServiceRoute(domains, ns, targets)
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := v1alpha3.HTTPRoute{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "a.com"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "b.org"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "revision-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}
	if diff := cmp.Diff(&expected, route); diff != "" {
		t.Errorf("Unexpected route  (-want +got): %v", diff)
	}
}

func TestMakeVirtualServiceRoute_ZeroPercentTarget(t *testing.T) {
	targets := []traffic.RevisionTarget{{
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "config",
			RevisionName:      "revision",
			Percent:           100,
		},
		Active: true,
	}, {
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "new-config",
			RevisionName:      "new-revision",
			Percent:           0,
		},
		Active: true,
	}}
	domains := []string{"test.org"}
	ns := "test-ns"
	route := makeVirtualServiceRoute(domains, ns, targets)
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := v1alpha3.HTTPRoute{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "test.org"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "revision-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}
	if diff := cmp.Diff(&expected, route); diff != "" {
		t.Errorf("Unexpected route  (-want +got): %v", diff)
	}
}

// Two active targets.
func TestMakeVirtualServiceRoute_TwoTargets(t *testing.T) {
	targets := []traffic.RevisionTarget{{
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "config",
			RevisionName:      "revision",
			Percent:           90,
		},
		Active: true,
	}, {
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "new-config",
			RevisionName:      "new-revision",
			Percent:           10,
		},
		Active: true,
	}}
	domains := []string{"test.org"}
	ns := "test-ns"
	route := makeVirtualServiceRoute(domains, ns, targets)
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := v1alpha3.HTTPRoute{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "test.org"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "revision-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 90,
		}, {
			Destination: v1alpha3.Destination{
				Host: "new-revision-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 10,
		}},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}
	if diff := cmp.Diff(&expected, route); diff != "" {
		t.Errorf("Unexpected route  (-want +got): %v", diff)
	}
}

// One Inactive target.
func TestMakeVirtualServiceRoute_VanillaScaledToZero(t *testing.T) {
	targets := []traffic.RevisionTarget{{
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "config",
			RevisionName:      "revision",
			Percent:           100,
		},
		Active: false,
	}}
	domains := []string{"a.com", "b.org"}
	ns := "test-ns"
	route := makeVirtualServiceRoute(domains, ns, targets)
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := v1alpha3.HTTPRoute{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "a.com"},
		}, {
			Authority: &istiov1alpha1.StringMatch{Exact: "b.org"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "activator-service.knative-serving.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		AppendHeaders: map[string]string{
			"knative-serving-revision":      "revision",
			"knative-serving-configuration": "config",
			"knative-serving-namespace":     "test-ns",
		},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}
	if diff := cmp.Diff(&expected, route); diff != "" {
		t.Errorf("Unexpected route  (-want +got): %v", diff)
	}
}

// Two inactive targets.
func TestMakeVirtualServiceRoute_TwoInactiveTargets(t *testing.T) {
	targets := []traffic.RevisionTarget{{
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "config",
			RevisionName:      "revision",
			Percent:           90,
		},
		Active: false,
	}, {
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "new-config",
			RevisionName:      "new-revision",
			Percent:           10,
		},
		Active: false,
	}}
	domains := []string{"test.org"}
	ns := "test-ns"
	route := makeVirtualServiceRoute(domains, ns, targets)
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := v1alpha3.HTTPRoute{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "test.org"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "activator-service.knative-serving.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		AppendHeaders: map[string]string{
			"knative-serving-revision":      "revision",
			"knative-serving-configuration": "config",
			"knative-serving-namespace":     "test-ns",
		},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}
	if diff := cmp.Diff(&expected, route); diff != "" {
		t.Errorf("Unexpected route  (-want +got): %v", diff)
	}
}

// Named target scaled to 0.
func TestMakeVirtualServiceRoute_ZeroPercentNamedTargetScaledToZero(t *testing.T) {
	targets := []traffic.RevisionTarget{{
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "config",
			RevisionName:      "revision",
			Percent:           100,
		},
		Active: true,
	}, {
		TrafficTarget: v1beta1.TrafficTarget{
			ConfigurationName: "new-config",
			RevisionName:      "new-revision",
			Percent:           0,
		},
		Active: false,
	}}
	domains := []string{"test.org"}
	ns := "test-ns"
	route := makeVirtualServiceRoute(domains, ns, targets)
	timeout := fmt.Sprintf("%ds", v1beta1.DefaultRevisionTimeoutSeconds)
	expected := v1alpha3.HTTPRoute{
		Match: []v1alpha3.HTTPMatchRequest{{
			Authority: &istiov1alpha1.StringMatch{Exact: "test.org"},
		}},
		Route: []v1alpha3.DestinationWeight{{
			Destination: v1alpha3.Destination{
				Host: "revision-service.test-ns.svc.cluster.local",
				Port: v1alpha3.PortSelector{Number: 80},
			},
			Weight: 100,
		}},
		Timeout: timeout,
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      DefaultRouteRetryAttempts,
			PerTryTimeout: timeout,
		},
	}
	if diff := cmp.Diff(&expected, route); diff != "" {
		t.Errorf("Unexpected route  (-want +got): %v", diff)
	}
}
