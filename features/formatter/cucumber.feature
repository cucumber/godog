Feature: cucumber json formatter
  In order to support tools that import cucumber json output
  I need to be able to support cucumber json formatted output

  Scenario: Support of Feature Plus Scenario Node
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
      """ application/json
        [
          {
            "uri": "features/simple.feature",
            "id": "simple-feature",
            "keyword": "Feature",
            "name": "simple feature",
            "description": "        simple feature description",
            "line": 1,
            "elements": [
              {
                "id": "simple-feature;simple-scenario",
                "keyword": "Scenario",
                "name": "simple scenario",
                "description": "        simple scenario description",
                "line": 3,
                "type": "scenario"
              }
            ]
          }
        ]
      """

  Scenario: Support of Feature Plus Scenario Node With Tags
    Given a feature "features/simple.feature" file:
    """
        @TAG1
        Feature: simple feature
            simple feature description
        @TAG2 @TAG3
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
      """ application/json
        [
          {
            "uri": "features/simple.feature",
            "id": "simple-feature",
            "keyword": "Feature",
            "name": "simple feature",
            "description": "        simple feature description",
            "line": 2,
            "tags": [
              {
                "name": "@TAG1",
                "line": 1
              }
            ],
            "elements": [
              {
                "id": "simple-feature;simple-scenario",
                "keyword": "Scenario",
                "name": "simple scenario",
                "description": "        simple scenario description",
                "line": 5,
                "type": "scenario",
                "tags": [
                  {
                    "name": "@TAG1",
                    "line": 1
                  },
                  {
                    "name": "@TAG2",
                    "line": 4
                  },
                  {
                    "name": "@TAG3",
                    "line": 4
                  }
                ]
              }
            ]
          }
      ]
      """
  Scenario: Support of Feature Plus Scenario Outline
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Scenario Outline: simple scenario
            simple scenario description

        Examples: simple examples
        | status |
        | pass   |
        | fail   |
    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
      [
        {
          "uri": "features/simple.feature",
          "id": "simple-feature",
          "keyword": "Feature",
          "name": "simple feature",
          "description": "        simple feature description",
          "line": 1,
          "elements": [
            {
              "id": "simple-feature;simple-scenario;simple-examples;2",
              "keyword": "Scenario Outline",
              "name": "simple scenario",
              "description": "        simple scenario description",
              "line": 9,
              "type": "scenario"
            },
            {
              "id": "simple-feature;simple-scenario;simple-examples;3",
              "keyword": "Scenario Outline",
              "name": "simple scenario",
              "description": "        simple scenario description",
              "line": 10,
              "type": "scenario"
            }
          ]
        }
      ]
    """

  Scenario: Support of Feature Plus Scenario Outline With Tags
    Given a feature "features/simple.feature" file:
    """
        @TAG1
        Feature: simple feature
            simple feature description

        @TAG2
        Scenario Outline: simple scenario
            simple scenario description

        @TAG3
        Examples: simple examples
        | status |
        | pass   |
        | fail   |
    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
        [
          {
            "uri": "features/simple.feature",
            "id": "simple-feature",
            "keyword": "Feature",
            "name": "simple feature",
            "description": "        simple feature description",
            "line": 2,
            "tags": [
              {
                "name": "@TAG1",
                "line": 1
              }
            ],
            "elements": [
              {
                "id": "simple-feature;simple-scenario;simple-examples;2",
                "keyword": "Scenario Outline",
                "name": "simple scenario",
                "description": "        simple scenario description",
                "line": 12,
                "type": "scenario",
                "tags": [
                  {
                    "name": "@TAG1",
                    "line": 1
                  },
                  {
                    "name": "@TAG2",
                    "line": 5
                  },
                  {
                    "name": "@TAG3",
                    "line": 9
                  }
                ]
              },
              {
                "id": "simple-feature;simple-scenario;simple-examples;3",
                "keyword": "Scenario Outline",
                "name": "simple scenario",
                "description": "        simple scenario description",
                "line": 13,
                "type": "scenario",
                "tags": [
                  {
                    "name": "@TAG1",
                    "line": 1
                  },
                  {
                    "name": "@TAG2",
                    "line": 5
                  },
                  {
                    "name": "@TAG3",
                    "line": 9
                  }
                ]
              }
            ]
          }
        ]
    """
  Scenario: Support of Feature Plus Scenario With Steps
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Scenario: simple scenario
            simple scenario description

        Given passing step
        Then a failing step

    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
      [
        {
          "uri": "features/simple.feature",
          "id": "simple-feature",
          "keyword": "Feature",
          "name": "simple feature",
          "description": "        simple feature description",
          "line": 1,
          "elements": [
            {
              "id": "simple-feature;simple-scenario",
              "keyword": "Scenario",
              "name": "simple scenario",
              "description": "        simple scenario description",
              "line": 4,
              "type": "scenario",
              "steps": [
                {
                  "keyword": "Given ",
                  "name": "passing step",
                  "line": 7,
                  "match": {
                    "location": "suite_test.go:69"
                  },
                  "result": {
                    "status": "passed",
                    "duration": 0
                  }
                },
                {
                  "keyword": "Then ",
                  "name": "a failing step",
                  "line": 8,
                  "match": {
                    "location": "suite_test.go:52"
                  },
                  "result": {
                    "status": "failed",
                    "error_message": "intentional failure",
                    "duration": 0
                  }
                }
              ]
            }
          ]
        }
      ]
    """
  Scenario: Support of Feature Plus Scenario Outline With Steps
    Given a feature "features/simple.feature" file:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario
        simple scenario description

          Given <status> step

        Examples: simple examples
        | status |
        | passing |
        | failing |

    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
      [
        {
          "uri": "features/simple.feature",
          "id": "simple-feature",
          "keyword": "Feature",
          "name": "simple feature",
          "description": "    simple feature description",
          "line": 1,
          "elements": [
            {
              "id": "simple-feature;simple-scenario;simple-examples;2",
              "keyword": "Scenario Outline",
              "name": "simple scenario",
              "description": "    simple scenario description",
              "line": 11,
              "type": "scenario",
              "steps": [
                {
                  "keyword": "Given ",
                  "name": "passing step",
                  "line": 11,
                  "match": {
                    "location": "suite_test.go:69"
                  },
                  "result": {
                    "status": "passed",
                    "duration": 0
                  }
                }
              ]
            },
            {
              "id": "simple-feature;simple-scenario;simple-examples;3",
              "keyword": "Scenario Outline",
              "name": "simple scenario",
              "description": "    simple scenario description",
              "line": 12,
              "type": "scenario",
              "steps": [
                {
                  "keyword": "Given ",
                  "name": "failing step",
                  "line": 12,
                  "match": {
                    "location": "suite_test.go:52"
                  },
                  "result": {
                    "status": "failed",
                    "error_message": "intentional failure",
                    "duration": 0
                  }
                }
              ]
            }
          ]
        }
      ]
    """

  # Currently godog only supports comments on Feature and not
  # scenario and steps.
  Scenario: Support of Comments
    Given a feature "features/simple.feature" file:
    """
        #Feature comment
        Feature: simple feature
          simple description

          Scenario: simple scenario
          simple feature description
    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
        [
          {
            "uri": "features/simple.feature",
            "id": "simple-feature",
            "keyword": "Feature",
            "name": "simple feature",
            "description": "      simple description",
            "line": 2,
            "comments": [
              {
                "value": "#Feature comment",
                "line": 1
              }
            ],
            "elements": [
              {
                "id": "simple-feature;simple-scenario",
                "keyword": "Scenario",
                "name": "simple scenario",
                "description": "      simple feature description",
                "line": 5,
                "type": "scenario"
              }
            ]
          }
        ]
    """
  Scenario: Support of Docstrings
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
          simple description

          Scenario: simple scenario
          simple feature description

          Given passing step
          \"\"\" content type
          step doc string
          \"\"\"
    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
        [
      {
        "uri": "features/simple.feature",
        "id": "simple-feature",
        "keyword": "Feature",
        "name": "simple feature",
        "description": "      simple description",
        "line": 1,
        "elements": [
          {
            "id": "simple-feature;simple-scenario",
            "keyword": "Scenario",
            "name": "simple scenario",
            "description": "      simple feature description",
            "line": 4,
            "type": "scenario",
            "steps": [
              {
                "keyword": "Given ",
                "name": "passing step",
                "line": 7,
                "doc_string": {
                  "value": "step doc string",
                  "content_type": "content type",
                  "line": 8
                },
                "match": {
                  "location": "suite_test.go:69"
                },
                "result": {
                  "status": "passed",
                  "duration": 0
                }
              }
            ]
          }
        ]
      }
    ]
    """
  Scenario: Support of Undefined, Pending and Skipped status
    Given a feature "features/simple.feature" file:
    """
      Feature: simple feature
      simple feature description

      Scenario: simple scenario
      simple scenario description

        Given passing step
        And pending step
        And undefined
        And passing step

    """
    When I run feature suite with formatter "cucumber"
    Then the rendered json will be as follows:
    """
      [
        {
          "uri": "features/simple.feature",
          "id": "simple-feature",
          "keyword": "Feature",
          "name": "simple feature",
          "description": "  simple feature description",
          "line": 1,
          "elements": [
            {
              "id": "simple-feature;simple-scenario",
              "keyword": "Scenario",
              "name": "simple scenario",
              "description": "  simple scenario description",
              "line": 4,
              "type": "scenario",
              "steps": [
                {
                  "keyword": "Given ",
                  "name": "passing step",
                  "line": 7,
                  "match": {
                    "location": "suite_test.go:69"
                  },
                  "result": {
                    "status": "passed",
                    "duration": 0
                  }
                },
                {
                  "keyword": "And ",
                  "name": "pending step",
                  "line": 8,
                  "match": {
                    "location": "features/simple.feature:8"
                  },
                  "result": {
                    "status": "pending"
                  }
                },
                {
                  "keyword": "And ",
                  "name": "undefined",
                  "line": 9,
                  "match": {
                    "location": "features/simple.feature:9"
                  },
                  "result": {
                    "status": "undefined"
                  }
                },
                {
                  "keyword": "And ",
                  "name": "passing step",
                  "line": 10,
                  "match": {
                    "location": "suite_test.go:69"
                  },
                  "result": {
                    "status": "skipped"
                  }
                }
              ]
            }
          ]
        }
      ]
    """


