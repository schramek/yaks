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

package org.citrusframework.yaks.kafka;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

import com.consol.citrus.Citrus;
import com.consol.citrus.TestCaseRunner;
import com.consol.citrus.annotations.CitrusFramework;
import com.consol.citrus.annotations.CitrusResource;
import com.consol.citrus.context.TestContext;
import com.consol.citrus.exceptions.CitrusRuntimeException;
import com.consol.citrus.kafka.endpoint.KafkaEndpoint;
import com.consol.citrus.kafka.endpoint.KafkaEndpointBuilder;
import com.consol.citrus.kafka.message.KafkaMessage;
import com.consol.citrus.kafka.message.KafkaMessageHeaders;
import com.consol.citrus.message.Message;
import com.consol.citrus.util.FileUtils;
import io.cucumber.datatable.DataTable;
import io.cucumber.java.Before;
import io.cucumber.java.Scenario;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;

import static com.consol.citrus.actions.ReceiveMessageAction.Builder.receive;
import static com.consol.citrus.actions.SendMessageAction.Builder.send;

public class KafkaSteps {

    @CitrusResource
    private TestCaseRunner runner;

    @CitrusResource
    private TestContext context;

    @CitrusFramework
    private Citrus citrus;

    private Map<String, Object> headers = new HashMap<>();
    private String body;

    private KafkaEndpoint kafkaEndpoint;

    private String messageKey;
    private Integer partition;
    private String topic = "test";

    private String endpointName = KafkaSettings.getEndpointName();

    private long timeout = KafkaSettings.getConsumerTimeout();

    @Before
    public void before(Scenario scenario) {
        if (kafkaEndpoint == null) {
            if (citrus.getCitrusContext().getReferenceResolver().resolveAll(KafkaEndpoint.class).size() == 1L) {
                kafkaEndpoint = citrus.getCitrusContext().getReferenceResolver().resolve(KafkaEndpoint.class);
            } else if (citrus.getCitrusContext().getReferenceResolver().isResolvable(endpointName)) {
                kafkaEndpoint = citrus.getCitrusContext().getReferenceResolver().resolve(endpointName, KafkaEndpoint.class);
            } else {
                kafkaEndpoint = new KafkaEndpointBuilder().build();
                citrus.getCitrusContext().getReferenceResolver().bind(endpointName, kafkaEndpoint);
            }
        }

        headers = new HashMap<>();
        body = null;

        messageKey = null;
        partition = null;
    }

    @Given("^(?:Kafka|kafka) connection$")
    public void setConnection(DataTable properties) {
        Map<String, String> connectionProps = properties.asMap(String.class, String.class);

        String url = connectionProps.getOrDefault("url", "localhost:9092");
        String topicName = connectionProps.getOrDefault("topic", this.topic);
        String consumerGroup = connectionProps.getOrDefault("consumerGroup", KafkaMessageHeaders.KAFKA_PREFIX + "group");
        String offsetReset = connectionProps.getOrDefault("offsetReset", "earliest");

        setTopic(context.replaceDynamicContentInString(topicName));
        kafkaEndpoint.getEndpointConfiguration().setServer(context.replaceDynamicContentInString(url));
        kafkaEndpoint.getEndpointConfiguration().setOffsetReset(context.replaceDynamicContentInString(offsetReset));
        kafkaEndpoint.getEndpointConfiguration().setConsumerGroup(context.replaceDynamicContentInString(consumerGroup));
    }

    @Given("^(?:Kafka|kafka) producer configuration$")
    public void setProducerConfig(DataTable properties) {
        Map<String, Object> producerProperties = properties.asMap(String.class, Object.class);
        kafkaEndpoint.getEndpointConfiguration().setProducerProperties(producerProperties);
    }

    @Given("^(?:Kafka|kafka) consumer configuration$")
    public void setConsumerConfig(DataTable properties) {
        Map<String, Object> consumerProperties = properties.asMap(String.class, Object.class);
        kafkaEndpoint.getEndpointConfiguration().setConsumerProperties(consumerProperties);
    }

    @Given("^(?:Kafka|kafka) endpoint \"([^\"\\s]+)\"$")
    public void setServer(String name) {
        this.endpointName = name;
        if (citrus.getCitrusContext().getReferenceResolver().isResolvable(name)) {
            kafkaEndpoint = citrus.getCitrusContext().getReferenceResolver().resolve(name, KafkaEndpoint.class);
        } else if (kafkaEndpoint != null) {
            citrus.getCitrusContext().getReferenceResolver().bind(endpointName, kafkaEndpoint);
            kafkaEndpoint.setName(endpointName);
        }
    }

    @Given("^(?:Kafka|kafka) message key: (.+)$")
    public void setMessageKey(String key) {
        this.messageKey = key;
    }

    @Given("^(?:Kafka|kafka) consumer timeout is (\\d+)(?: ms| milliseconds)$")
    public void setConsumerTimeout(int milliseconds) {
        this.timeout = milliseconds;
    }

    @Given("^(?:Kafka|kafka) topic partition: (\\d+)$")
    public void setPartition(int partition) {
        this.partition = partition;
    }

    @Given("^(?:Kafka|kafka) topic: (.+)$")
    public void setTopic(String topicName) {
        this.topic = topicName;
        kafkaEndpoint.getEndpointConfiguration().setTopic(topicName);
    }

    @Given("^(?:Kafka|kafka) message header ([^\\s]+)(?:=| is )\"(.+)\"$")
    @Then("^(?:expect|verify) (?:Kafka|kafka) message header ([^\\s]+)(?:=| is )\"(.+)\"$")
    public void addMessageHeader(String name, Object value) {
        headers.put(name, value);
    }

    @Given("^(?:Kafka|kafka) message headers$")
    public void addMessageHeaders(DataTable headers) {
        Map<String, Object> headerPairs = headers.asMap(String.class, Object.class);
        headerPairs.forEach(this::addMessageHeader);
    }

    @Given("^(?:Kafka|kafka) message body$")
    @Then("^(?:expect|verify) (?:Kafka|kafka) message body$")
    public void setMessageBodyMultiline(String body) {
        setMessageBody(body);
    }

    @Given("^load (?:Kafka|kafka) message body ([^\\s]+)$")
    @Given("^(?:expect|verify) (?:Kafka|kafka) message body loaded from ([^\\s]+)$")
    public void loadMessageBody(String file) {
        try {
            setMessageBody(FileUtils.readToString(FileUtils.getFileResource(file)));
        } catch (IOException e) {
            throw new CitrusRuntimeException(String.format("Failed to load body from file resource %s", file));
        }
    }

    @Given("^(?:Kafka|kafka) message body: (.+)$")
    @Then("^(?:expect|verify) (?:Kafka|kafka) message body: (.+)$")
    public void setMessageBody(String body) {
        this.body = body;
    }

    @When("^send (?:Kafka|kafka) message$")
    public void sendMessage() {
        runner.run(send().endpoint(kafkaEndpoint)
                .message(createKafkaMessage()));

        body = null;
        headers.clear();
    }

    @Then("^receive (?:Kafka|kafka) message$")
    public void receiveMessage() {
        runner.run(receive().endpoint(kafkaEndpoint)
                .timeout(timeout)
                .message(createKafkaMessage()));

        body = null;
        headers.clear();
    }

    @When("^send (?:Kafka|kafka) message to topic (.+)$")
    public void sendMessage(String topicName) {
        setTopic(topicName);
        sendMessage();
    }

    @Then("^receive (?:Kafka|kafka) message on topic (.+)")
    public void receiveMessage(String topicName) {
        setTopic(topicName);
        receiveMessage();
    }

    @When("^send (?:Kafka|kafka) message with body and headers: (.+)$")
    @Given("^message in (?:Kafka|kafka) with body and headers: (.+)$")
    public void sendMessageBodyAndHeaders(String body, DataTable headers) {
        setMessageBody(body);
        addMessageHeaders(headers);
        sendMessage();
    }

    @When("^send (?:Kafka|kafka) message with body: (.+)$")
    @Given("^message in (?:Kafka|kafka) with body: (.+)$")
    public void sendMessageBody(String body) {
        setMessageBody(body);
        sendMessage();
    }

    @When("^send (?:Kafka|kafka) message with body$")
    @Given("^message in (?:Kafka|kafka) with body$")
    public void sendMessageBodyMultiline(String body) {
        sendMessageBody(body);
    }

    @Then("^(?:receive|expect|verify) (?:Kafka|kafka) message with body and headers: (.+)$")
    public void receiveFromKafka(String body, DataTable headers) {
        setMessageBody(body);
        addMessageHeaders(headers);
        receiveMessage();
    }

    @Then("^(?:receive|expect|verify) (?:Kafka|kafka) message with body: (.+)$")
    public void receiveMessageBody(String body) {
        setMessageBody(body);
        receiveMessage();
    }

    @Then("^(?:receive|expect|verify) (?:Kafka|kafka) message with body$")
    public void receiveMessageBodyMultiline(String body) {
        receiveMessageBody(body);
    }

    private Message createKafkaMessage() {
        KafkaMessage message = new KafkaMessage(body, headers)
                .topic(topic);

        if (messageKey != null) {
            message.messageKey(messageKey);
        }

        if (partition != null) {
            message.partition(partition);
        }
        return message;
    }

}
