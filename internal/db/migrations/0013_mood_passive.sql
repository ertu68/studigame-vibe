-- Mood regenerates from mood_at (the moment the stored mood was last settled).
-- passive_paid_at anchors the hourly income of passive_income shop items; NULL
-- until the first such item is bought. subject_required was never used.

ALTER TABLE characters ADD COLUMN mood_at TEXT NULL;
ALTER TABLE characters ADD COLUMN passive_paid_at TEXT NULL;
UPDATE characters SET mood_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now');

DROP INDEX idx_shop_items_subject;
ALTER TABLE shop_items DROP COLUMN subject_required;
