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
$ go run cmd/stepchecker/main.go -- demo/*.go features/*.feature
Found 5 feature step(s):
1. "I can continue on my way"
2. "I accidentally poured concrete down my drain and clogged the sewer line"
3. "I fixed it"
  - 2 matching godog step(s) found:
    from: features/cloggedDrain.feature:7
    from: features/flatTire.feature:7
      to: demo/dupsteps_test.go:93:11
      to: demo/dupsteps_test.go:125:11
4. "I can once again use my sink"
5. "I ran over a nail and got a flat tire"

Found 5 godog step(s):
1. "^I can continue on my way$"
2. "^I accidentally poured concrete down my drain and clogged the sewer line$"
3. "^I fixed it$"
4. "^I can once again use my sink$"
5. "^I ran over a nail and got a flat tire$"

2024/10/10 20:18:57 1 issue(s) found
exit status 1
```
