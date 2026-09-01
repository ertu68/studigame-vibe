CREATE TABLE character_items (
    id INTEGER PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES shop_items(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    purchased_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(character_id, item_id)
);

CREATE INDEX idx_character_items_character ON character_items(character_id);
CREATE INDEX idx_character_items_item ON character_items(item_id);