#!/bin/bash
set -xe

for i in `seq 1 1000`; do
	go test -count=1 ./pkg/reconciler/taskrun/... -run Test_UseConfigMapQuickly -mod=vendor
done
