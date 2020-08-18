// +build integration

package checker

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/require"
)

func TestStorage_Save(t *testing.T) {
	db := setUp(t)
	defer tearDown(t, db)
	s := NewStorage(db)

	type dbRow struct {
		URL          string `db:"url"`
		Period       string `db:"period"`
		Regexp       string `db:"regexp"`
		Status       string `db:"status"`
		ErrorMessage string `db:"error_message"`
		HTTPCode     int    `db:"http_code"`
		Duration     int64  `db:"duration"`
	}

	var testCases = []struct {
		check    *Check
		expected *dbRow
	}{
		{
			&Check{
				Task: &Task{
					URL:    "http://httpbin.org/status/200",
					Period: Duration{time.Second},
				},
				Result: &CheckResult{
					Time:           time.Now(),
					CheckStatus:    CheckStatusOK,
					HTTPStatusCode: http.StatusOK,
					Duration:       42,
				},
			},
			&dbRow{
				URL:      "http://httpbin.org/status/200",
				Period:   time.Second.String(),
				Status:   CheckStatusOK.String(),
				HTTPCode: 200,
				Duration: 42,
			},
		},
		{
			&Check{
				Task: &Task{
					URL:    "http://httpbin.org/status/500",
					Period: Duration{time.Second},
				},
				Result: &CheckResult{
					Time:        time.Now(),
					CheckStatus: CheckStatusHTTPErrorCode,
				},
			},
			&dbRow{
				URL:    "http://httpbin.org/status/500",
				Period: time.Second.String(),
				Status: CheckStatusHTTPErrorCode.String(),
			},
		},
		{
			&Check{
				Task: &Task{
					URL:    "http://httpbin.org/html",
					Period: Duration{time.Second},
					Regexp: ".*",
				},
				Result: &CheckResult{
					Time:           time.Now(),
					CheckStatus:    CheckStatusOK,
					HTTPStatusCode: http.StatusOK,
					Duration:       42,
				},
			},
			&dbRow{
				URL:      "http://httpbin.org/html",
				Period:   time.Second.String(),
				Regexp:   ".*",
				Status:   CheckStatusOK.String(),
				HTTPCode: 200,
				Duration: 42,
			},
		},
	}

	for _, tc := range testCases {
		err := s.Save(context.Background(), tc.check)
		require.NoError(t, err)

		actual := new(dbRow)
		q := `SELECT url, Period, regexp, status, error_message, http_code, duration FROM checks WHERE url = $1`
		err = db.QueryRowx(q, tc.check.Task.URL).StructScan(actual)
		require.NoError(t, err)

		require.Equal(t, tc.expected, actual)
	}
}

func setUp(t *testing.T) *sqlx.DB {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Fatal("DATABASE_DSN is empty")
	}

	db := sqlx.MustConnect("pgx", dsn)

	err := goose.Up(db.DB, "../../migrations")
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func tearDown(t *testing.T, db *sqlx.DB) {
	// clean up after run
	db.MustExec("DELETE FROM checks")
	db.Close()
}
