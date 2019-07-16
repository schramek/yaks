/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jboss-fuse/yaks/pkg/apis/yaks/v1alpha1"
	"github.com/jboss-fuse/yaks/pkg/client"
	"github.com/jboss-fuse/yaks/pkg/util/kubernetes"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func newCmdTest(rootCmdOptions *RootCmdOptions) *cobra.Command {
	options := testCmdOptions{
		RootCmdOptions: rootCmdOptions,
	}

	cmd := cobra.Command{
		Use:   "test [test file to execute]",
		Short: "Execute a test on Kubernetes",
		Long:  `Deploys and execute a pod on Kubernetes for running tests.`,
		RunE:  options.run,
	}

	return &cmd
}

type testCmdOptions struct {
	*RootCmdOptions
}

func (o *testCmdOptions) validateArgs(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New(fmt.Sprintf("accepts exactly 1 arg, received %d", len(args)))
	}

	return nil
}

func (o *testCmdOptions) run(_ *cobra.Command, args []string) error {
	c, err := o.GetCmdClient()
	if err != nil {
		return err
	}

	_, err = o.createTest(c, args)
	return err
}

func (o *testCmdOptions) createTest(c client.Client, sources []string) (*v1alpha1.Test, error) {
	namespace := o.Namespace

	rawName := sources[0]
	fileName := kubernetes.SanitizeFileName(rawName)
	name := kubernetes.SanitizeName(rawName)

	if name == "" {
		return nil, errors.New("unable to determine test name")
	}

	data, err := o.loadData(sources[0])
	if err != nil {
		return nil, err
	}

	test := v1alpha1.Test{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.TestKind,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.TestSpec{
			Source: v1alpha1.SourceSpec{
				Name:     fileName,
				Content:  data,
				Language: v1alpha1.LanguageGherkin,
			},
		},
	}

	existed := false
	err = c.Create(o.Context, &test)
	if err != nil && k8serrors.IsAlreadyExists(err) {
		existed = true
		clone := test.DeepCopy()
		var key k8sclient.ObjectKey
		key, err = k8sclient.ObjectKeyFromObject(clone)
		if err != nil {
			return nil, err
		}
		err = c.Get(o.Context, key, clone)
		if err != nil {
			return nil, err
		}
		test.ResourceVersion = clone.ResourceVersion
		err = c.Update(o.Context, &test)
		if err != nil {
			return nil, err
		}
		// Reset status as well
		test.Status = v1alpha1.TestStatus{}
		err = c.Status().Update(o.Context, &test)
	}

	if err != nil {
		return nil, err
	}

	if !existed {
		fmt.Printf("test \"%s\" created\n", name)
	} else {
		fmt.Printf("test \"%s\" updated\n", name)
	}
	return &test, nil
}

func (*testCmdOptions) loadData(fileName string) (string, error) {
	var content []byte
	var err error

	if !strings.HasPrefix(fileName, "http://") && !strings.HasPrefix(fileName, "https://") {
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			return "", err
		}
	} else {
		/* #nosec */
		resp, err := http.Get(fileName)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
	}

	return string(content), nil
}
