Feature: Godog should be able to retry flaky tests
    In order to help Go developers deal with flaky tests
    As a test suite
    I need to be able to return godog.Err to mark which steps should be retry

    Scenario: Test cases that pass aren't retried
        Given a step that always passes

    Scenario: Test cases that fail are retried if within the limit
        Given a step that passes the second time

    Scenario: Test cases that fail will continue to retry up to the limit
        Given a step that passes the third time

    Scenario: Test cases won't retry after failing more than the limit
        Given a step that always fails

    Scenario: Test cases won't retry when the status is UNDEFINED
        Given a non-existent step
