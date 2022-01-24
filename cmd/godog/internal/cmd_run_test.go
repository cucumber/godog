// We like to test and run the buildAndRunGodog function in the context where builder.Build
// fails. And then make sure that error gets reported to the user in the sensible way.

package internal_test

import (
	"fmt"
	"testing"

	"github.com/cucumber/godog/cmd/godog/internal"
)

func Test_CmdRun(t *testing.T) {
	cmd := internal.CreateRunCmd()
	err := cmd.Execute()
	if err != nil {
		fmt.Printf("Erorrrr!!! %v", err)
	}
}
