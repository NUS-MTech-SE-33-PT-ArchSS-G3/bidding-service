package list_bids

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name     string
		cursor   *Cursor
		wantErr  bool
		wantNil  bool
	}{
		{
			name:    "nil cursor returns nil",
			cursor:  nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name: "valid cursor encodes successfully",
			cursor: &Cursor{
				At: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				ID: "bid-123",
			},
			wantNil: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encodeCursor(tt.cursor)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.NotEmpty(t, *result)
			}
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	validCursor := &Cursor{
		At: fixedTime,
		ID: "bid-123",
	}
	validEncoded, _ := encodeCursor(validCursor)

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantNil  bool
		expected *Cursor
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			wantNil: true,
			wantErr: false,
		},
		{
			name:     "valid encoded cursor decodes successfully",
			input:    *validEncoded,
			wantNil:  false,
			wantErr:  false,
			expected: validCursor,
		},
		{
			name:    "invalid base64 returns error",
			input:   "not-valid-base64!@#$",
			wantErr: true,
		},
		{
			name:    "invalid json returns error",
			input:   "aW52YWxpZC1qc29u", // "invalid-json" in base64
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeCursor(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, result)
			} else if tt.expected != nil {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.True(t, tt.expected.At.Equal(result.At))
			}
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	t.Run("encode and decode produces same cursor", func(t *testing.T) {
		original := &Cursor{
			At: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			ID: "bid-123",
		}

		encoded, err := encodeCursor(original)
		assert.NoError(t, err)
		assert.NotNil(t, encoded)

		decoded, err := decodeCursor(*encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)

		assert.Equal(t, original.ID, decoded.ID)
		assert.True(t, original.At.Equal(decoded.At))
	})
}
