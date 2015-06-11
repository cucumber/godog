[![Build Status](https://travis-ci.org/DATA-DOG/godog.png)](https://travis-ci.org/DATA-DOG/godog)
[![GoDoc](https://godoc.org/github.com/DATA-DOG/godog/gherkin?status.svg)](https://godoc.org/github.com/DATA-DOG/godog/gherkin)

# Gherkin Parser for GO

Package gherkin is a gherkin language parser based on [specification][gherkin]
specification. It parses a feature file into the it's structural representation. It also
creates an AST tree of gherkin Tokens read from the file.

With gherkin language you can describe your application behavior as features in
human-readable and machine friendly language. See the example directory.

### Be aware that

The work is still in progress and is not functional yet, neither it is intended for production usage.
The general public API may brake.

### Documentation

See [godoc][godoc].

### License

Licensed under the [three clause BSD license][license]

[godoc]: http://godoc.org/github.com/DATA-DOG/godog/gherkin "Documentation on godoc for gherkin"
[gherkin]: https://cucumber.io/docs/reference "Gherkin feature file language"
[license]: http://en.wikipedia.org/wiki/BSD_licenses "The three clause BSD license"
