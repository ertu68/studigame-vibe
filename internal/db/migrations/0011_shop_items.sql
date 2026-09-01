CREATE TABLE shop_items (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price_cents INTEGER NOT NULL,
    category TEXT NOT NULL, -- 'equipment', 'instrument', 'housing'
    effect_type TEXT NOT NULL, -- 'passive_income', 'mood_bonus', 'container_bonus', etc.
    effect_value INTEGER NOT NULL, -- value of the effect (e.g., income per hour in cents)
    subject_required TEXT, -- null for all subjects, or specific subject
    max_quantity INTEGER NOT NULL DEFAULT 1, -- how many of this item a character can own
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_shop_items_category ON shop_items(category);
CREATE INDEX idx_shop_items_subject ON shop_items(subject_required);

-- Insert some basic shop items
-- Equipment items
INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity) VALUES
(1, 'Studentenausweis', 'Ermöglicht Studentenrabatte in der Stadt', 500, 'equipment', 'passive_income', 10, 1),
(2, 'Notebook', 'Für die Bibliothek und zum Lernen', 1500, 'equipment', 'mood_bonus', 5, 1),
(3, 'Fahrrad', 'Schneller unterwegs auf dem Campus', 3000, 'equipment', 'container_bonus', 1000, 1);

-- Instrument items (passive income)
INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity) VALUES
(4, 'Akustische Gitarre', 'Für WG-Flur-Konzerte und Straßengeschäfte', 2000, 'instrument', 'passive_income', 50, 1),
(5, 'Keyboard', 'Elektronische Klänge für elektronische Zeiten', 3500, 'instrument', 'passive_income', 75, 1);

-- Housing items (progression ladder)
INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity) VALUES
(6, 'WG-Zimmer', 'Ein eigenes Zimmer in einer WG', 10000, 'housing', 'container_bonus', 5000, 1),
(7, 'Wohnheim-Einzelzimmer', 'Eigenes Zimmer im Uni-Wohnheim', 25000, 'housing', 'container_bonus', 15000, 1),
(8, '1-Zimmer-Wohnung', 'Deine erste eigene Wohnung', 50000, 'housing', 'container_bonus', 30000, 1);