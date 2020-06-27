package formatters

import (
	"testing"
	"time"

	"github.com/cucumber/godog/internal/utils"
)

// this zeroes the time throughout whole test suite
// and makes it easier to assert output
// activated only when godog tests are being run
func init() {
	utils.TimeNowFunc = func() time.Time {
		return time.Time{}
	}
}

func TestTimeNowFunc(t *testing.T) {
	now := utils.TimeNowFunc()
	if !now.IsZero() {
		t.Fatalf("expected zeroed time, but got: %s", now.Format(time.RFC3339))
	}
}
