// +build e2e

package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/makhov/aiven-checker/internal/checker"
)

func TestChecker(t *testing.T) {
	db := setUp(t)
	defer tearDown(t, db)

	var testCases = []struct {
		url            string
		regex          string
		expectedStatus checker.CheckStatus
	}{
		{"http://httpbin.org/status/200", "", checker.CheckStatusOK},
		{"http://httpbin.org/status/404", "", checker.CheckStatusHTTPErrorCode},
		{"http://httpbin.org/html", "Moby.*Dick", checker.CheckStatusOK},
		{"http://httpbin.org/html", "non-existent-pattern", checker.CheckStatusRegexMatchFailed},
	}

	for _, tc := range testCases {
		// For e2e tests, we run a checker with a list of tasks (see tasks.e2e.json) and
		// here we just need to check that the results appear in the database
		waiter(t, 100*time.Millisecond, 5*time.Second, func() error {
			var actualStatus string
			err := db.Get(&actualStatus, "SELECT status FROM checks WHERE url = $1 AND regexp = $2", tc.url, tc.regex)
			if err != nil {
				return err
			}

			if actualStatus != tc.expectedStatus.String() {
				return fmt.Errorf("wrong check status. want: %s, got: %s", tc.expectedStatus.String(), actualStatus)
			}

			return nil
		})
	}
}

func waiter(t *testing.T, interval, timeout time.Duration, f func() error) {
	var err error
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		err = f()
		if err == nil {
			return
		}

		time.Sleep(interval)
	}

	t.Errorf("waiter deadline excideed with error: %v", err)
}

func setUp(t *testing.T) *sqlx.DB {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Fatal("DATABASE_DSN is empty")
	}

	db := sqlx.MustConnect("pgx", dsn)

	return db
}

func tearDown(t *testing.T, db *sqlx.DB) {
	// clean up after run
	db.MustExec("DELETE FROM checks")
	db.Close()
}
