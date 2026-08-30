-- Core gameplay tables for Phase 3: character creation and economy.
-- Edition = a game world (MVP: one fictional uni-city). District has no
-- gameplay role anymore — Studienfach (subject) is the class system instead.

CREATE TABLE editions (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    slug   TEXT NOT NULL UNIQUE,
    name   TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE subjects (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    edition_id INTEGER NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,

    UNIQUE(edition_id, name)
);

CREATE TABLE characters (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    edition_id      INTEGER NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    subject_id      INTEGER NOT NULL REFERENCES subjects(id),
    money_cents     INTEGER NOT NULL DEFAULT 1000,    -- 10 € start capital
    container_level INTEGER NOT NULL DEFAULT 0,       -- 0-4 for 5 tiers
    mood            INTEGER NOT NULL DEFAULT 100,     -- 0-100 scale
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, edition_id),   -- Max 1 character per edition per user
    UNIQUE(edition_id, name)       -- Name unique within edition
);

CREATE TABLE ledger (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id   INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    delta_cents    INTEGER NOT NULL,     -- Positive for credit, negative for debit
    balance_after  INTEGER NOT NULL,     -- Balance snapshot after this txn
    reason         TEXT NOT NULL,        -- Reason for transaction
    ref_id         INTEGER,              -- Reference to related record (action, etc.)
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE actions (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    kind         TEXT NOT NULL,        -- Type of action (collect, crime, etc.)
    started_at   DATETIME NOT NULL,    -- When action started
    ends_at      DATETIME NOT NULL,    -- When action completes
    resolved_at  DATETIME,             -- When resolved (NULL = pending)
    payload_json TEXT NOT NULL,        -- Action parameters as JSON
    result_json  TEXT                  -- Result when resolved
);

-- Bootstrap: one fictional uni-city as edition, one active world.
-- Studistadt is the final MVP uni-city (single edition).
INSERT INTO editions (slug, name, active) VALUES
    ('studistadt', 'Studistadt', TRUE);

-- Studienfächer for the MVP class system (see studigame-plan, Abschnitt 3).
INSERT INTO subjects (edition_id, name) VALUES
    ((SELECT id FROM editions WHERE slug='studistadt'), 'Informatik'),
    ((SELECT id FROM editions WHERE slug='studistadt'), 'BWL'),
    ((SELECT id FROM editions WHERE slug='studistadt'), 'Jura'),
    ((SELECT id FROM editions WHERE slug='studistadt'), 'Geisteswissenschaften');
