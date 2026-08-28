-- Add email verification for Double-Opt-In

-- Track account verification status
ALTER TABLE users ADD COLUMN verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Store verification tokens (similar to password_resets)
CREATE TABLE verification_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Track waitlist verification
ALTER TABLE waitlist ADD COLUMN verified BOOLEAN NOT NULL DEFAULT FALSE;
