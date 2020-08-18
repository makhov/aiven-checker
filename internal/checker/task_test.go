package checker

import (
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTask_UnmarshalJSON_Success(t *testing.T) {
	var testCases = []struct {
		json         string
		expectedTask *Task
	}{
		{
			`{"url": "http://httpbin.org", "period": "5s", "regexp": ".*"}`,
			&Task{
				URL:    "http://httpbin.org",
				Period: Duration{5 * time.Second},
				Regexp: `.*`,
				regexp: regexp.MustCompile(`.*`),
			},
		},
		{
			`{"url": "http://httpbin.org", "period": "10m"}`,
			&Task{
				URL:    "http://httpbin.org",
				Period: Duration{10 * time.Minute},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			var actualTask *Task
			err := json.Unmarshal([]byte(tc.json), &actualTask)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTask, actualTask)
		})
	}
}

func TestTask_UnmarshalJSON_Error(t *testing.T) {

	var testCases = []struct {
		json          string
		expectedError error
	}{
		{
			`{"url": "http://httpbin.org", "period": "one minute"}`,
			ErrInvalidDuration,
		},
		{
			`{"url": "http://httpbin.org"}`,
			ErrInvalidDuration,
		},
		{
			`{"url": "http://httpbin.org", "period": "1m", "regexp": "("}`,
			ErrInvalidRegexp,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			var actualTask *Task
			err := json.Unmarshal([]byte(tc.json), &actualTask)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tc.expectedError))
		})
	}
}
