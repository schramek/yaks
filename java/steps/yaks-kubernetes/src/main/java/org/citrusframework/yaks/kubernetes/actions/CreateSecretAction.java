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

package org.citrusframework.yaks.kubernetes.actions;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Optional;

import com.consol.citrus.context.TestContext;
import com.consol.citrus.exceptions.CitrusRuntimeException;
import com.consol.citrus.util.FileUtils;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.SecretBuilder;
import org.springframework.core.io.Resource;

/**
 * @author Christoph Deppisch
 */
public class CreateSecretAction extends AbstractKubernetesAction implements KubernetesAction {

    private final String secretName;
    private final String filePath;
    private final Map<String, String> properties;

    public CreateSecretAction(Builder builder) {
        super("create-secret", builder);

        this.secretName = builder.secretName;
        this.filePath = builder.filePath;
        this.properties = builder.properties;
    }

    @Override
    public void doExecute(TestContext context) {
        Map<String, String> data = new LinkedHashMap<>();
        if (filePath != null) {
            try {
                Resource file = FileUtils.getFileResource(context.replaceDynamicContentInString(filePath), context);
                String resolvedFileContent = context.replaceDynamicContentInString(FileUtils.readToString(file, StandardCharsets.UTF_8));
                data.put(Optional.ofNullable(file.getFilename()).orElse("application.properties"),
                        Base64.getEncoder().encodeToString(resolvedFileContent.getBytes(StandardCharsets.UTF_8)));
            } catch (IOException e) {
                throw new CitrusRuntimeException("Failed to read properties file", e);
            }
        }

        context.resolveDynamicValuesInMap(properties)
                .forEach((k, v) -> data.put(k, Base64.getEncoder().encodeToString(v.getBytes(StandardCharsets.UTF_8))));

        Secret secret = new SecretBuilder()
                .withNewMetadata()
                    .withNamespace(namespace(context))
                    .withName(context.replaceDynamicContentInString(secretName))
                .endMetadata()
                .withType("generic")
                .withData(data)
                .build();

        getKubernetesClient().secrets()
                .inNamespace(namespace(context))
                .createOrReplace(secret);
    }

    /**
     * Action builder.
     */
    public static class Builder extends AbstractKubernetesAction.Builder<CreateSecretAction, Builder> {

        private String secretName;
        private String filePath;
        private final Map<String, String> properties = new HashMap<>();

        public Builder name(String secretName) {
            this.secretName = secretName;
            return this;
        }

        public Builder fromFile(String filePath) {
            this.filePath = filePath;
            return this;
        }

        public Builder properties(Map<String, String> properties) {
            this.properties.putAll(properties);
            return this;
        }

        @Override
        public CreateSecretAction build() {
            return new CreateSecretAction(this);
        }
    }
}
