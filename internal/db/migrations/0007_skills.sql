-- Skills and skill training for the MVP skill tree.
-- skills holds the current level of each known skill for a character.
-- skill_trainings holds at most one in-progress training per character (MVP:
-- one skill at a time, plan §3b reduced variant). Money is debited up front
-- and 80 % refunded on cancel; levelling up on completion only bumps the level.

CREATE TABLE skills (
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    skill        TEXT NOT NULL,
    level        INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, skill)
);

CREATE TABLE skill_trainings (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL UNIQUE REFERENCES characters(id) ON DELETE CASCADE,
    skill        TEXT NOT NULL,
    level        INTEGER NOT NULL,   -- the level being trained to (current + 1)
    started_at   DATETIME NOT NULL,
    ends_at      DATETIME NOT NULL,
    cost_cents   INTEGER NOT NULL
);
