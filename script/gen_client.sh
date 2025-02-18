#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

location=$(dirname $0)
rootdir=$(realpath ${location}/../)

unset GOPATH
GO111MODULE=on

# Entering the client module
cd $rootdir/pkg/client/yaks

echo "Generating Go client code..."

go run k8s.io/code-generator/cmd/client-gen \
	--input=yaks/v1alpha1 \
	--go-header-file=$rootdir/script/headers/default.txt \
	--clientset-name "versioned"  \
	--input-base=github.com/citrusframework/yaks/pkg/apis \
	--output-base=. \
	--output-package=github.com/citrusframework/yaks/pkg/client/yaks/clientset

go run k8s.io/code-generator/cmd/lister-gen \
	--input-dirs=github.com/citrusframework/yaks/pkg/apis/yaks/v1alpha1 \
	--go-header-file=$rootdir/script/headers/default.txt \
	--output-base=. \
	--output-package=github.com/citrusframework/yaks/pkg/client/yaks/listers

go run k8s.io/code-generator/cmd/informer-gen \
  --versioned-clientset-package=github.com/citrusframework/yaks/pkg/client/yaks/clientset/versioned \
	--listers-package=github.com/citrusframework/yaks/pkg/client/yaks/listers \
	--input-dirs=github.com/citrusframework/yaks/pkg/apis/yaks/v1alpha1 \
	--go-header-file=$rootdir/script/headers/default.txt \
	--output-base=. \
	--output-package=github.com/citrusframework/yaks/pkg/client/yaks/informers

# hack to fix non go-module compliance
rm -rf ./clientset
rm -rf ./informers
rm -rf ./listers
cp -R ./github.com/citrusframework/yaks/pkg/client/yaks/* .
rm -rf ./github.com

# hack to fix test custom resource generated fake. otherwise generated fake file is handled as test scoped file
mv $rootdir/pkg/client/yaks/clientset/versioned/typed/yaks/v1alpha1/fake/fake_test.go \
   $rootdir/pkg/client/yaks/clientset/versioned/typed/yaks/v1alpha1/fake/fake_tests.go
