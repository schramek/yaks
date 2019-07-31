Feature: Http client

  Background:
    Given URL: http://localhost:8080

  Scenario: GET
    When send GET /todo
    Then verify HTTP response body: {"id": "@ignore@", "task": "Sample task", "completed": 0}
    And receive HTTP 200 OK

  Scenario: POST
    Given variable id is "citrus:randomNumber(5)"
    Given HTTP request body
    """
    {"id": "${id}", "task": "New task", "completed": 0}
    """
    When send POST /todo/${id}
    Then receive HTTP 201 CREATED

  Scenario: DELETE
    Given variable id is "citrus:randomNumber(5)"
    When send DELETE /todo/${id}
    Then receive HTTP 204 NO_CONTENT

  Scenario: PUT
    Given variable id is "citrus:randomNumber(5)"
    Given HTTP request body
    """
    {"id": "${id}", "task": "Task update", "completed": 0}
    """
    When send PUT /todo/${id}
    And verify HTTP response body
    """
    {"id": "${id}", "task": "Task update", "completed": 0}
    """
    Then receive HTTP 200 OK

  Scenario: Request header
    Given HTTP request header Accept is "application/json"
    Given HTTP request header Accept-Encoding="gzip"
    When send GET /todo
    Then receive HTTP 200 OK

  Scenario: Request headers
    Given HTTP request headers
      | Accept          | application/json |
      | Accept-Encoding | gzip |
    When send GET /todo
    Then receive HTTP 200 OK

  Scenario: Verify response header
    When send GET /todo
    Then verify HTTP response header X-TodoId is "@isNumber()@"
    And verify HTTP response header Content-Type="application/json"
    And receive HTTP 200 OK

  Scenario: Verify response headers
    When send GET /todo
    Then verify HTTP response headers
      | X-TodoId      | @isNumber()@ |
      | Content-Type  | application/json |
    And receive HTTP 200 OK

  Scenario: Verify multiline response body
    When send GET /todo
    Then verify HTTP response body
    """
    {
      "id": "@ignore@",
      "task": "Sample task",
      "completed": 0
    }
    """
    And receive HTTP 200 OK
