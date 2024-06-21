# An example of intercepting the step result for post-processing

There are situations where it is useful to be able to post-process the outcome of all the steps in a suite in a generic manner in order to manipulate the status result, any errors returned or the context.Context.

In order to facilitate this use case godog provides a seam where is is possible to inject a handler function to manipulate these values.

## Run the example

You must use the '-v' flag or you will not see the cucumber JSON output.

go test -v interceptor_test.go


