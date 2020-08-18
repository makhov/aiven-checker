package checker

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

var (
	ErrInvalidDuration = fmt.Errorf("invalid duration")
	ErrInvalidRegexp   = fmt.Errorf("invalid regexp")
)

// TaskConfig describes tasks file format
type TaskConfig struct {
	Tasks []Task `json:"tasks"`
}

// Task describes task
type Task struct {
	URL    string   `json:"url"`
	Period Duration `json:"period"`
	Regexp string   `json:"regexp"`
	regexp *regexp.Regexp
}

// UnmarshalJSON implements json.Unmarshaler interface
func (t *Task) UnmarshalJSON(b []byte) error {
	temp := struct {
		URL    string `json:"url"`
		Period string `json:"period"`
		Regexp string `json:"regexp"`
	}{}
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	t.URL = temp.URL

	period, err := time.ParseDuration(temp.Period)
	if err != nil {
		return ErrInvalidDuration
	}
	t.Period = Duration{period}

	if temp.Regexp != "" {
		t.Regexp = temp.Regexp
		t.regexp, err = regexp.Compile(temp.Regexp)
		if err != nil {
			return ErrInvalidRegexp
		}
	}

	return nil
}

// Duration is wrapper for time.Duration. There are no marshal/unmarshal methods for time.Duration
type Duration struct {
	time.Duration
}

// MarshalJSON implements json.Marshaler interface
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid duration")
	}
}
