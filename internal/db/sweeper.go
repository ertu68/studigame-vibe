// Package db provides database operations including cleanup functions.
package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// SweepExpired removes expired sessions and password reset tokens from the database.
// It should be called periodically (e.g., every minute) to prevent the tables from growing indefinitely.
func SweepExpired(ctx context.Context, db *sql.DB) error {
	now := time.Now()

	// Delete expired sessions
	sessionsResult, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, now)
	if err != nil {
		return err
	}
	sessionsCount, _ := sessionsResult.RowsAffected()

	// Delete expired password resets
	passwordResetsResult, err := db.ExecContext(ctx, `DELETE FROM password_resets WHERE expires_at < ?`, now)
	if err != nil {
		return err
	}
	passwordResetsCount, _ := passwordResetsResult.RowsAffected()

	// Delete expired verification tokens
	verificationResult, err := db.ExecContext(ctx, `DELETE FROM verification_tokens WHERE expires_at < ?`, now)
	if err != nil {
		return err
	}
	verificationCount, _ := verificationResult.RowsAffected()

	// Log cleanup results
	if sessionsCount > 0 || passwordResetsCount > 0 || verificationCount > 0 {
		slog.Info("Sweeper cleaned up expired records",
			"sessions", sessionsCount,
			"passwordResets", passwordResetsCount,
			"verificationTokens", verificationCount)
	}

	return nil
}
