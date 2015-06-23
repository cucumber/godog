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
