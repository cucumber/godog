package godog

import (
	"testing"
	"time"
)

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
