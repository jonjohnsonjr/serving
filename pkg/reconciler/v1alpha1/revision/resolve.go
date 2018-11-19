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

package revision

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/k8stransport"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/client-go/kubernetes"
)

type digestResolver struct {
	client    kubernetes.Interface
	transport http.Transport

	// Holds a list of paths we have already added to the transport's cert bundle.
	trusted map[string]struct{}
}

// Resolve resolves the image references that use tags to digests.
func (r *digestResolver) Resolve(
	image string,
	opt k8schain.Options,
	registriesToSkip map[string]struct{},
	certFilesToTrust []string,
) (string, error) {
	kc, err := k8schain.New(r.client, opt)
	if err != nil {
		return "", err
	}

	if _, err := name.NewDigest(image, name.WeakValidation); err == nil {
		// Already a digest
		return image, nil
	}

	if err := r.trust(certFilesToTrust); err != nil {
		return "", err
	}

	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return "", err
	}

	if _, ok := registriesToSkip[tag.Registry.RegistryStr()]; ok {
		return "", nil
	}

	img, err := remote.Image(tag, remote.WithTransport(&r.transport), remote.WithAuthFromKeychain(kc))
	if err != nil {
		return "", err
	}
	digest, err := img.Digest()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", tag.Repository.String(), digest), nil
}

// trust adds certificate bundles to the resolver's transport, as necessary.
func (r *digestResolver) trust(certFiles []string) error {
	if r.trusted == nil {
		r.trusted = make(map[string]struct{})
	}

	// Construct a map of the new certs we need to trust.
	newCerts := make(map[string]struct{})
	for _, crt := range certFiles {
		newCerts[crt] = struct{}{}
	}

	if r.transport.TLSClientConfig == nil {
		r.transport.TLSClientConfig = &tls.Config{}
	}

	pool := r.transport.TLSClientConfig.RootCAs

	// If we can't reuse the cert pool (or don't have one), just create a new one.
	if !canReuseCertPool(r.trusted, newCerts) || pool == nil {
		roots, err := k8stransport.NewCertPool(certFiles...)
		if err != nil {
			return err
		}
		r.transport.TLSClientConfig.RootCAs = roots
		r.trusted = newCerts
		return nil
	}

	// Append anything in newCerts that isn't already trusted.
	for crt := range newCerts {
		if _, ok := r.trusted[crt]; !ok {
			if err := k8stransport.AppendCertBundle(pool, crt); err != nil {
				return err
			}
			r.trusted[crt] = struct{}{}
		}
	}

	return nil
}

func canReuseCertPool(oldCerts, newCerts map[string]struct{}) bool {
	for crt, _ := range oldCerts {
		if _, ok := newCerts[crt]; !ok {
			// We are "untrusting" a path we trusted before, throw the CertPool away.
			return false
		}
	}

	// newCerts is a superset of oldCerts, so we can reuse the CertPool.
	return true
}
