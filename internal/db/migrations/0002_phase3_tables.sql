-- Phase 3 tables for core gameplay loop
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
    unlock_at_cents INTEGER NOT NULL,   -- Money required to unlock this district
    risk_factor     REAL NOT NULL,      -- Risk multiplier for crimes
    distance_rank   INTEGER NOT NULL    -- Used for fight duration calculation
);

CREATE TABLE characters (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    edition_id      INTEGER NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    district_id     INTEGER NOT NULL REFERENCES districts(id),
    money_cents     INTEGER NOT NULL DEFAULT 1000,  -- 10€ start capital
    container_level INTEGER NOT NULL DEFAULT 0,     -- 0-4 for 5 container tiers
    cleanliness     INTEGER NOT NULL DEFAULT 100,   -- 0-100 scale
    mood            INTEGER NOT NULL DEFAULT 100,   -- 0-100 scale
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, edition_id),  -- Max 1 character per edition per user
    UNIQUE(edition_id, name)      -- Name unique within edition
);

CREATE TABLE character_skills (
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    skill        TEXT NOT NULL,    -- Skill name/identifier
    level        INTEGER NOT NULL DEFAULT 0,
    
    PRIMARY KEY (character_id, skill)
);

CREATE TABLE skill_queue (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    skill       TEXT NOT NULL,
    ends_at     DATETIME NOT NULL,     -- When training completes
    cost_cents  INTEGER NOT NULL,      -- Cost of training
    slot        INTEGER NOT NULL,      -- 0 or 1 for 2 training slots
    
    CHECK (slot IN (0, 1))
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

CREATE TABLE inventory (
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_id      TEXT NOT NULL,        -- Item identifier
    qty          INTEGER NOT NULL DEFAULT 0,
    
    PRIMARY KEY (character_id, item_id)
);

CREATE TABLE ledger (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id   INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    delta_cents    INTEGER NOT NULL,     -- Positive for credit, negative for debit
    balance_after  INTEGER NOT NULL,     -- Balance after this transaction
    reason         TEXT NOT NULL,        -- Reason for transaction
    ref_id         INTEGER,              -- Reference to related record
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE fights (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    attacker_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    defender_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    winner_id   INTEGER REFERENCES characters(id),
    loot_cents  INTEGER NOT NULL DEFAULT 0,
    log_json    TEXT NOT NULL          -- Fight log details as JSON
);

CREATE TABLE fight_cooldowns (
    attacker_id INTEGER NOT NULL,
    defender_id INTEGER NOT NULL,
    until       DATETIME NOT NULL,     -- Cooldown expires at this time
    
    PRIMARY KEY (attacker_id, defender_id),
    FOREIGN KEY (attacker_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (defender_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE gangs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    bank_cents   INTEGER NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gang_members (
    gang_id    INTEGER NOT NULL REFERENCES gangs(id) ON DELETE CASCADE,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'member',  -- 'chief' or 'member'
    joined_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (gang_id, character_id)
);

CREATE TABLE messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    from_id    INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    to_id      INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    subject    TEXT NOT NULL,
    body       TEXT NOT NULL,
    read_at    DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE referrals (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id  INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    code          TEXT NOT NULL UNIQUE,
    clicks        INTEGER NOT NULL DEFAULT 0,
    last_click_at DATETIME
);

CREATE TABLE charity_pool (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    cents      INTEGER NOT NULL,
    source     TEXT NOT NULL,         -- Source of donation (e.g., 'container_overflow')
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_actions_character_resolved ON actions(character_id, resolved_at);
CREATE INDEX idx_ledger_character_created ON ledger(character_id, created_at);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
CREATE INDEX idx_fight_cooldowns_until ON fight_cooldowns(until);
CREATE INDEX idx_characters_edition_money ON characters(edition_id, money_cents);
CREATE INDEX idx_districts_edition ON districts(edition_id);