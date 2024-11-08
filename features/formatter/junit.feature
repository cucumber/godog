Feature: junit formatter
  Smoke test of junit formatter.
  Comprehensive tests at internal/formatters.

  Scenario: check formatter is available

    Given a feature "test.feature" file:
      """
        Feature: check the formatter is available
          Scenario: trivial scenario
            Given a passing step
      """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """
    <?xml version="1.0" encoding="UTF-8"?>
    <testsuites name="godog" tests="1" skipped="0" failures="0" errors="0" time="9999.9999">
    <testsuite name="check the formatter is available" tests="1" skipped="0" failures="0" errors="0" time="9999.9999">
    <testcase name="trivial scenario" status="passed" time="9999.9999"></testcase>
    </testsuite>
    </testsuites>
    """


  Scenario: Support of Feature Plus Scenario Node
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
    <?xml version="1.0" encoding="UTF-8"?>
    <testsuites name="godog" tests="1" skipped="0" failures="0" errors="0" time="0">
      <testsuite name="simple feature" tests="1" skipped="0" failures="0" errors="0" time="0">
        <testcase name="simple scenario" status="" time="0"></testcase>
      </testsuite>
    </testsuites>
    """

  Scenario: Support of Feature Plus Scenario Node With Tags
    Given a feature "features/simple.feature" file:
    """
        @TAG1
        Feature: simple feature
            simple feature description
        @TAG2 @TAG3
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="1" skipped="0" failures="0" errors="0" time="0">
        <testsuite name="simple feature" tests="1" skipped="0" failures="0" errors="0" time="0">
          <testcase name="simple scenario" status="" time="0"></testcase>
        </testsuite>
      </testsuites>
    """
  Scenario: Support of Feature Plus Scenario Outline
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Scenario Outline: simple scenario
            simple scenario description

        Examples: simple examples
        | status |
        | pass   |
        | fail   |
    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="2" skipped="0" failures="0" errors="0" time="0">
        <testsuite name="simple feature" tests="2" skipped="0" failures="0" errors="0" time="0">
          <testcase name="simple scenario #1" status="" time="0"></testcase>
          <testcase name="simple scenario #2" status="" time="0"></testcase>
        </testsuite>
      </testsuites>
    """

  Scenario: Support of Feature Plus Scenario Outline With Tags
    Given a feature "features/simple.feature" file:
    """
        @TAG1
        Feature: simple feature
            simple feature description

        @TAG2
        Scenario Outline: simple scenario
            simple scenario description

        @TAG3
        Examples: simple examples
        | status |
        | pass   |
        | fail   |
    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="2" skipped="0" failures="0" errors="0" time="0">
        <testsuite name="simple feature" tests="2" skipped="0" failures="0" errors="0" time="0">
          <testcase name="simple scenario #1" status="" time="0"></testcase>
          <testcase name="simple scenario #2" status="" time="0"></testcase>
        </testsuite>
      </testsuites>
    """
  Scenario: Support of Feature Plus Scenario With Steps
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Scenario: simple scenario
            simple scenario description

        Given passing step
        Then a failing step

    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="1" skipped="0" failures="1" errors="0" time="0">
        <testsuite name="simple feature" tests="1" skipped="0" failures="1" errors="0" time="0">
          <testcase name="simple scenario" status="failed" time="0">
            <failure message="Step a failing step: intentional failure"></failure>
          </testcase>
        </testsuite>
      </testsuites>
    """
  Scenario: Support of Feature Plus Scenario Outline With Steps
    Given a feature "features/simple.feature" file:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario
        simple scenario description

          Given <status> step

        Examples: simple examples
        | status |
        | passing |
        | failing |

    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="2" skipped="0" failures="1" errors="0" time="0">
        <testsuite name="simple feature" tests="2" skipped="0" failures="1" errors="0" time="0">
          <testcase name="simple scenario #1" status="passed" time="0"></testcase>
          <testcase name="simple scenario #2" status="failed" time="0">
            <failure message="Step failing step: intentional failure"></failure>
          </testcase>
        </testsuite>
      </testsuites>
    """

  # Currently godog only supports comments on Feature and not
  # scenario and steps.
  Scenario: Support of Comments
    Given a feature "features/simple.feature" file:
    """
        #Feature comment
        Feature: simple feature
          simple description

          Scenario: simple scenario
          simple feature description
    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="1" skipped="0" failures="0" errors="0" time="0">
        <testsuite name="simple feature" tests="1" skipped="0" failures="0" errors="0" time="0">
          <testcase name="simple scenario" status="" time="0"></testcase>
        </testsuite>
      </testsuites>
    """
  Scenario: Support of Docstrings
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
          simple description

          Scenario: simple scenario
          simple feature description

          Given passing step
          \"\"\" content type
          step doc string
          \"\"\"
    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="1" skipped="0" failures="0" errors="0" time="0">
        <testsuite name="simple feature" tests="1" skipped="0" failures="0" errors="0" time="0">
          <testcase name="simple scenario" status="passed" time="0"></testcase>
        </testsuite>
      </testsuites>
    """
  Scenario: Support of Undefined, Pending and Skipped status
    Given a feature "features/simple.feature" file:
    """
      Feature: simple feature
      simple feature description

      Scenario: simple scenario
      simple scenario description

        Given passing step
        And pending step
        And undefined
        And passing step

    """
    When I run feature suite with formatter "junit"
    Then the rendered xml will be as follows:
    """ application/xml
      <?xml version="1.0" encoding="UTF-8"?>
      <testsuites name="godog" tests="1" skipped="0" failures="0" errors="1" time="0">
        <testsuite name="simple feature" tests="1" skipped="0" failures="0" errors="1" time="0">
          <testcase name="simple scenario" status="undefined" time="0">
            <error message="Step pending step: TODO: write pending definition" type="pending"></error>
            <error message="Step undefined" type="undefined"></error>
            <error message="Step passing step" type="skipped"></error>
          </testcase>
        </testsuite>
      </testsuites>
    """
