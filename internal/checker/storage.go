package checker

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Storage describes storage
type Storage struct {
	db *sqlx.DB
}

// NewStorage creates new storage instance
func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

// Save saves check to the database
func (s *Storage) Save(ctx context.Context, c *Check) error {
	q := `INSERT INTO checks 
			(url, Period, regexp, check_time, status, error_message, http_code, duration, created) 
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	_, err := s.db.ExecContext(ctx, q,
		c.Task.URL, c.Task.Period.String(), c.Task.Regexp, c.Result.Time,
		c.Result.CheckStatus, c.Result.ErrorMessage, c.Result.HTTPStatusCode, c.Result.Duration)
	return err
}
