package helpers

import (
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func NewPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func NewPgNumeric(n *int) pgtype.Numeric {
	if n == nil {
		return pgtype.Numeric{Valid: false}
	}

	// pgtype.Numeric represents a base-10 number as Int * 10^Exp.
	// For an integer input, Exp should be 0 and Int should be the integer value.
	// (NaN/Infinity not supported here.)
	return pgtype.Numeric{Int: big.NewInt(int64(*n)), Exp: 0, Valid: true}
}

func NewPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func NewPgUUID(u *string) (pgtype.UUID, error) {
	if u == nil {
		return pgtype.UUID{Valid: false}, nil
	}

	parsed, err := uuid.Parse(*u)
	if err != nil {
		return pgtype.UUID{Valid: false}, err
	}

	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

func NewPgDate(d *time.Time) pgtype.Date {
	if d == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *d, Valid: true}
}
