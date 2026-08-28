-- Core gameplay tables for Phase 3: character creation and economy

CREATE TABLE editions (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    slug   TEXT NOT NULL UNIQUE,
    name   TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE districts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    edition_id      INTEGER NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    yield_factor    REAL NOT NULL,      -- Multiplier for collection yield
    unlock_at_cents INTEGER NOT NULL,   -- Money required to unlock
    risk_factor     REAL NOT NULL,      -- Risk multiplier for crimes
    distance_rank   INTEGER NOT NULL,   -- For action duration calculation
    
    UNIQUE(edition_id, name)
);

CREATE TABLE characters (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    edition_id      INTEGER NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    district_id     INTEGER NOT NULL REFERENCES districts(id),
    money_cents     INTEGER NOT NULL DEFAULT 1000,    -- 10€ start capital
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
    district_id  INTEGER NOT NULL REFERENCES districts(id),
    started_at   DATETIME NOT NULL,    -- When action started
    ends_at      DATETIME NOT NULL,    -- When action completes
    resolved_at  DATETIME,             -- When resolved (NULL = pending)
    payload_json TEXT NOT NULL,        -- Action parameters as JSON
    result_json  TEXT                  -- Result when resolved
);

-- Bootstrap: Four NRW cities as editions. Köln and Dortmund launch active.
INSERT INTO editions (slug, name, active) VALUES
    ('koeln', 'Köln', TRUE),
    ('dortmund', 'Dortmund', TRUE),
    ('duisburg', 'Duisburg', FALSE),
    ('duesseldorf', 'Düsseldorf', FALSE);

-- Bootstrap: Two districts per city
-- Köln: Altstadt (easy, high yield) + Kalk (hard, risky)
INSERT INTO districts (edition_id, name, yield_factor, unlock_at_cents, risk_factor, distance_rank) VALUES
    ((SELECT id FROM editions WHERE slug='koeln'), 'Altstadt', 1.0, 0, 0.8, 1),
    ((SELECT id FROM editions WHERE slug='koeln'), 'Kalk', 1.2, 2000, 1.2, 3);

-- Dortmund: Innenstadt + Brackel
INSERT INTO districts (edition_id, name, yield_factor, unlock_at_cents, risk_factor, distance_rank) VALUES
    ((SELECT id FROM editions WHERE slug='dortmund'), 'Innenstadt', 1.1, 0, 0.9, 2),
    ((SELECT id FROM editions WHERE slug='dortmund'), 'Brackel', 1.3, 3000, 1.1, 4);

-- Duisburg: Ruhrort + Walsum (locked)
INSERT INTO districts (edition_id, name, yield_factor, unlock_at_cents, risk_factor, distance_rank) VALUES
    ((SELECT id FROM editions WHERE slug='duisburg'), 'Ruhrort', 0.9, 0, 1.3, 2),
    ((SELECT id FROM editions WHERE slug='duisburg'), 'Walsum', 1.4, 5000, 1.4, 5);

-- Düsseldorf: Oberstadt + Benrath (locked)
INSERT INTO districts (edition_id, name, yield_factor, unlock_at_cents, risk_factor, distance_rank) VALUES
    ((SELECT id FROM editions WHERE slug='duesseldorf'), 'Oberstadt', 1.5, 0, 0.7, 2),
    ((SELECT id FROM editions WHERE slug='duesseldorf'), 'Benrath', 1.6, 8000, 1.2, 4);
