/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.citrusframework.yaks.camelk.actions;

import com.consol.citrus.TestAction;
import com.consol.citrus.context.TestContext;
import io.fabric8.kubernetes.client.KubernetesClient;
import org.citrusframework.yaks.camelk.VariableNames;
import org.citrusframework.yaks.kubernetes.KubernetesSettings;

/**
 * Base action provides access to Knative properties such as broker name. These properties are read from
 * environment settings or explicitly set as part of the test case and get stored as test variables in the current context.
 * This base class gives convenient access to the test variables and provides a fallback if no variable is set.
 *
 * @author Christoph Deppisch
 */
public interface CamelKAction extends TestAction {

    /**
     * Gets the Kubernetes client.
     * @return
     */
    KubernetesClient getKubernetesClient();

    /**
     * Resolves namespace name from given test context using the stored test variable.
     * Fallback to the namespace given in Kubernetes environment settings when no test variable is present.
     *
     * @param context
     * @return
     */
    default String namespace(TestContext context) {
        if (context.getVariables().containsKey(VariableNames.CAMEL_K_NAMESPACE.value())) {
            return context.getVariable(VariableNames.CAMEL_K_NAMESPACE.value());
        }

        return KubernetesSettings.getNamespace();
    }
}

