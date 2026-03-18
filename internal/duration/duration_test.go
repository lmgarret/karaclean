package duration_test

import (
	"strings"
	"testing"
	"time"

	"github.com/lm/karaclean/internal/duration"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        time.Duration
		wantErr     bool
		errContains string
	}{
		// Valid cases
		{name: "30 days", input: "30d", want: 30 * 24 * time.Hour},
		{name: "2 weeks", input: "2w", want: 14 * 24 * time.Hour},
		{name: "6 hours", input: "6h", want: 6 * time.Hour},
		{name: "1 month", input: "1mo", want: 30 * 24 * time.Hour},
		{name: "1 year", input: "1y", want: 365 * 24 * time.Hour},
		{name: "zero hours", input: "0h", want: 0},
		{name: "zero days", input: "0d", want: 0},

		// Invalid cases
		{name: "empty string", input: "", wantErr: true, errContains: "invalid duration"},
		{name: "missing unit", input: "30", wantErr: true, errContains: "invalid duration"},
		{name: "letters only", input: "abc", wantErr: true, errContains: "invalid duration"},
		{name: "invalid unit dd", input: "30dd", wantErr: true, errContains: "invalid duration"},
		{name: "minutes not supported", input: "30m", wantErr: true, errContains: "invalid duration"},
		{name: "negative value", input: "-1d", wantErr: true, errContains: "invalid duration"},
		{name: "trailing space", input: "30d ", wantErr: true, errContains: "invalid duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := duration.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) = %v, want error containing %q", tt.input, got, tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Parse(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
