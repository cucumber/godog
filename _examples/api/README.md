# An example of API feature

The following example demonstrates steps how we describe and test our API using **godog**.

### Step 1

Describe our feature. Imagine we need a REST API with `json` format. Lets from the point, that
we need to have a `/version` endpoint, which responds with a version number. We also need to manage
error responses.

``` gherkin
# file: version.feature
Feature: get version
  In order to know godog version
  As an API user
  I need to be able to request version

  Scenario: does not allow POST method
    When I send "POST" request to "/version"
    Then the response code should be 405
    And the response should match json:
      """
      {
        "error": "Method not allowed"
      }
      """

  Scenario: should get version number
    When I send "GET" request to "/version"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "version": "v0.0.0-dev"
      }
      """
```

Save it as `features/version.feature`.
Now we have described a success case and an error when the request method is not allowed.

### Step 2

Execute `godog run`. You should see the following result, which says that all of our
steps are yet undefined and provide us with the snippets to implement them.

![Screenshot](https://raw.github.com/cucumber/godog/master/_examples/api/screenshots/undefined.png)

### Step 3

Lets copy the snippets to `api_test.go` and modify it for our use case. Since we know that we will
need to store state within steps (a response), we should introduce a structure with some variables.

``` go
// file: api_test.go
package main

import (
	"github.com/cucumber/godog"
)

type apiFeature struct {
}

func (a *apiFeature) iSendrequestTo(method, endpoint string) error {
	return godog.ErrPending
}

func (a *apiFeature) theResponseCodeShouldBe(code int) error {
	return godog.ErrPending
}

func (a *apiFeature) theResponseShouldMatchJSON(body *godog.DocString) error {
	return godog.ErrPending
}

func TestFeatures(t *testing.T) {
  suite := godog.TestSuite{
    ScenarioInitializer: InitializeScenario,
    Options: &godog.Options{
      Format:   "pretty",
      Paths:    []string{"features"},
      TestingT: t, // Testing instance that will run subtests.
    },
  }

  if suite.Run() != 0 {
    t.Fatal("non-zero status returned, failed to run feature tests")
  }
}

func InitializeScenario(s *godog.ScenarioContext) {
	api := &apiFeature{}
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, api.iSendrequestTo)
	s.Step(`^the response code should be (\d+)$`, api.theResponseCodeShouldBe)
	s.Step(`^the response should match json:$`, api.theResponseShouldMatchJSON)
}
```

### Step 4

Now we can implement steps, since we know what behavior we expect:

``` go
// file: api_test.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cucumber/godog"
)

type apiFeature struct {
	resp *httptest.ResponseRecorder
}

func (a *apiFeature) resetResponse(*godog.Scenario) {
	a.resp = httptest.NewRecorder()
}

func (a *apiFeature) iSendrequestTo(method, endpoint string) (err error) {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return
	}

	// handle panic
	defer func() {
		switch t := recover().(type) {
		case string:
			err = fmt.Errorf(t)
		case error:
			err = t
		}
	}()

	switch endpoint {
	case "/version":
		getVersion(a.resp, req)
	default:
		err = fmt.Errorf("unknown endpoint: %s", endpoint)
	}
	return
}

func (a *apiFeature) theResponseCodeShouldBe(code int) error {
	if code != a.resp.Code {
		return fmt.Errorf("expected response code to be: %d, but actual is: %d", code, a.resp.Code)
	}
	return nil
}

func (a *apiFeature) theResponseShouldMatchJSON(body *godog.DocString) (err error) {
	var expected, actual interface{}

	// re-encode expected response
	if err = json.Unmarshal([]byte(body.Content), &expected); err != nil {
		return
	}

	// re-encode actual response too
	if err = json.Unmarshal(a.resp.Body.Bytes(), &actual); err != nil {
		return
	}

	// the matching may be adapted per different requirements.
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("expected JSON does not match actual, %v vs. %v", expected, actual)
	}
	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	api := &apiFeature{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		api.resetResponse(sc)
		return ctx, nil
	})
	ctx.Step(`^I send "(GET|POST|PUT|DELETE)" request to "([^"]*)"$`, api.iSendrequestTo)
	ctx.Step(`^the response code should be (\d+)$`, api.theResponseCodeShouldBe)
	ctx.Step(`^the response should match json:$`, api.theResponseShouldMatchJSON)
}
```

**NOTE:** the `getVersion` handler is called on `/version` endpoint.
Executing `godog run` or `go test -v` will provide `undefined: getVersion` error, so we actually need to implement it now.
If we made some mistakes in step implementations, we will know about it when we run the tests.

Though, we could also improve our `JSON` comparison function to range through the interfaces and
match their types and values.

In case if some router is used, you may search the handler based on the endpoint. Current example
uses a standard http package.

### Step 5

Finally, lets implement the `API` server:

``` go
// file: api.go
// Example - demonstrates REST API server implementation tests.
package main

import (
	"encoding/json"
	"net/http"

	"github.com/cucumber/godog"
)

func getVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Version string `json:"version"`
	}{Version: godog.Version}

	ok(w, data)
}

// fail writes a json response with error msg and status header
func fail(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)

	data := struct {
		Error string `json:"error"`
	}{Error: msg}
	resp, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// ok writes data to response with 200 status
func ok(w http.ResponseWriter, data interface{}) {
	resp, err := json.Marshal(data)
	if err != nil {
		fail(w, "Oops something evil has happened", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func main() {
	http.HandleFunc("/version", getVersion)
	http.ListenAndServe(":8080", nil)
}
```

The implementation details are clearly production ready and the imported **godog** package is only
used to respond with the correct constant version number.

### Step 6

Run our tests to see whether everything is happening as we have expected: `go test -v`

![Screenshot](https://raw.github.com/cucumber/godog/master/_examples/api/screenshots/passed.png)

### Conclusions

Hope you have enjoyed it like I did.

Any developer (who is the target of our application) can read and remind himself about how API behaves.
