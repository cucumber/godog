
Feature: docstring parsing

  Scenario: should be able to convert a Doc String to a `*godog.DocString` argument
    Given call func(*godog.DocString) with 'text':
    """
    text
    """

  Scenario: should be able to convert a Doc String to a `string` argument
    Given call func(string) with 'text':
    """
    text
    """

