-- Heimatstadt als Profil-Anzeige, ohne Spielfunktion bisher.
ALTER TABLE users ADD COLUMN city TEXT NOT NULL DEFAULT '';
