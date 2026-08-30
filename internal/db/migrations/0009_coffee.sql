-- Campus-Kiosk coffee: a consumable Konzentrations-Boost that halves skill
-- training while active. caffeine_until is the end of the window; NULL means
-- no boost. Buying more coffee extends the window (stacks). Migrations run in
-- order for fresh databases too, so this ALTER covers both upgrades and new DBs.

ALTER TABLE characters ADD COLUMN caffeine_until TEXT NULL;
