#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/operator-framework/operator-sdk-samples/ipfs-operator/pkg/generated \
github.com/operator-framework/operator-sdk-samples/ipfs-operator/pkg/apis \
storage:v1alpha1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
