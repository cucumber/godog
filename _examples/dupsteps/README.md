# Duplicate Steps

## Problem Statement

In a testing pipeline in which many `godog` scenario test cases for a single feature were
written and maintained by different developers, it quickly became apparent that Cucumber's
approach of considering step definitions as "global" - i.e., the need for a step's text to
identify it uniquely unique across all scenarios being tested - was problematic.

In the example provided here (see the `demo` and associated `features` folders),
the fact that the step with text `I fixed it` is used in two scenarios, but needs to
be interpreted _differently_ based upon which scenario is executing, implies that
either the development of both scenarios needs to be coordinated (i.e., to come
to a common implementation of that step, like an API), _or_ the step text needs
to be unique between them so that the implementations can be different.

Running the tests for the two scenarios separately (e.g., using a separate 
`godog.TestSuite`) would "solve" the problem, as the step implementations 
would be unique within each testing context.  However, a hard requirement for
a single testing phase within our build pipeline requires a single "cucumber
report" file to be produced as evidence of the success or failure of each
test scenario.  And, `godog` produces a separate report for each invocation
of `godog.TestSuite`, so _something needed to change._

## Proposed Solution

See a proposed "solution" of using a simple "report combiner" in conjunction
with establishment of separate, standard `go` tests, within the nested `solution`
folder.

The main approach is to feed each `godog.TestSuite` used to partition execution
of the test's scenarios with its own `Output` buffer, and to combine them
into a single "cucumber report" meeting the needs of our build pipeline.

## Notes

- See [PR-636](https://github.com/cucumber/godog/pull/636) dealing with a related issue: when `godog` chooses an
  incorrect step function when more than one step function matches the text
  from a scenario being tested (i.e., an "ambiguous" step, such as illustrated
  by the `I fixed it` step in the nested `demo` folder).
- _NOTE: until the change made in [PR-636](https://github.com/cucumber/godog/pull/636)
  is made available (it's not yet been released as of this writing), you can use something
  like the [stepchecker](cmd/stepchecker/README.md) to detect such cases._

