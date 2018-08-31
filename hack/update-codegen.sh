#!/bin/bash

# Copyright 2018 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SERVING_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SERVING_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/knative/serving/pkg/client github.com/knative/serving/pkg/apis \
  "serving:v1alpha1 autoscaling:v1alpha1 serving:v1beta1 autoscaling:v1beta1" \
  --go-header-file ${SERVING_ROOT}/hack/boilerplate/boilerplate.go.txt

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen --input-dirs \
  github.com/knative/serving/pkg/reconciler/v1alpha1/revision/config,github.com/knative/serving/pkg/reconciler/v1beta1/revision/config,github.com/knative/serving/pkg/autoscaler,github.com/knative/serving/pkg/logging \
  -O zz_generated.deepcopy \
  --go-header-file ${SERVING_ROOT}/hack/boilerplate/boilerplate.go.txt

# Make sure our dependencies are up-to-date
${SERVING_ROOT}/hack/update-deps.sh
