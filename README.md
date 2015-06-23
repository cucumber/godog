[![Build Status](https://travis-ci.org/DATA-DOG/godog.svg?branch=master)](https://travis-ci.org/DATA-DOG/godog)
[![GoDoc](https://godoc.org/github.com/DATA-DOG/godog?status.svg)](https://godoc.org/github.com/DATA-DOG/godog)

# Godog

**Godog** is an open source behavior-driven development framework for [go][golang] programming language.
What is behavior-driven development, you ask? It’s the idea that you start by writing human-readable sentences that
describe a feature of your application and how it should work, and only then implement this behavior in software.

The project is inspired by [behat][behat] and [cucumber][cucumber] and is based on cucumber [gherkin specification][gherkin].

**Godog** does not intervene with the standard **go test** command and it's behavior. You can leverage both frameworks
to functionally test your application while maintaining all test related source code in **_test.go** files.

**Godog** acts similar compared to **go test** command. It builds all package sources to a single main package file
and replaces **main** func with it's own and runs the build to test described application behavior in feature files.
Production builds remains clean without any overhead.

### Install

    go get github.com/DATA-DOG/godog/cmd/godog

### Example

Imagine we have a **godog cart** to serve godogs for dinner. At first, we describe our feature:

``` gherkin
# file: /tmp/godog/godog.feature
Feature: eat godogs
  In order to be satiated
  As an user
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining
```

As a developer, your work is done as soon as you’ve made the program behave as
described in the Scenario.

If you run `godog godog.feature` inside the **/tmp/godog** directory.
You should see that the steps are undefined:

![Screenshot](https://raw.github.com/DATA-DOG/godog/master/screenshots/undefined.png)

``` go
/* file: /tmp/godog/godog.go */
package main

type GodogCart struct {
	reserve int
}

func (c *GodogCart) Eat(num int) { c.reserve -= num }

func (c *GodogCart) Available() int { return c.reserve }

func main() { /* usual main func */ }
```

If you run `godog godog.feature` inside the **/tmp/godog** directory.
You should see that the steps are undefined.

Now lets describe all steps to test the application behavior:

``` go
/* file: /tmp/godog/godog_test.go */
package main

import (
	"fmt"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

func (c *GodogCart) resetReserve(*gherkin.Scenario) {
	c.reserve = 0
}

func (c *GodogCart) thereAreNumGodogsInReserve(args ...*godog.Arg) error {
	c.reserve = args[0].Int()
	return nil
}

func (c *GodogCart) iEatNum(args ...*godog.Arg) error {
	c.Eat(args[0].Int())
	return nil
}

func (c *GodogCart) thereShouldBeNumRemaining(args ...*godog.Arg) error {
	if c.Available() != args[0].Int() {
		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", args[0].Int(), c.Available())
	}
	return nil
}

func godogCartContext(s godog.Suite) {
	c := &GodogCart{}
	// each time before running scenario reset reserve
	s.BeforeScenario(c.resetReserve)
	// register steps
	s.Step(`^there are (\d+) godogs?$`, c.thereAreNumGodogsInReserve)
	s.Step(`^I eat (\d+)$`, c.iEatNum)
	s.Step(`^there should be (\d+) remaining$`, c.thereShouldBeNumRemaining)
}
```

Now when you run the `godog godog.feature` again, you should see:

![Screenshot](https://raw.github.com/DATA-DOG/godog/master/screenshots/passed.png)

### Documentation

See [godoc][godoc] and [gherkin godoc][godoc_gherkin] for general API details.
See **.travis.yml** for supported **go** versions.

The public API is stable enough, but it may break until **1.0.0** version, see `godog --version`.

### Contributions

Feel free to open a pull request. Note, if you wish to contribute an extension to public (exported methods or types) -
please open an issue before to discuss whether these changes can be accepted. All backward incompatible changes are
and will be treated cautiously.

### License

All package dependencies are **MIT** or **BSD** licensed.

**Godog** is licensed under the [three clause BSD license][license]

[godoc]: http://godoc.org/github.com/DATA-DOG/godog "Documentation on godoc"
[godoc_gherkin]: http://godoc.org/github.com/DATA-DOG/godog/gherkin "Documentation on godoc for gherkin"
[golang]: https://golang.org/  "GO programming language"
[behat]: http://docs.behat.org/ "Behavior driven development framework for PHP"
[cucumber]: https://cucumber.io/ "Behavior driven development framework for Ruby"
[gherkin]: https://cucumber.io/docs/reference "Gherkin feature file language"
[license]: http://en.wikipedia.org/wiki/BSD_licenses "The three clause BSD license"
