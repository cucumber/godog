Feature: godog bdd suite
  In order to test application behavior
  As a suite
  I need to be able to register and run features

  Scenario: parses all features in path
    Given a feature path "features"
    When I parse features
    Then I should have 1 feature file
