// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8stransport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	// Kubernetes CA certificate bundle is mounted into the pod at this path.
	// Exposed for convenience.
	//
	// https://k8s.io/docs/tasks/tls/managing-tls-in-a-cluster/#trusting-tls-in-a-cluster
	KubernetesCertBundle = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	// OpenShift cluster-created CA certificate bundle is mounted into the pod at
	// this path. Exposed for convenience.
	//
	// https://docs.openshift.com/online/dev_guide/secrets.html#service-serving-certificate-secrets
	OpenShiftCertBundle = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
)

// NewTransport returns a new http.Transport that trusts the cert bundles at the given
// paths.
//
// Use this with KubernetesCertBundle to make requests from a pod that trust the
// same certificates as the underlying cluster nodes.
func NewTransport(paths ...string) (*http.Transport, error) {
	roots, err := NewCertPool(paths...)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: roots,
		},
	}, nil
}

// NewCertPool returns a crypto/x509.CertPool that trusts the cert bundles at
// the given paths.
//
// It appends the k8s CA cert bundle to the system cert pool, so it respects
// the SSL_CERT_FILE and SSL_CERT_DIR environment variables.
// See: https://golang.org/pkg/crypto/x509/
func NewCertPool(paths ...string) (*x509.CertPool, error) {
	// Respect SSL_CERT_FILE and SSL_CERT_DIR environment variables.
	roots, err := x509.SystemCertPool()
	if err != nil {
		// If we can't read the SystemCertPool, just create a new CertPool.
		roots = x509.NewCertPool()
	}

	for _, path := range paths {
		if err := AppendCertBundle(roots, path); err != nil {
			return nil, err
		}
	}

	return roots, nil
}

// AppendCertBundle appends the CA bundle at the given path to the given crypto/x509.CertPool.
func AppendCertBundle(pool *x509.CertPool, path string) error {
	if pool == nil {
		return fmt.Errorf("nil CertPool")
	}

	if crt, err := ioutil.ReadFile(path); err != nil {
		return fmt.Errorf("failed to read cert bundle %s: %v", path, err)
	} else if ok := pool.AppendCertsFromPEM(crt); !ok {
		return fmt.Errorf("failed to append cert bundle to cert pool")
	}

	return nil
}
