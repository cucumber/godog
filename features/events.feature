Feature: suite events
  In order to run tasks before and after important events
  As a test suite
  I need to provide a way to hook into these events

  Background:
    Given I'm listening to suite events

  Scenario: triggers before scenario event
    Given a feature path "features/load.feature:6"
    When I run feature suite
    Then there was event triggered before scenario "load features within path"

  Scenario: triggers appropriate events for a single scenario
    Given a feature path "features/load.feature:6"
    When I run feature suite
    Then these events had to be fired for a number of times:
      | BeforeSuite    | 1 |
      | BeforeScenario | 1 |
      | BeforeStep     | 3 |
      | AfterStep      | 3 |
      | AfterScenario  | 1 |
      | AfterSuite     | 1 |

  Scenario: triggers appropriate events whole feature
    Given a feature path "features/load.feature"
    When I run feature suite
    Then these events had to be fired for a number of times:
      | BeforeSuite    | 1  |
      | BeforeScenario | 6  |
      | BeforeStep     | 19 |
      | AfterStep      | 19 |
      | AfterScenario  | 6  |
      | AfterSuite     | 1  |

  Scenario: triggers appropriate events for two feature files
    Given a feature path "features/load.feature:6"
    And a feature path "features/multistep.feature:6"
    When I run feature suite
    Then these events had to be fired for a number of times:
      | BeforeSuite    | 1 |
      | BeforeScenario | 2 |
      | BeforeStep     | 7 |
      | AfterStep      | 7 |
      | AfterScenario  | 2 |
      | AfterSuite     | 1 |

  Scenario: should not trigger events on empty feature
    Given a feature "normal.feature" file:
      """
      Feature: empty

        Scenario: one

        Scenario: two
      """
    When I run feature suite
    Then these events had to be fired for a number of times:
      | BeforeSuite    | 1 |
      | BeforeScenario | 0 |
      | BeforeStep     | 0 |
      | AfterStep      | 0 |
      | AfterScenario  | 0 |
      | AfterSuite     | 1 |

  Scenario: should not trigger events on empty scenarios
    Given a feature "normal.feature" file:
      """
      Feature: half empty

        Scenario: one

        Scenario: two
          Then passing step
          And adding step state to context
          And having correct context
          And failing step

        Scenario Outline: three
          Then passing step

          Examples:
            | a |
            | 1 |
      """
    When I run feature suite
    Then these events had to be fired for a number of times:
      | BeforeSuite    | 1 |
      | BeforeScenario | 2 |
      | BeforeStep     | 5 |
      | AfterStep      | 5 |
      | AfterScenario  | 2 |
      | AfterSuite     | 1 |

    And the suite should have failed


  Scenario: should add scenario hook errors to steps
    Given a feature "normal.feature" file:
      """
      Feature: scenario hook errors

        Scenario: failing before and after scenario
          Then adding step state to context
          And passing step

        Scenario: failing before scenario
          Then adding step state to context
          And passing step

        Scenario: failing after scenario
          Then adding step state to context
          And passing step

      """
    When I run feature suite with formatter "pretty"

    Then the suite should have failed
    And the rendered output will be as follows:
    """
      Feature: scenario hook errors

        Scenario: failing before and after scenario # normal.feature:3
          Then adding step state to context         # suite_context_test.go:0 -> InitializeScenario.func17
          after scenario hook failed: failed in after scenario hook, step error: before scenario hook failed: failed in before scenario hook
          And passing step                          # suite_context_test.go:0 -> InitializeScenario.func2

        Scenario: failing before scenario   # normal.feature:7
          Then adding step state to context # suite_context_test.go:0 -> InitializeScenario.func17
          before scenario hook failed: failed in before scenario hook
          And passing step                  # suite_context_test.go:0 -> InitializeScenario.func2

        Scenario: failing after scenario    # normal.feature:11
          Then adding step state to context # suite_context_test.go:0 -> InitializeScenario.func17
          And passing step                  # suite_context_test.go:0 -> InitializeScenario.func2
          after scenario hook failed: failed in after scenario hook

      --- Failed steps:

        Scenario: failing before and after scenario # normal.feature:3
          Then adding step state to context # normal.feature:4
            Error: after scenario hook failed: failed in after scenario hook, step error: before scenario hook failed: failed in before scenario hook

        Scenario: failing before scenario # normal.feature:7
          Then adding step state to context # normal.feature:8
            Error: before scenario hook failed: failed in before scenario hook

        Scenario: failing after scenario # normal.feature:11
          And passing step # normal.feature:13
            Error: after scenario hook failed: failed in after scenario hook


      3 scenarios (3 failed)
      6 steps (1 passed, 3 failed, 2 skipped)
      0s
    """