# An example of API feature

The following example demonstrates steps how we describe and test our API using **godog**.

### Step 1

Describe our feature. Imagine we need a REST API with **json** format. Lets from the point, that
we need to have a **/version** endpoint, which responds with a version number. We also need to manage
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
        "version": "v0.5.3"
      }
      """
```

Save it as **version.feature**.
Now we have described a success case and an error when the request method is not allowed.

### Step 2

Run **godog version.feature**. You should see the following result, which says that all of our
steps are yet undefined and provide us with the snippets to implement them.

![Screenshot](https://raw.github.com/DATA-DOG/godog/master/examples/api/screenshots/undefined.png)

### Step 3

Lets copy the snippets to **api_test.go** and modify it for our use case. Since we know that we will
need to store state within steps (a response), we should introduce a structure with some variables.

``` go
// file: api_test.go
package main

import (
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

type apiFeature struct {
}

func (a *apiFeature) iSendrequestTo(method, endpoint string) error {
	return godog.ErrPending
}

func (a *apiFeature) theResponseCodeShouldBe(code int) error {
	return godog.ErrPending
}

func (a *apiFeature) theResponseShouldMatchJSON(body *gherkin.DocString) error {
	return godog.ErrPending
}

func FeatureContext(s *godog.Suite) {
	api := &apiFeature{}
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, api.iSendrequestTo)
	s.Step(`^the response code should be (\d+)$`, api.theResponseCodeShouldBe)
	s.Step(`^the response should match json:$`, api.theResponseShouldMatchJSON)
}
```

### Step 4

Now we can implemented steps, since we know what behavior we expect:

``` go
// file: api_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

type apiFeature struct {
	resp *httptest.ResponseRecorder
}

func (a *apiFeature) resetResponse(interface{}) {
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

func (a *apiFeature) theResponseShouldMatchJSON(body *gherkin.DocString) (err error) {
	var expected, actual []byte
	var data interface{}
	if err = json.Unmarshal([]byte(body.Content), &data); err != nil {
		return
	}
	if expected, err = json.Marshal(data); err != nil {
		return
	}
	actual = a.resp.Body.Bytes()
	if !bytes.Equal(actual, expected) {
		err = fmt.Errorf("expected json, does not match actual: %s", string(actual))
	}
	return
}

func FeatureContext(s *godog.Suite) {
	api := &apiFeature{}

	s.BeforeScenario(api.resetResponse)

	s.Step(`^I send "(GET|POST|PUT|DELETE)" request to "([^"]*)"$`, api.iSendrequestTo)
	s.Step(`^the response code should be (\d+)$`, api.theResponseCodeShouldBe)
	s.Step(`^the response should match json:$`, api.theResponseShouldMatchJSON)
}
```

**NOTE:** the `getVersion` handler call on **/version** endpoint. We actually need to implement it now.
If we made some mistakes in step implementations, we will know about it when we run the tests.

Though, we could also improve our **JSON** comparison function to range through the interfaces and
match their types and values.

In case if some router is used, you may search the handler based on the endpoint. Current example
uses a standard http package.

### Step 5

Finally, lets implement the **api** server:

``` go
// file: api.go
// Example - demonstrates REST API server implementation tests.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/godog"
)

func getVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fail(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data := struct {
		Version string `json:"version"`
	}{Version: godog.Version}

	ok(w, data)
}

func main() {
	http.HandleFunc("/version", getVersion)
	http.ListenAndServe(":8080", nil)
}

// fail writes a json response with error msg and status header
func fail(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")

	data := struct {
		Error string `json:"error"`
	}{Error: msg}

	resp, _ := json.Marshal(data)
	w.WriteHeader(status)

	fmt.Fprintf(w, string(resp))
}

// ok writes data to response with 200 status
func ok(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if s, ok := data.(string); ok {
		fmt.Fprintf(w, s)
		return
	}

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fail(w, "oops something evil has happened", 500)
		return
	}

	fmt.Fprintf(w, string(resp))
}
```

The implementation details are clearly production ready and the imported **godog** package is only
used to respond with the correct constant version number.

### Step 6

Run our tests to see whether everything is happening as we have expected: `godog version.feature`

![Screenshot](https://raw.github.com/DATA-DOG/godog/master/examples/api/screenshots/passed.png)

### Conclusions

Hope you have enjoyed it like I did.

Any developer (who is the target of our application) can read and remind himself about how API behaves.
