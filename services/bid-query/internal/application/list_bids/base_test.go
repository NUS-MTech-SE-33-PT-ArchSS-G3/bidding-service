package list_bids

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirection_String(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		expected  string
	}{
		{
			name:      "DirectionAsc returns 'asc'",
			direction: DirectionAsc,
			expected:  "asc",
		},
		{
			name:      "DirectionDesc returns 'desc'",
			direction: DirectionDesc,
			expected:  "desc",
		},
		{
			name:      "invalid direction defaults to 'desc'",
			direction: Direction(999),
			expected:  "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.direction.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
