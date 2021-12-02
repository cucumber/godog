package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/cucumber/godog"
)

func init() {
	// we register an sql driver txdb
	txdb.Register("txdb", "mysql", "root@/godog_test")
}

type apiFeature struct {
	server
	resp *httptest.ResponseRecorder
}

func (a *apiFeature) resetResponse(*godog.Scenario) {
	a.resp = httptest.NewRecorder()
	if a.db != nil {
		a.db.Close()
	}
	db, err := sql.Open("txdb", "api")
	if err != nil {
		panic(err)
	}
	a.db = db
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
	case "/users":
		a.users(a.resp, req)
		if a.db != nil {//close db to rollback prev insert data
			a.db.Close()
		}
	default:
		err = fmt.Errorf("unknown endpoint: %s", endpoint)
	}
	return
}

func (a *apiFeature) theResponseCodeShouldBe(code int) error {
	if code != a.resp.Code {
		if a.resp.Code >= 400 {
			return fmt.Errorf("expected response code to be: %d, but actual is: %d, response message: %s", code, a.resp.Code, string(a.resp.Body.Bytes()))
		}
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

func (a *apiFeature) thereAreUsers(users *godog.Table) error {
	var fields []string
	var marks []string
	head := users.Rows[0].Cells
	for _, cell := range head {
		fields = append(fields, cell.Value)
		marks = append(marks, "?")
	}

	stmt, err := a.db.Prepare("INSERT INTO users (" + strings.Join(fields, ", ") + ") VALUES(" + strings.Join(marks, ", ") + ")")
	if err != nil {
		return err
	}
	for i := 1; i < len(users.Rows); i++ {
		var vals []interface{}
		for n, cell := range users.Rows[i].Cells {
			switch head[n].Value {
			case "username":
				vals = append(vals, cell.Value)
			case "email":
				vals = append(vals, cell.Value)
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}
		if _, err = stmt.Exec(vals...); err != nil {
			return err
		}
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	api := &apiFeature{}

	ctx.BeforeScenario(api.resetResponse)

	ctx.Step(`^I send "(GET|POST|PUT|DELETE)" request to "([^"]*)"$`, api.iSendrequestTo)
	ctx.Step(`^the response code should be (\d+)$`, api.theResponseCodeShouldBe)
	ctx.Step(`^the response should match json:$`, api.theResponseShouldMatchJSON)
	ctx.Step(`^there are users:$`, api.thereAreUsers)
}
