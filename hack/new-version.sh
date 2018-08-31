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

OLD_VERSION=$1
NEW_VERSION=$2

UPPERCASE_OLD_VERSION=$(echo ${OLD_VERSION} | sed 's/./V/1')
UPPERCASE_NEW_VERSION=$(echo ${NEW_VERSION} | sed 's/./V/1')

# I'm lazy, so assuming this 
SERVING_ROOT=$(dirname ${BASH_SOURCE})/..

DIRS=$(find $SERVING_ROOT -type d |
       grep -v vendor | # Ignore vendor
       # grep -v -e "pkg/client" | # Ignore generated
       grep -e "${OLD_VERSION}$")

for DIR in $DIRS; do
  PARENT=$(dirname $DIR)
  NEW_DIR="${PARENT}/${NEW_VERSION}"
  GO_FROM=$(echo "$DIR" | sed "s@${SERVING_ROOT}@github.com/knative/serving@")
  GO_TO=$(echo "$NEW_DIR" | sed "s@${SERVING_ROOT}@github.com/knative/serving@")
  gomvpkg -from ${GO_FROM} -to ${GO_TO}
done

for DIR in $DIRS; do
  PARENT=$(dirname $DIR)
  NEW_DIR="${PARENT}/${NEW_VERSION}"
  git add $NEW_DIR
done

git commit -m "Fork $OLD_VERSION to $NEW_VERSION"

${SERVING_ROOT}/hack/update-deps.sh

git commit -am "Run update-deps.sh"

find ${SERVING_ROOT}/ -type f -exec sed -i -e "s/Serving${UPPERCASE_OLD_VERSION}/Serving${UPPERCASE_NEW_VERSION}/g" {} \;
find ${SERVING_ROOT}/ -type f -exec sed -i -e "s/Autoscaling${UPPERCASE_OLD_VERSION}/Autoscaling${UPPERCASE_NEW_VERSION}/g" {} \;
find ${SERVING_ROOT}/ -type f -exec sed -i "s/Autoscaling().${UPPERCASE_OLD_VERSION}()/Autoscaling().${UPPERCASE_NEW_VERSION}()/g" {} \;
find ${SERVING_ROOT}/ -type f -exec sed -i "s/Serving().${UPPERCASE_OLD_VERSION}()/Serving().${UPPERCASE_NEW_VERSION}()/g" {} \;

sed -i "s/${OLD_VERSION}/${NEW_VERSION}/g" "${SERVING_ROOT}/pkg/apis/autoscaling/${NEW_VERSION}/kpa_validation_test.go"
sed -i "s/${OLD_VERSION}/${NEW_VERSION}/g" "${SERVING_ROOT}/pkg/apis/serving/${NEW_VERSION}/register.go"
sed -i "s/${OLD_VERSION}/${NEW_VERSION}/g" "${SERVING_ROOT}/pkg/apis/autoscaling/${NEW_VERSION}/register.go"


# TODO: Things that are hard.
# * kpa_validation_test has error messages hard-coded with versions embedded.
# * pkg/apis/serving/${NEW_VERSION}/register.go s/old/new/
# * pkg/apis/autoscaling/${NEW_VERSION}/register.go s/old/new/
#
# * NewController takes listers
