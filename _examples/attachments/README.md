# An example of Making attachments to the reports

The JSON (and in future NDJSON) report formats allow the inclusion of data attachments.

These attachments could be console logs or file data or images for instance.

The example in this directory shows how the godog API is used to add attachments to the JSON report.


## Run the example

You must use the '-v' flag or you will not see the cucumber JSON output.

go test -v atttachment_test.go


