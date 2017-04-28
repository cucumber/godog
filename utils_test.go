package godog

import (
	"testing"
	"time"
)

// this zeroes the time throughout whole test suite
// and makes it easier to assert output
// activated only when godog tests are being run
func init() {
	timeNowFunc = func() time.Time {
		return time.Time{}
	}
}

func TestTimeNowFunc(t *testing.T) {
	now := timeNowFunc()
	if !now.IsZero() {
		t.Fatalf("expected zeroed time, but got: %s", now.Format(time.RFC3339))
	}
}
