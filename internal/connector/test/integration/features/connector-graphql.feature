Feature: create a a connector
  In order to use connectors api
  As an API user
  I need to be able to manage connectors

  Background:
    Given the path prefix is "/api/connector_mgmt"
    Given a user named "Larry"
    Given a user named "Evil Bob"

  Scenario: Anyone can get the schema.graphql
    When I GET path "/v1/schema.graphql"
    Then the response code should be 200

  Scenario: Larry lists all connector types
    Given I am logged in as "Larry"
    When I POST to "/v1/graphql" a GraphQL query:
      """
      query {
        listConnectorTypes {
          size
          page
          total
          items {
            id
            name
            version
          }
        }
      }
      """

    Then the response code should be 200
    And the response should match json:
      """
      {
        "data": {
          "listConnectorTypes": {
            "items": [
              {
                "id": "aws-sqs-source-v1alpha1",
                "name": "aws-sqs-source",
                "version": "v1alpha1"
              },
              {
                "id": "log_sink_0.1",
                "name": "Log Sink",
                "version": "0.1"
              }
            ],
            "page": 1,
            "size": 2,
            "total": 2
          }
        }
      }
      """

  Scenario: Larry tries to create a connector with an invalid configuration spec
    Given I am logged in as "Larry"
    Given I store json as ${input}:
      """
        {
          "kind": "Connector",
          "metadata": {
            "name": "example 1",
            "kafka_id":"mykafka"
          },
          "deployment_location": {
            "kind": "addon",
            "cluster_id": "default"
          },
          "kafka": {
            "bootstrap_server": "kafka.hostname",
            "client_id": "myclient",
            "client_secret": "test"
          },
          "connector_type_id": "aws-sqs-source-v1alpha1",
          "connector_spec": "{}"
        }
      """

    When I POST to "/v1/graphql" a GraphQL query:
      """
      mutation createConnector($input: ConnectorInput!) {
          createConnector(async: true, body: $input) {
              id
          }
      }
      """

    Then the response code should be 200
    And the response should match json:
      """
      {
        "data": {},
        "errors": [
          {
            "extensions": {
              "status": 400,
              "response": {
                "code": "CONNECTOR-MGMT-21",
                "href": "/api/connector_mgmt/v1/errors/21",
                "id": "21",
                "kind": "Error",
                "operation_id": "${response.errors[0].extensions.response.operation_id}",
                "reason": "connector spec not conform to the connector type schema. 4 errors encountered.  1st error: (root): queueNameOrArn is required"
              }
            },
            "message": "http response status code: 400",
            "path": [
              "createConnector"
            ]
          }
        ]
      }
      """

  Scenario: Larry creates connectors creates and lists connectors that
  Evil Bob can't access.

    Given I am logged in as "Larry"

    Given I store json as ${input}:
      """
      {
        "kind": "Connector",
        "metadata": {
          "name": "example 1",
          "kafka_id":"mykafka"
        },
        "deployment_location": {
          "kind": "addon",
          "cluster_id": "default"
        },
        "connector_type_id": "aws-sqs-source-v1alpha1",
        "kafka": {
          "bootstrap_server": "kafka.hostname",
          "client_id": "myclient",
          "client_secret": "test"
        },
        "connector_spec": "{\"queueNameOrArn\": \"test\",\"accessKey\": \"test\",\"secretKey\": \"test\",\"region\": \"east\"}"
      }
      """

    When I POST to "/v1/graphql" a GraphQL query:
      """
      mutation createConnector($input: ConnectorInput!) {
          connector1: createConnector(async: true, body: $input) {
              status
          }
          connector2: createConnector(async: true, body: $input) {
              status
          }
      }
      """

    Then the response code should be 200
    And the response should match json:
      """
      {
        "data": {
          "connector1": {
            "status": "assigning"
          },
          "connector2": {
            "status": "assigning"
          }
        }
      }
      """

    When I POST to "/v1/graphql" a GraphQL query:
      """
      query {
          listConnectors {
            total
            items {
              channel
              status
              connector_type {
                description
              }
            }
          }
      }
      """

    Then the response code should be 200
    And the response should match json:
      """
      {
        "data": {
          "listConnectors": {
            "items": [
              {
                "channel": "stable",
                "connector_type": {
                  "description": "Receive data from AWS SQS"
                },
                "status": "assigning"
              },
              {
                "channel": "stable",
                "connector_type": {
                  "description": "Receive data from AWS SQS"
                },
                "status": "assigning"
              }
            ],
            "total": 2
          }
        }
      }
      """

    Given I am logged in as "Evil Bob"
    When I POST to "/v1/graphql" a GraphQL query:
      """
      query {
          listConnectors {
            total
            items {
              channel
              status
            }
          }
      }
      """

    Then the response code should be 200
    And the response should match json:
      """
      {
        "data": {
          "listConnectors": {
            "items": null,
            "total": 0
          }
        }
      }
      """