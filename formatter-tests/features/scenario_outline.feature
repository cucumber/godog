@outline @tag
Feature: outline

  @scenario
  Scenario Outline: outline
    Given passing step
    When passing step
    Then odd <odd> and even <even> number

    @tagged
    Examples: tagged
      | odd | even |
      | 1   | 2    |
      | 2   | 0    |
      | 3   | 11   |

    @tag2
    Examples:
      | odd | even |
      | 1   | 14   |
      | 3   | 9    |
