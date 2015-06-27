/*
Package godog is a behavior-driven development framework, a tool to describe your
application based on the behavior and run these specifications. The features are
described by a human-readable gherkin language.

Godog does not intervene with the standard "go test" command and it's behavior.
You can leverage both frameworks to functionally test your application while
maintaining all test related source code in *_test.go files.

Godog acts similar compared to "go test" command. It builds all package sources
to a single main package file and replaces main func with it's own and runs the
build to test described application behavior in feature files.
Production builds remains clean without any overhead.

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

Godog was inspired by Behat and the above description is taken from it's documentation.
*/
package godog

// Version of package - based on Semantic Versioning 2.0.0 http://semver.org/
const Version = "v0.1.0"
