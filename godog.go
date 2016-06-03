/*
Package godog is the official Cucumber BDD framework for Golang, it merges specification
and test documentation into one cohesive whole.

Godog does not intervene with the standard "go test" command and it's behavior.
You can leverage both frameworks to functionally test your application while
maintaining all test related source code in *_test.go files.

Godog acts similar compared to go test command. It leverages
a TestMain function introduced in go1.4 and clones the package sources
to a temporary build directory. The only change it does is adding a runner
test.go file and replaces TestMain func if it was used in tests.
Godog uses standard go ast and build utils to generate test suite package,
compiles it with go test -c command. It accepts all your environment exported
build related vars.

For example, imagine you’re about to create the famous UNIX ls command.
Before you begin, you describe how the feature should work, see the example below..

Example:
	Feature: ls
	  In order to see the directory structure
	  As a UNIX user
	  I need to be able to list the current directory's contents

	  Scenario:
		Given I am in a directory "test"
		And I have a file named "foo"
		And I have a file named "bar"
		When I run ls
		Then I should get output:
		  """
		  bar
		  foo
		  """

As a developer, your work is done as soon as you’ve made the ls command behave as
described in the Scenario.

Now, wouldn’t it be cool if something could read this sentence and use it to actually
run a test against the ls command? Hey, that’s exactly what this package does!
As you’ll see, Godog is easy to learn, quick to use, and will put the fun back into tests.

Godog was inspired by Behat and Cucumber the above description is taken from it's documentation.
*/
package godog

// Version of package - based on Semantic Versioning 2.0.0 http://semver.org/
const Version = "v0.4.3"
