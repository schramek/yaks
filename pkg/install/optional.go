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

package install

import (
	"context"

	"github.com/citrusframework/yaks/pkg/client"
	"github.com/go-logr/logr"
)

// OperatorStartupOptionalTools tries to install optional tools at operator startup and warns if something goes wrong.
func OperatorStartupOptionalTools(ctx context.Context, c client.Client, log logr.Logger) {

	// Try to register the OpenShift CLI Download link if possible
	if err := OpenShiftConsoleDownloadLink(ctx, c); err != nil {
		log.Info("Cannot install OpenShift CLI download link: skipping.")
		log.V(8).Info("Error while installing OpenShift CLI download link", "error", err)
	}

	// Try to register the cluster role for standard admin and edit users
	if clusterRoleInstalled, err := isClusterRoleInstalled(ctx, c, "yaks-edit"); err != nil {
		log.Info("Cannot detect user cluster role: skipping.")
		log.V(8).Info("Error while getting user cluster role", "error", err)
	} else if !clusterRoleInstalled {
		if err := installResource(ctx, c, nil, "/rbac/user-cluster-role.yaml"); err != nil {
			log.Info("Cannot install user cluster role: skipping.")
			log.V(8).Info("Error while installing user cluster role", "error", err)
		}
	}

}
