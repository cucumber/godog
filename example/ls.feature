Feature: ls
  In order to see the directory structure
  As a UNIX user
  I need to be able to list directory contents

  Background:
    Given I am in a directory "test"

  Scenario: lists files in directory
    Given I have a file named "foo"
    And I have a file named "bar"
    When I run ls
    Then I should get output:
      """
      bar
      foo
      """

  Scenario: lists files and directories
    Given I have a file named "foo"
    And I have a directory named "dir"
    When I run ls
    Then I should get output:
      """
      dir
      foo
      """
