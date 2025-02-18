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

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	yaml2 "gopkg.in/yaml.v2"
)

// ToJSON --.
func ToJSON(value runtime.Object) ([]byte, error) {
	return json.Marshal(value)
}

// ToYAML --.
func ToYAML(value runtime.Object) ([]byte, error) {
	data, err := ToJSON(value)
	if err != nil {
		return nil, err
	}

	return JSONToYAML(data)
}

// JSONToYAML --.
func JSONToYAML(src []byte) ([]byte, error) {
	jsondata := map[string]interface{}{}
	err := json.Unmarshal(src, &jsondata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %w", err)
	}
	yamldata, err := yaml2.Marshal(&jsondata)
	if err != nil {
		return nil, fmt.Errorf("error marshalling to yaml: %w", err)
	}

	return yamldata, nil
}

// GetConfigMap --.
func GetConfigMap(context context.Context, client k8sclient.Reader, name string, namespace string) (*corev1.ConfigMap, error) {
	key := k8sclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	answer := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := client.Get(context, key, &answer); err != nil {
		return nil, err
	}

	return &answer, nil
}

// GetSecret --.
func GetSecret(context context.Context, client k8sclient.Reader, name string, namespace string) (*corev1.Secret, error) {
	key := k8sclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	answer := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := client.Get(context, key, &answer); err != nil {
		return nil, err
	}

	return &answer, nil
}
