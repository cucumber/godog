# Duplicate Steps

## Updated Statement

After submitting this example / feedback, while looking through the `godog` source
code, I came across [PR-636](https://github.com/cucumber/godog/pull/636) which deals
with the very same issue as described below - i.e., duplicate (ambiguous) steps.

- _NOTE: until the change made in [PR-636](https://github.com/cucumber/godog/pull/636)
  is made available (it's not yet been released as of this writing), you can use something
  like the [stepchecker](cmd/stepchecker/README.md) to detect such cases._

Indeed, this accepted and merged PR wasn't available in the `godog` library version
we are using (you can be sure I'll put in an upgrade request to the team, though!),
so I was not familiar with it.  

Now that this PR is in place and enabled when using `strict` mode (which we _are_
using), it will help us - at least our developers will receive a repeatable test
failure _right away_ when duplicate (ambiguous) steps are detected (though the
"error" message is a bit cryptic: i.e., it seems like it just says **"IS STRICT=2"**
in the given example case).  However, ultimately as described below we'd like
to be able to encapsulate step implementations completely within a scenario,
and it seems this desired functionality is not (yet?) available.

Therefore, the focus of this PR is now upon tips / suggestions for how to implement
the desired functionality of limiting the scope of a step definition to either
a scenario and/or perhaps a feature (as opposed to always globally).  Or possibly,
why that's a bad idea (I find it suspicious that hasn't been done already ;-0).

The original (pre-update above) text of this README is preserved for posterity,
below:

## Original Statement

This example reproduces the problem wherein duplicate steps are silently overridden.

### Motivation 

As a relatively new user of Cucumber & its [Gherkin syntax](https://cucumber.io/docs/gherkin/), and
as an implementer of steps for a scenario using `godog`, I'd like to have the ability to encapsulate
implementation of steps for a scenario without concern that a step function from a different scenario
will be called instead of mine.  At least, I'd like to have more control over automatic "re-use" of
step functions; either: (1) choose to limit the assignment of a step function to step text (or regex)
to be matched only within the scope of its enclosing scenario (preferred), (2) have `godog` throw an
error when an ambiguous match to a step function is defined or detected for a step's implementation.

Though I've begun to understand that re-use of step implementations is somewhat fundamental to the Gherkin design
(e.g., I've read about _[feature coupled step definitions](https://cucumber.io/docs/guides/anti-patterns/?lang=java#feature-coupled-step-definitions)_
and how Cucumber _"effectively flattens the features/ directory tree"_), it's still annoying that
the `godog` scaffold seems to _require us to conform_ to the Gherkin recommendation to _"organise
your steps by domain concept"_, and not by Feature or even Scenario, as would better suit our project.

What's ended up happening to several of our developers in our distributed BDD testing initiative is
they end up force-tweaking the text of their steps in order to avoid duplicates, as they don't have agency
over other scenario implementations which they need to run with in our Jenkins pipeline.  Later, they're
annoyed when their scenarios suddenly start failing after new scenarios are added having step
implementations with regex which _happens_ to match (thereby overriding) theirs.

_NOTE:_ due to a limitation in our Jenkins pipeline, all of our features & scenarios _must_ be executed
within the same `godog.Suite`, else (I realize) we could just "solve" this problem by running each scenario
in its own invocation of `godog`.  

### Summary

In light of the specifics of the "Motivation" above, the stated "problem" here might then be more
effectively re-characterized as a "request" to give more control to the end-user, as suggested above.
