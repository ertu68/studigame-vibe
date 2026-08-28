-- Charity pool: money that overflows a full container is donated, per edition.
CREATE TABLE charity_pool (
    edition_id INTEGER PRIMARY KEY REFERENCES editions(id) ON DELETE CASCADE,
    cents      INTEGER NOT NULL DEFAULT 0
);

INSERT INTO charity_pool (edition_id, cents) SELECT id, 0 FROM editions;

-- One unresolved action per character at a time.
CREATE UNIQUE INDEX idx_actions_open ON actions(character_id) WHERE resolved_at IS NULL;
