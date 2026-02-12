package postgres

import (
	"errors"
	"log/slog"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// translateError converts database errors to domain errors.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	// Log unexpected database errors
	slog.Error("database error",
		slog.String("error", err.Error()),
	)

	return err
}

// uuidToPgtype converts a google/uuid to pgtype.UUID.
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// pgtypeToUUID converts a pgtype.UUID to google/uuid.
func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

// pgtypeToTime converts pgtype.Timestamp to *time.Time.
func pgtypeToTime(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// stringToPgtype converts a string to pgtype.Text.
func stringToPgtype(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// pgtypeToString converts a pgtype.Text to string.
func pgtypeToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// pgtypeToStringPtr converts a pgtype.Text to *string.
func pgtypeToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// pgtypeToTimePtr is an alias for pgtypeToTime for consistency naming.
func pgtypeToTimePtr(ts pgtype.Timestamp) *time.Time {
	return pgtypeToTime(ts)
}
