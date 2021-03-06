#!/usr/bin/env bash

set -e

if grep --include \*.go -r 'log\.' . | grep -v vendor \
  | grep -v envoy \
  | grep -v contrib \
  | grep -v logging.go \
  | grep -v pkg/k8s/slim/k8s/apis/util/intstr/intsrt.go \
  | grep -F "\n"; then
  echo "found newline(s) in log call(s), please remove ending \n"
  exit 1
fi
