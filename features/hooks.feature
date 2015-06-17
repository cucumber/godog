Feature: suite hooks
  In order to run tasks before and after important events
  As a test suite
  I need to provide a way to hook into these events

  Background:
    Given I have a before scenario hook
    And a feature path "features/load_features.feature:6"
    # When I parse and run features
