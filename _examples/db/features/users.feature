Feature: users
  In order to use users api
  As an API user
  I need to be able to manage users

  Scenario: should get empty users
    When I send "GET" request to "/users"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "users": []
      }
      """

  Scenario: should get users
    Given there are users:
      | username | email             |
      | john     | john.doe@mail.com |
      | jane     | jane.doe@mail.com |
    When I send "GET" request to "/users"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "users": [
          {
            "username": "john"
          },
          {
            "username": "jane"
          }
        ]
      }
      """

  Scenario: should get users when there is only one
    Given there are users:
      | username | email           |
      | gopher   | gopher@mail.com |
    When I send "GET" request to "/users"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "users": [
          {
            "username": "gopher"
          }
        ]
      }
      """
