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

package org.citrusframework.yaks.openapi;

import javax.net.ssl.HttpsURLConnection;
import javax.net.ssl.SSLContext;
import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;
import java.security.KeyManagementException;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.util.Objects;

import com.consol.citrus.util.FileUtils;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.apicurio.datamodels.Library;
import io.apicurio.datamodels.openapi.models.OasDocument;
import org.apache.http.HttpHeaders;
import org.apache.http.conn.ssl.NoopHostnameVerifier;
import org.apache.http.conn.ssl.TrustAllStrategy;
import org.apache.http.ssl.SSLContexts;
import org.springframework.http.HttpMethod;
import org.springframework.http.MediaType;
import org.yaml.snakeyaml.Yaml;
import org.yaml.snakeyaml.constructor.SafeConstructor;

/**
 * Loads Open API specifications from different locations like file resource or web resource.
 * @author Christoph Deppisch
 */
public final class OpenApiResourceLoader {

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();
    private static final Yaml YAML_PARSER = new Yaml(new SafeConstructor());

    /**
     * Prevent instantiation of utility class.
     */
    private OpenApiResourceLoader() {
        super();
    }

    /**
     * Loads the specification from a file resource. Either classpath or file system resource path is supported.
     * @param resource
     * @return
     */
    public static OasDocument fromFile(String resource) {
        try {
            return resolve(FileUtils.readToString(FileUtils.getFileResource(resource)));
        } catch (IOException e) {
            throw new IllegalStateException("Failed to parse Open API specification: " + resource, e);
        }
    }

    /**
     * Loads specification from given web URL location.
     * @param url
     * @return
     */
    public static OasDocument fromWebResource(URL url) {
        HttpURLConnection con = null;
        try {
            con = (HttpURLConnection) url.openConnection();
            con.setRequestMethod(HttpMethod.GET.name());
            con.setRequestProperty(HttpHeaders.ACCEPT, MediaType.APPLICATION_JSON_VALUE);

            int status = con.getResponseCode();
            if (status > 299) {
                throw new IllegalStateException("Failed to retrieve Open API specification: " + url.toString(),
                        new IOException(FileUtils.readToString(con.getErrorStream())));
            } else {
                return resolve(FileUtils.readToString(con.getInputStream()));
            }
        } catch (IOException e) {
            throw new IllegalStateException("Failed to retrieve Open API specification: " + url.toString(), e);
        } finally {
            if (con != null) {
                con.disconnect();
            }
        }
    }

    /**
     * Loads specification from given web URL location using secured Http connection.
     * @param url
     * @return
     */
    public static OasDocument fromSecuredWebResource(URL url) {
        Objects.requireNonNull(url);

        HttpsURLConnection con = null;
        try {
            SSLContext sslcontext = SSLContexts
                    .custom()
                    .loadTrustMaterial(TrustAllStrategy.INSTANCE)
                    .build();

            HttpsURLConnection.setDefaultSSLSocketFactory(sslcontext.getSocketFactory());
            HttpsURLConnection.setDefaultHostnameVerifier(NoopHostnameVerifier.INSTANCE);

            con = (HttpsURLConnection) url.openConnection();
            con.setRequestMethod(HttpMethod.GET.name());
            con.setRequestProperty(HttpHeaders.ACCEPT, MediaType.APPLICATION_JSON_VALUE);

            int status = con.getResponseCode();
            if (status > 299) {
                throw new IllegalStateException("Failed to retrieve Open API specification: " + url.toString(),
                        new IOException(FileUtils.readToString(con.getErrorStream())));
            } else {
                return resolve(FileUtils.readToString(con.getInputStream()));
            }
        } catch (NoSuchAlgorithmException | KeyStoreException | KeyManagementException e) {
            throw new IllegalStateException("Failed to create https client for ssl connection", e);
        } catch (IOException e) {
            throw new IllegalStateException("Failed to retrieve Open API specification: " + url.toString(), e);
        } finally {
            if (con != null) {
                con.disconnect();
            }
        }
    }

    private static OasDocument resolve(String specification) {
        if (isJsonSpec(specification)) {
            return (OasDocument) Library.readDocumentFromJSONString(specification);
        }

        final JsonNode node = OBJECT_MAPPER.convertValue(YAML_PARSER.load(specification), JsonNode.class);
        return (OasDocument) Library.readDocument(node);
    }

    private static boolean isJsonSpec(final String specification) {
        return specification.trim().startsWith("{");
    }
}
