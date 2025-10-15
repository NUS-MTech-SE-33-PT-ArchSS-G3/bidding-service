package list_bids

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Cursor struct {
	At time.Time `json:"at"`
	ID string    `json:"id"`
}

func encodeCursor(c *Cursor) (*string, error) {
	if c == nil {
		return nil, nil
	}

	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("encode cursor: %w", err)
	}

	s := base64.RawURLEncoding.EncodeToString(b)
	return &s, nil
}

func decodeCursor(s string) (*Cursor, error) {
	if s == "" {
		return nil, nil
	}

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}

	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("decode cursor json: %w", err)
	}

	if c.ID == "" || c.At.IsZero() {
		return nil, ErrInvalidCursor
	}
	return &c, nil
}
