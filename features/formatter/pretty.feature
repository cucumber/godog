Feature: pretty formatter
  In order to support tools that import pretty output
  I need to be able to support pretty formatted output

  Scenario: Support of Feature Plus Scenario Node
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:3

      1 scenarios (1 undefined)
      No steps
      0s
    """

  Scenario: Support of Feature Plus Background and Scenario Node
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Background: simple background
            Given passing step
            And passing step
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Background: simple background
            Given passing step
            And passing step

        Scenario: simple scenario # features/simple.feature:6

      1 scenarios (1 undefined)
      No steps
      0s
    """

  Scenario: Support of Feature Plus Rule Node
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Rule: simple rule
            Scenario: simple scenario
                simple scenario description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Rule: simple rule

          Scenario: simple scenario # features/simple.feature:4

      1 scenarios (1 undefined)
      No steps
      0s
    """

  Scenario: Support of Feature Plus Rule Node with Background
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Rule: simple rule

        Background: simple background
            Given something
            And another thing

        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Rule: simple rule

        Background: simple background
            Given something
            And another thing

          Scenario: simple scenario # features/simple.feature:10

      1 scenarios (1 undefined)
      No steps
      0s
    """

  Scenario: Support of Feature Plus Rule Node with multiple scenarios
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Rule: simple rule

        Scenario: simple scenario
            simple scenario description

            Given passing step
            Then passing step

        Scenario: simple second scenario
            simple second scenario description

            Given passing step
            Then passing step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Rule: simple rule

          Scenario: simple scenario # features/simple.feature:6
            Given passing step      # suite_context_test.go:0 -> InitializeScenario.func2
            Then passing step       # suite_context_test.go:0 -> InitializeScenario.func2

          Scenario: simple second scenario # features/simple.feature:12
            Given passing step             # suite_context_test.go:0 -> InitializeScenario.func2
            Then passing step              # suite_context_test.go:0 -> InitializeScenario.func2

      2 scenarios (2 passed)
      4 steps (4 passed)
      0s
    """

  Scenario: Support of Feature Plus Rule Node with Background and multiple scenarios
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Rule: simple rule

        Background: simple background
            Given passing step
            And passing step

        Scenario: simple scenario
            simple scenario description

            Given passing step
            Then passing step

        Scenario: simple second scenario
            simple second scenario description

            Given passing step
            Then passing step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Rule: simple rule

        Background: simple background
            Given passing step      # suite_context_test.go:0 -> InitializeScenario.func2
            And passing step        # suite_context_test.go:0 -> InitializeScenario.func2

          Scenario: simple scenario # features/simple.feature:9
            Given passing step      # suite_context_test.go:0 -> InitializeScenario.func2
            Then passing step       # suite_context_test.go:0 -> InitializeScenario.func2

          Scenario: simple second scenario # features/simple.feature:15
            Given passing step             # suite_context_test.go:0 -> InitializeScenario.func2
            Then passing step              # suite_context_test.go:0 -> InitializeScenario.func2

      2 scenarios (2 passed)
      8 steps (8 passed)
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:5

      1 scenarios (1 undefined)
      No steps
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario # features/simple.feature:4

          Examples: simple examples
            | status |
            | pass   |
            | fail   |

      2 scenarios (2 undefined)
      No steps
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario # features/simple.feature:6

          Examples: simple examples
            | status |
            | pass   |
            | fail   |

      2 scenarios (2 undefined)
      No steps
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:4
          Given passing step      # suite_context.go:0 -> SuiteContext.func2
          Then a failing step     # suite_context.go:0 -> *suiteContext
          intentional failure

      --- Failed steps:

        Scenario: simple scenario # features/simple.feature:4
          Then a failing step # features/simple.feature:8
            Error: intentional failure


      1 scenarios (1 failed)
      2 steps (1 passed, 1 failed)
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario # features/simple.feature:4
          Given <status> step             # suite_context.go:0 -> SuiteContext.func2

          Examples: simple examples
            | status  |
            | passing |
            | failing |
              intentional failure

      --- Failed steps:

        Scenario Outline: simple scenario # features/simple.feature:4
          Given failing step # features/simple.feature:7
            Error: intentional failure


      2 scenarios (1 passed, 1 failed)
      2 steps (1 passed, 1 failed)
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple description

        Scenario: simple scenario # features/simple.feature:5

      1 scenarios (1 undefined)
      No steps
      0s
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
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple description

        Scenario: simple scenario # features/simple.feature:4
          Given passing step      # suite_context.go:0 -> SuiteContext.func2
            \"\"\"  content type
            step doc string
            \"\"\"

      1 scenarios (1 passed)
      1 steps (1 passed)
      0s
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
        And undefined doc string
        \"\"\"
        abc
        \"\"\"
        And undefined table
        | a | b | c |
        | 1 | 2 | 3 |
        And passing step

    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario  # features/simple.feature:4
          Given passing step       # suite_context.go:0 -> SuiteContext.func2
          And pending step         # suite_context.go:0 -> SuiteContext.func1
            TODO: write pending definition
          And undefined doc string
          \"\"\"
          abc
          \"\"\"
          And undefined table
          | a | b | c |
          | 1 | 2 | 3 |
          And passing step         # suite_context.go:0 -> SuiteContext.func2

      1 scenarios (1 pending, 1 undefined)
      5 steps (1 passed, 1 pending, 2 undefined, 1 skipped)
      0s

      You can implement step definitions for undefined steps with these snippets:

      func undefinedDocString(arg1 *godog.DocString) error {
        return godog.ErrPending
      }

      func undefinedTable(arg1 *godog.Table) error {
        return godog.ErrPending
      }

      func InitializeScenario(ctx *godog.ScenarioContext) {
        ctx.Step(`^undefined doc string$`, undefinedDocString)
        ctx.Step(`^undefined table$`, undefinedTable)
      }
    """

  # Ensure s will not break when injecting data from BeforeStep
  Scenario: Support data injection in BeforeStep
    Given a feature "features/inject.feature" file:
    """
      Feature: inject long value

      Scenario: test scenario
        Given Ignore I save some value X under key Y
        And I allow variable injection
        When Ignore I use value {{Y}}
        Then Ignore Godog rendering should not break
        And Ignore test
          | key | val |
          | 1   | 2   |
          | 3   | 4   |
        And I disable variable injection
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: inject long value

        Scenario: test scenario                        # features/inject.feature:3
          Given Ignore I save some value X under key Y # suite_context.go:0 -> SuiteContext.func7
          And I allow variable injection               # suite_context.go:0 -> *suiteContext
          When Ignore I use value someverylonginjectionsoweacanbesureitsurpasstheinitiallongeststeplenghtanditwillhelptestsmethodsafety # suite_context.go:0 -> SuiteContext.func7
          Then Ignore Godog rendering should not break # suite_context.go:0 -> SuiteContext.func7
          And Ignore test                              # suite_context.go:0 -> SuiteContext.func7
            | key | val |
            | 1   | 2   |
            | 3   | 4   |
          And I disable variable injection             # suite_context.go:0 -> *suiteContext

      1 scenarios (1 passed)
      6 steps (6 passed)
      0s
    """

  Scenario: Should scenarios identified with path:line and preserve the order.
    Given a feature path "features/load.feature:6"
    And a feature path "features/multistep.feature:6"
    And a feature path "features/load.feature:26"
    And a feature path "features/multistep.feature:23"
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: load features
      In order to run features
      As a test suite
      I need to be able to load features

      Scenario: load features within path    # features/load.feature:6
        Given a feature path "features"      # suite_context_test.go:0 -> *godogFeaturesScenario
        When I parse features                # suite_context_test.go:0 -> *godogFeaturesScenario
        Then I should have 13 feature files: # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          features/background.feature
          features/events.feature
          features/formatter/cucumber.feature
          features/formatter/events.feature
          features/formatter/junit.feature
          features/formatter/pretty.feature
          features/lang.feature
          features/load.feature
          features/multistep.feature
          features/outline.feature
          features/run.feature
          features/snippets.feature
          features/tags.feature
          \"\"\"

    Feature: run features with nested steps
      In order to test multisteps
      As a test suite
      I need to be able to execute multisteps

      Scenario: should run passing multistep successfully # features/multistep.feature:6
        Given a feature "normal.feature" file:            # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          Feature: normal feature

            Scenario: run passing multistep
              Given passing step
              Then passing multistep
          \"\"\"
        When I run feature suite                          # suite_context_test.go:0 -> *godogFeaturesScenario
        Then the suite should have passed                 # suite_context_test.go:0 -> *godogFeaturesScenario
        And the following steps should be passed:         # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          passing step
          passing multistep
          \"\"\"

    Feature: load features
      In order to run features
      As a test suite
      I need to be able to load features

      Scenario: load a specific feature file         # features/load.feature:26
        Given a feature path "features/load.feature" # suite_context_test.go:0 -> *godogFeaturesScenario
        When I parse features                        # suite_context_test.go:0 -> *godogFeaturesScenario
        Then I should have 1 feature file:           # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          features/load.feature
          \"\"\"

    Feature: run features with nested steps
      In order to test multisteps
      As a test suite
      I need to be able to execute multisteps

      Scenario: should fail multistep              # features/multistep.feature:23
        Given a feature "failed.feature" file:     # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          Feature: failed feature

            Scenario: run failing multistep
              Given passing step
              When failing multistep
              Then I should have 1 scenario registered
          \"\"\"
        When I run feature suite                   # suite_context_test.go:0 -> *godogFeaturesScenario
        Then the suite should have failed          # suite_context_test.go:0 -> *godogFeaturesScenario
        And the following step should be failed:   # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          failing multistep
          \"\"\"
        And the following steps should be skipped: # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          I should have 1 scenario registered
          \"\"\"
        And the following steps should be passed:  # suite_context_test.go:0 -> *godogFeaturesScenario
          \"\"\"
          passing step
          \"\"\"

    4 scenarios (4 passed)
    16 steps (16 passed)
    0s
    """
