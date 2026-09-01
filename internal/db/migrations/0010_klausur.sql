-- Add ECTS tracking for the Klausur action.
-- Stored as INTEGER cents (123 = 1.23 ECTS) for exact decimal arithmetic.

ALTER TABLE characters ADD COLUMN ects_cents INTEGER NOT NULL DEFAULT 0;

-- Index for fast ECTS lookups (e.g. leaderboards).
CREATE INDEX idx_characters_ects ON characters(ects_cents DESC);