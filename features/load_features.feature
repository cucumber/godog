Feature: load features
  In order to run features
  As a test suite
  I need to be able to load features

  Scenario: load features within path
    Given a feature path "features"
    When I parse features
    Then I should have 2 feature files:
      """
      features/hooks.feature
      features/load_features.feature
      """

  Scenario: load a specific feature file
    Given a feature path "features/load_features.feature"
    When I parse features
    Then I should have 1 feature file:
      """
      features/load_features.feature
      """

  Scenario: load a feature file with a specified scenario
    Given a feature path "features/load_features.feature:6"
    When I parse features
    Then I should have 1 scenario registered

  Scenario: load a number of feature files
    Given a feature path "features/load_features.feature"
    And a feature path "features/hooks.feature"
    When I parse features
    Then I should have 2 feature files:
      """
      features/load_features.feature
      features/hooks.feature
      """
