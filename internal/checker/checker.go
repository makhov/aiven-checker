package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

const defaultHTTPTimeout = 60 * time.Second

const ResultsTopic = "check_results"

type Publisher interface {
	Publish(topic string, messages ...*message.Message) error
}

// Checker describes checker struct
type Checker struct {
	tasks     []Task
	client    *http.Client
	publisher Publisher

	done    chan bool
	checkCh chan Check
}

// NewChecker creates new checker instance
func NewChecker(tasksFile string, publisher Publisher) (*Checker, error) {
	log.Printf("Task file: %s", tasksFile)
	b, err := ioutil.ReadFile(tasksFile)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}

	var tc TaskConfig
	err = json.Unmarshal(b, &tc)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	log.Printf("Read %d tasks", len(tc.Tasks))

	return &Checker{
		tasks: tc.Tasks,
		// Will use a base HTTP client for simplicity.
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		publisher: publisher,
		done:      make(chan bool, 1),
		checkCh:   make(chan Check, 1000),
	}, nil
}

// Run runs all checks
func (c *Checker) Run() {
	for _, t := range c.tasks {
		go func(t Task) {

			ticker := time.NewTicker(t.Period.Duration)
			for {
				select {
				case <-c.done:
					ticker.Stop()
					return
				case <-ticker.C:
					// nolint:govet
					ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
					cr := c.check(ctx, t)

					c.checkCh <- Check{
						Task:   &t,
						Result: cr,
					}
				}
			}
		}(t)
	}

	for check := range c.checkCh {
		log.Printf("Publish check result for task. URL: '%s', regexp: '%s'", check.Task.URL, check.Task.Regexp)
		err := c.publish(check)
		if err != nil {
			log.Printf("publish message error: %v", err)
		}
	}
}

func (c *Checker) publish(check Check) error {
	payload, err := json.Marshal(check)
	if err != nil {
		return fmt.Errorf("payload marshalling error: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	err = c.publisher.Publish(ResultsTopic, msg)
	if err != nil {
		return fmt.Errorf("publishing error: %w", err)
	}

	return nil
}

// Close gracefully stops checker
func (c *Checker) Close() error {
	close(c.done)
	close(c.checkCh)

	return nil
}
