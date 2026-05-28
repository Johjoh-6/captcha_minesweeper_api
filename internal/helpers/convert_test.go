package helpers

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPgText(t *testing.T) {
	tests := []struct {
		name     string
		in       *string
		valid    bool
		expected string
	}{
		{name: "nil returns invalid", in: nil, valid: false},
		{name: "non-nil returns valid", in: ptr("hello"), valid: true, expected: "hello"},
		{name: "empty string still valid", in: ptr(""), valid: true, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := NewPgText(tt.in)
			if out.Valid != tt.valid {
				t.Fatalf("Valid=%v, want %v", out.Valid, tt.valid)
			}
			if tt.valid && out.String != tt.expected {
				t.Fatalf("String=%q, want %q", out.String, tt.expected)
			}
		})
	}
}

func TestNewPgNumeric(t *testing.T) {
	tests := []struct {
		name  string
		in    *int
		valid bool
		// If valid, we check the numeric value by converting it to *big.Int.
		expectedInt64 int64
	}{
		{name: "nil returns invalid", in: nil, valid: false},
		{name: "zero returns valid", in: ptrInt(0), valid: true, expectedInt64: 0},
		{name: "positive returns valid", in: ptrInt(123), valid: true, expectedInt64: 123},
		{name: "negative returns valid", in: ptrInt(-45), valid: true, expectedInt64: -45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := NewPgNumeric(tt.in)
			if out.Valid != tt.valid {
				t.Fatalf("Valid=%v, want %v", out.Valid, tt.valid)
			}
			if !tt.valid {
				return
			}

			if out.Int == nil {
				t.Fatalf("Numeric.Int is nil")
			}

			want := big.NewInt(tt.expectedInt64)
			if out.Int.Cmp(want) != 0 {
				t.Fatalf("Numeric.Int=%v, want %v", out.Int, want)
			}
		})
	}
}

func TestNewPgTimestamptz(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	tests := []struct {
		name  string
		in    *time.Time
		valid bool
		want  time.Time
	}{
		{name: "nil returns invalid", in: nil, valid: false},
		{name: "non-nil returns valid", in: &now, valid: true, want: now},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := NewPgTimestamptz(tt.in)
			if out.Valid != tt.valid {
				t.Fatalf("Valid=%v, want %v", out.Valid, tt.valid)
			}
			if tt.valid && !out.Time.Equal(tt.want) {
				t.Fatalf("Time=%v, want %v", out.Time, tt.want)
			}
		})
	}
}

func TestNewPgUUID(t *testing.T) {
	good := uuid.New().String()
	bad := "not-a-uuid"

	tests := []struct {
		name  string
		in    *string
		valid bool
		want  string
	}{
		{name: "nil returns invalid", in: nil, valid: false},
		{name: "bad uuid returns invalid", in: &bad, valid: false},
		{name: "good uuid returns valid", in: &good, valid: true, want: good},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NewPgUUID(tt.in)

			if tt.valid {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else {
				// nil input => no error, invalid input => error
				if tt.in == nil {
					if err != nil {
						t.Fatalf("expected nil error for nil input, got %v", err)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error for invalid UUID input")
					}
				}
			}

			if out.Valid != tt.valid {
				t.Fatalf("Valid=%v, want %v", out.Valid, tt.valid)
			}
			if !tt.valid {
				return
			}

			u, err2 := uuid.FromBytes(out.Bytes[:])
			if err2 != nil {
				t.Fatalf("uuid.FromBytes error: %v", err2)
			}
			if u.String() != tt.want {
				t.Fatalf("UUID=%q, want %q", u.String(), tt.want)
			}
		})
	}
}

func TestNewPgDate(t *testing.T) {
	d := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		in    *time.Time
		valid bool
		want  time.Time
	}{
		{name: "nil returns invalid", in: nil, valid: false},
		{name: "non-nil returns valid", in: &d, valid: true, want: d},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := NewPgDate(tt.in)
			if out.Valid != tt.valid {
				t.Fatalf("Valid=%v, want %v", out.Valid, tt.valid)
			}
			if tt.valid && !out.Time.Equal(tt.want) {
				t.Fatalf("Time=%v, want %v", out.Time, tt.want)
			}
		})
	}
}

func ptr(s string) *string { return &s }
func ptrInt(i int) *int    { return &i }
