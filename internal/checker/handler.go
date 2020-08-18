package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

// ErrBadMessage error for broken json
var ErrBadMessage = fmt.Errorf("bad message")

// Saver saves checks
type Saver interface {
	Save(ctx context.Context, c *Check) error
}

// NewHandler returns message handling functions
func NewHandler(s Saver) func(msg *message.Message) error {
	return func(msg *message.Message) error {
		var payload *Check
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return ErrBadMessage
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return s.Save(ctx, payload)
	}
}
