package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChecker_check(t *testing.T) {

	var testCases = []struct {
		handler  http.Handler
		regexp   *regexp.Regexp
		timeout  time.Duration
		expected *CheckResult
	}{
		{
			successHandler,
			nil,
			time.Second * 5,
			&CheckResult{
				CheckStatus:    CheckStatusOK,
				HTTPStatusCode: 200,
			},
		},
		{
			successHandler,
			regexp.MustCompile("hitchhiker"),
			time.Second * 5,
			&CheckResult{
				CheckStatus:    CheckStatusOK,
				HTTPStatusCode: 200,
			},
		},
		{
			successHandler,
			regexp.MustCompile("nonexistent-regex"),
			time.Second * 5,
			&CheckResult{
				CheckStatus:    CheckStatusRegexMatchFailed,
				HTTPStatusCode: 200,
			},
		},
		{
			successHandler,
			regexp.MustCompile("nonexistent-regex"),
			time.Second * 5,
			&CheckResult{
				CheckStatus:    CheckStatusRegexMatchFailed,
				HTTPStatusCode: 200,
			},
		},
		{
			slowHandler,
			nil,
			time.Second,
			&CheckResult{
				CheckStatus: CheckStatusRequestTimeout,
			},
		},
	}

	c := &Checker{
		client: &http.Client{},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			s := httptest.NewServer(tc.handler)
			defer s.Close()

			task := Task{
				URL:    s.URL,
				regexp: tc.regexp,
			}
			ctx, _ := context.WithTimeout(context.Background(), tc.timeout)

			actual := c.check(ctx, task)
			tc.expected.Time = actual.Time
			tc.expected.Duration = actual.Duration

			assert.Equal(t, tc.expected, actual)
		})
	}
}

var successHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`A towel, it says, is about the most massively useful thing an interstellar 
					hitchhiker can have. Partly it has great practical value.`))
})

var slowHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	time.Sleep(2 * time.Second)
	w.WriteHeader(http.StatusOK)
})
