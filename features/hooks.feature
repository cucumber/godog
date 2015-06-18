Feature: suite hooks
  In order to run tasks before and after important events
  As a test suite
  I need to provide a way to hook into these events

  Background:
    Given I'm listening to suite events

  Scenario: triggers before scenario hook
    Given a feature path "features/load_features.feature:6"
    When I run feature suite
    Then there was event triggered before scenario "load features within path"
