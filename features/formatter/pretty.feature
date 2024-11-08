
Feature: pretty formatter
  Smoke test of pretty formatter.
  Comprehensive tests at internal/formatters.

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
    9.99s

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
    9.99s

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
    9.99s

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
    9.99s

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
        Given passing step      # functional_test.go:0 -> SuiteContext.func2
        Then a failing step     # functional_test.go:1 -> *godogFeaturesScenarioInner
        intentional failure

    --- Failed steps:

      Scenario: simple scenario # features/simple.feature:4
        Then a failing step # features/simple.feature:8
          Error: intentional failure


    1 scenarios (1 failed)
    2 steps (1 passed, 1 failed)
    9.99s

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
        Given <status> step             # functional_test.go:1 -> SuiteContext.func2

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
    9.99s

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
    9.99s

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
        Given passing step      # functional_test.go:0 -> SuiteContext.func2
          \"\"\"  content type
          step doc string
          \"\"\"

    1 scenarios (1 passed)
    1 steps (1 passed)
    9.99s

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
        Given passing step       # functional_test.go:0 -> SuiteContext.func2
        And pending step         # functional_test.go:0 -> SuiteContext.func1
          TODO: write pending definition
        And undefined doc string
          \"\"\"
          abc
          \"\"\"
        And undefined table
          | a | b | c |
          | 1 | 2 | 3 |
        And passing step         # functional_test.go:0 -> SuiteContext.func2

    1 scenarios (1 pending, 0 undefined)
    5 steps (1 passed, 1 pending, 2 undefined, 1 skipped)
    9.99s

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


  Scenario: Should scenarios identified with path:line and preserve the order.
    Given a feature file at "features/simple1.feature":
    """
        Feature: feature 1
          Scenario: scenario 1a
            Given passing step
          Scenario: scenario 1b
            Given passing step
    """
    And a feature file at "features/simple2.feature":
        """
        Feature: feature 2
          Scenario: scenario 2a
            Given passing step
        """
    And a feature file at "features/simple3.feature":
        """
        Feature: feature 3
          Scenario: scenario 3a
            Given passing step
        """
    Given a feature path "features/simple2.feature:2"
    Given a feature path "features/simple1.feature:4"
    Given a feature path "features/simple3.feature:2"
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: feature 2

      Scenario: scenario 2a # features/simple2.feature:2
        Given passing step  # <gofile>:<lineno> -> <gofunc>

    Feature: feature 1

      Scenario: scenario 1b # features/simple1.feature:4
        Given passing step  # <gofile>:<lineno> -> <gofunc>

    Feature: feature 3

      Scenario: scenario 3a # features/simple3.feature:2
        Given passing step  # <gofile>:<lineno> -> <gofunc>

    3 scenarios (3 passed)
    3 steps (3 passed)
    9.99s

    """


  Scenario: Support of Feature Plus Rule
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule
            simple feature description
         Rule: simple rule
             simple rule description
         Example: simple scenario
            simple scenario description
          Given passing step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: simple feature with a rule
      simple feature description

      Example: simple scenario # features/simple.feature:5
        Given passing step     # functional_test.go:0 -> SuiteContext.func2

    1 scenarios (1 passed)
    1 steps (1 passed)
    9.99s

    """

  Scenario: Support of Feature Plus Rule with Background
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule with Background
            simple feature description
         Rule: simple rule
             simple rule description
         Background:
             Given passing step
         Example: simple scenario
            simple scenario description
          Given passing step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: simple feature with a rule with Background
      simple feature description

      Background:
        Given passing step     # functional_test.go:0 -> SuiteContext.func2

      Example: simple scenario # features/simple.feature:7
        Given passing step     # functional_test.go:0 -> SuiteContext.func2

    1 scenarios (1 passed)
    2 steps (2 passed)
    9.99s

    """

  Scenario: Support of Feature Plus Rule with Scenario Outline
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule with Scenario Outline
            simple feature description
         Rule: simple rule
             simple rule description
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
    Feature: simple feature with a rule with Scenario Outline
      simple feature description

      Scenario Outline: simple scenario # features/simple.feature:5
        Given <status> step             # functional_test.go:0 -> SuiteContext.func2

        Examples: simple examples
          | status  |
          | passing |
          | failing |
            intentional failure

    --- Failed steps:

      Scenario Outline: simple scenario # features/simple.feature:5
        Given failing step # features/simple.feature:8
          Error: intentional failure


    2 scenarios (1 passed, 1 failed)
    2 steps (1 passed, 1 failed)
    9.99s

    """

  Scenario: Use 'given' keyword on a declared 'when' step
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule
            simple feature description
         Rule: simple rule
             simple rule description
         Example: simple scenario
            simple scenario description
          Given a when step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: simple feature with a rule
      simple feature description

      Example: simple scenario # features/simple.feature:5
        Given a when step

    1 scenarios (1 undefined)
    1 steps (1 undefined)
    9.99s

    You can implement step definitions for undefined steps with these snippets:

    func aWhenStep() error {
    	return godog.ErrPending
    }

    func InitializeScenario(ctx *godog.ScenarioContext) {
    	ctx.Step(`^a when step$`, aWhenStep)
    }


    """

  Scenario: Use 'when' keyword on a declared 'then' step
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule
            simple feature description
         Rule: simple rule
             simple rule description
         Example: simple scenario
            simple scenario description
          When a then step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: simple feature with a rule
      simple feature description

      Example: simple scenario # features/simple.feature:5
        When a then step

    1 scenarios (1 undefined)
    1 steps (1 undefined)
    9.99s

    You can implement step definitions for undefined steps with these snippets:

    func aThenStep() error {
    	return godog.ErrPending
    }

    func InitializeScenario(ctx *godog.ScenarioContext) {
    	ctx.Step(`^a then step$`, aThenStep)
    }


    """

  Scenario: Use 'then' keyword on a declared 'given' step
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule
            simple feature description
         Rule: simple rule
             simple rule description
         Example: simple scenario
            simple scenario description
          Then a given step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: simple feature with a rule
      simple feature description

      Example: simple scenario # features/simple.feature:5
        Then a given step

    1 scenarios (1 undefined)
    1 steps (1 undefined)
    9.99s

    You can implement step definitions for undefined steps with these snippets:

    func aGivenStep() error {
    	return godog.ErrPending
    }

    func InitializeScenario(ctx *godog.ScenarioContext) {
    	ctx.Step(`^a given step$`, aGivenStep)
    }


    """

  Scenario: Match keyword functions correctly
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature with a rule
            simple feature description
         Rule: simple rule
             simple rule description
         Example: simple scenario
            simple scenario description
          Given a given step
          When a when step
          Then a then step
          And a then step
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
    Feature: simple feature with a rule
      simple feature description

      Example: simple scenario # features/simple.feature:5
        Given a given step     # functional_test.go:0 -> InitializeScenario.func3
        When a when step       # functional_test.go:0 -> InitializeScenario.func4
        Then a then step       # functional_test.go:0 -> InitializeScenario.func5
        And a then step        # functional_test.go:0 -> InitializeScenario.func5

    1 scenarios (1 passed)
    4 steps (4 passed)
    9.99s

    """