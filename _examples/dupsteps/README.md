# Duplicate Steps

## Problem Statement

In a testing pipeline in which many `godog` Scenario test cases for a single Feature are run
within a single `godog.TestSuite` (in order to produce a single report of results), it quickly
became apparent that Cucumber's requirement for a global one-to-one pairing between the Pattern
used to match the text of a Step to the Function implementing it is problematic.

In the illustrative example provided (see the [demo](./demo) and associated [features](./features) folders),
Steps with matching text (e.g., `I fixed it`) appear in two Scenarios; but, each calls
for a _different_ implementation, specific to its Scenario.  Cucumber's requirement (as
mentioned above) would force a single implementation of this Step across both Scenarios.
To accommodate this requirement, either the Gherkin (i.e., the _given_ business language)
would have to change, or coding best practices (e.g., Single Responsibility Principle,
Separation of Concerns, Modularity, Encapsulation, Cohesion, etc.) would have to give.

Running the tests for the two Scenarios _separately_ (e.g., using a separate `godog.TestSuite`)
could "solve" the problem, as matching the common Step text to its scenario-specific Function
would then be unambiguous within the Scenario-specific testing run context.  However, a hard requirement
within our build pipeline requires a single "cucumber report" file to be produced as evidence
of the success or failure of _all_ required test Scenarios.  Because `godog` produces a
_separate_ report for each invocation of `godog.TestSuite`, _something had to change._

## Problem Detection

A ["step checker" tool](cmd/stepchecker/README.md) was created to facilitate early detection
of the problem situation described above.  Using this tool while modifying or adding tests
for new scenarios given by the business was proven useful as part of a "shift left" testing
strategy.

## Proposed Solution

A ["solution" was proposed](solution/README.md) of using a simple "report combiner" in conjunction
with establishment of separate, standard `go` tests.

The main idea in the proposed "solution" _is to use separate_ `godog.TestSuite` instances
to partition execution of the test Scenarios, then collect and combine their output into
the single "cucumber report" required by our build pipeline.

## Notes

- See [PR-636](https://github.com/cucumber/godog/pull/636) dealing with a related issue: when `godog` chooses an
  incorrect Step Function when more than one matches the text
  from a Scenario being tested (i.e., an "ambiguous" Step, such as illustrated
  by the `I fixed it` Step in the nested `demo` folder).

