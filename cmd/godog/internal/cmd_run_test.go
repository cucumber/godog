// We like to test and run the buildAndRunGodog function in the context where builder.Build
// fails. And then make sure that error gets reported to the user in the sensible way.

package internal_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/cucumber/godog/cmd/godog/internal"
	"github.com/stretchr/testify/assert"
)

func Test_CmdRun(t *testing.T) {
	cmd := internal.CreateRunCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	err := cmd.Execute()
	assert.Nil(t, err)
	out, err := ioutil.ReadAll(b)
	assert.Nil(t, err)
	assert.Equal(t, "Blahhhhh", string(out))
}
