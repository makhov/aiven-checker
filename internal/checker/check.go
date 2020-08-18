package checker

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CheckStatus string

const (
	CheckStatusOK               CheckStatus = "OK"
	CheckStatusHTTPErrorCode    CheckStatus = "HTTPErrorCode"
	CheckStatusInternalError    CheckStatus = "InternalError"
	CheckStatusRequestTimeout   CheckStatus = "RequestTimeout"
	CheckStatusRegexMatchFailed CheckStatus = "RegexMatchFailed"
)

// String implements fmt.Stringer interface
func (cs CheckStatus) String() string {
	return string(cs)
}

// Check describes check
type Check struct {
	Task   *Task        `json:"task"`
	Result *CheckResult `json:"result"`
}

// CheckResult describes result of check
type CheckResult struct {
	Time           time.Time   `json:"check_time"`
	CheckStatus    CheckStatus `json:"check_status"`
	ErrorMessage   string      `json:"error,omitempty"`
	HTTPStatusCode int         `json:"http_status_code"`
	Duration       int64       `json:"duration"`
}

func (c *Checker) check(ctx context.Context, t Task) *CheckResult {
	cr := &CheckResult{
		Time: time.Now(),
	}
	log.Printf("Perform check. URL: %s", t.URL)

	req, err := http.NewRequest(http.MethodGet, t.URL, nil)
	if err != nil {
		return errorResult(cr, CheckStatusInternalError, fmt.Sprintf("creating new request error: %v", err))
	}
	req = req.WithContext(ctx)

	startTime := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(startTime).Milliseconds()
	if err != nil {
		if errors.Is(err, http.ErrHandlerTimeout) || errors.Is(err, context.DeadlineExceeded) {
			return errorResult(cr, CheckStatusRequestTimeout, "")
		} else {
			return errorResult(cr, CheckStatusInternalError, fmt.Sprintf("making request error: %v", err))
		}
	}
	defer resp.Body.Close()

	cr.HTTPStatusCode = resp.StatusCode
	cr.Duration = duration

	if resp.StatusCode >= 400 {
		cr.CheckStatus = CheckStatusHTTPErrorCode
		return cr
	}

	if t.regexp != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errorResult(cr, CheckStatusInternalError, fmt.Sprintf("body read error: %v", err))
		}

		if !t.regexp.Match(b) {
			cr.CheckStatus = CheckStatusRegexMatchFailed
			return cr
		}
	}
	cr.CheckStatus = CheckStatusOK

	return cr

}

func errorResult(cr *CheckResult, status CheckStatus, msg string) *CheckResult {
	cr.CheckStatus = status
	cr.ErrorMessage = msg
	return cr
}
