# Step Checker

You can use this step checker tool to help detect missing, duplicate or ambiguous steps
implemented in a `godog` test suite.  Early detection of the presence of these prior to
submitting a PR or running a Jenkins pipeline with new or modified step implementations
can potentially save time & help prevent the appearance of unexpected false positives
or negative test results, as described in this example's Problem Statement (see outer
README file).

## Invocation

Example:

```shell
$ cd ~/repos/godog/_examples/dupsteps
$ go run cmd/stepchecker/main.go tests
Found 5 feature step(s):
1. "I ran over a nail and got a flat tire"
2. "I fixed it"
  - 2 matching godog step(s) found:
    from: tests/features/dupsteps.feature:7
    from: tests/features/dupsteps.feature:13
      to: tests/dupsteps_test.go:93:11
      to: tests/dupsteps_test.go:125:11
3. "I can continue on my way"
4. "I accidentally poured concrete down my drain and clogged the sewer line"
5. "I can once again use my sink"

Found 5 godog step(s):
1. "^I fixed it$"
2. "^I can once again use my sink$"
3. "^I ran over a nail and got a flat tire$"
4. "^I can continue on my way$"
5. "^I accidentally poured concrete down my drain and clogged the sewer line$"

2024/10/06 15:38:50 1 issue(s) found
exit status 1
$ _
```
