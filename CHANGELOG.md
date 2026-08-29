# Changelog

Alle nennenswerten Änderungen an diesem Projekt.

Format nach [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
Versionierung nach [SemVer](https://semver.org/lang/de/).

Bis zum ersten öffentlichen Launch steht alles unter `[Unreleased]`.

## [Unreleased]

### Hinzugefügt

- Uni-Stadt bei der Registrierung (Pflichtfeld, feste Liste deutscher Uni-Städte), erscheint neben dem Namen — bisher rein zur Anzeige, ohne Spielfunktion
- Landingpage mit Teaser-Text und Warteliste (`GET /`, `POST /waitlist`), Absenden per htmx ohne Reload
- Registrierung, Login, Logout mit bcrypt-Passworthashes
- Passwort-Reset per E-Mail-Token (gehasht gespeichert, 60 min gültig, keine User-Enumeration)
- Session-Auth über httpOnly-Cookies mit serverseitiger Session-Tabelle
- CSRF-Schutz nach dem Double-Submit-Cookie-Muster auf allen POST-Routen
- IP-Rate-Limiting auf Login, Registrierung und Passwort-Reset
- SQLite-Anbindung (WAL, Foreign Keys) mit eingebettetem Schema
- Basis-Layout und Navigation (ungestylt, Design folgt in Phase 2)
- `Mailer`-Interface mit Konsolen- und SMTP-Implementierung (SMTP via env vars konfigurierbar)
- Double-Opt-In für Registrierung: Accounts mit `verified=FALSE` bis zur E-Mail-Bestätigung, `GET /verify?token=...` aktiviert sie
- Spielwelt in der Datenbank: Editionen, Bezirke, Charaktere, Kassenbuch und Aktionen;
  Köln und Dortmund sind offen, Duisburg und Düsseldorf vorbereitet
- Charaktererstellung unter `/character/new`: Stadt wählen, Name setzen, 10 € Startkapital.
  Ein Charakter pro Stadt, Namen sind innerhalb einer Stadt eindeutig
- Dashboard zeigt Kasse, Laune und Bezirk des Charakters
- `internal/game`: reine Spielregeln für Sammelertrag, Behältergrößen und Charity-Überlauf,
  ohne Datenbank und ohne eigene Uhr
- Sammeln-Aktion unter `/actions`: Schichtlänge wählen (15 min bis 8 h), das Ergebnis wird
  erst beim Zurückkommen ausgewürfelt. Eine Schicht pro Charakter, keine Doppelauszahlung
- Geld-Behälter mit Kappung: Was nicht mehr hineinpasst, geht in den Spendentopf der Stadt
- Laufende Schicht zeigt Restzeit und Rückkehrzeit; die Seite lädt sich per Meta-Refresh
  genau einmal neu, wenn die Schicht fällig ist. Dazu ein Knopf zum manuellen Nachsehen
- Deploy-Setup: Dockerfile, docker-compose, Caddy als TLS-Reverse-Proxy
- Unit-Tests für Auth und Rate-Limiter
- Entwicklungsplan, Architektur und Roadmap als Dokumente im Repo
- Structured Logging mit `log/slog` in JSON-Format
- Panik-Handling-Middleware für Serverstabilität
- Security-Header (X-Content-Type-Options, X-Frame-Options, etc.)
- Lokales htmx-Serving statt CDN (CSP-kompatibel)
- Periodischer Sweeper für abgelaufene Sessions und Reset-Tokens
- Gesundheits-Endpoint `/healthz`
- CI/CD Pipeline mit `go vet`, `go test -race`
- Backup-Skript für SQLite-Datenbank
- **Design-System**:
  - Tailwind CSS mit benutzerdefinierten Farben und Schriftarten
  - Wiederverwendbare Komponenten (Buttons, Cards, Formulare, etc.)
  - Mobile-first Layout mit Bottom-Navigation
  - Sticky Statusleiste mit Spielerinformationen
  - Komponentenbibliothek (button, card, form_field, flash, progress, statusbar, bottom_nav)

### Behoben

- `Dockerfile` auf Go 1.27 (D1)
- CSRF + Rate-Limit auf `POST /waitlist` (D3)
- Sessions beim Passwort-Reset invalidieren (D4)
- `Register`: nur UNIQUE-Verletzung als „E-Mail vergeben“ werten (D5)
- Migrations-System (D11)
- Tailwind-Build lud `tailwind.config.js` nie (fehlende `@config`-Direktive), dadurch
  waren alle Custom-Farben und Komponentenklassen (`.card`, `.statusbar`,
  `.bottom-nav`, `.btn-primary`, `.form-input`, `.flash-*`, `.progress-*`) im
  ausgelieferten CSS nicht vorhanden — Seiten liefen unstyled aus
- Fehlender Logout-Button in der UI (Route existierte, aber kein Link/Formular
  dazu in Templates)
- Flash-Meldung wurde auf Login/Register/Reset doppelt gerendert (einmal im
  Layout, einmal in der Seite selbst)
- Docker-Build schlug fehl, weil das mitgelieferte `tailwindcss`-Binary nur auf
  macOS lief; Build und `make css` laden jetzt die passende Version selbst
- Überschriften und Fließtext nutzten Schrift-Klassen, die es im CSS nicht gab
- „Passwort zuruecksetzen" und „Passwort geändert erfolgreich" korrigiert

### Geändert

- Düsseldorf ist als dritte Edition offen; nur Duisburg ist noch geschlossen

- Navigation und Statusleiste ohne Emoji, mit Textbeschriftungen
- Nach dem Passwort-Reset führt ein normaler Link zum Login statt eines Buttons
  mit Inline-JavaScript; die Content-Security-Policy erlaubt keine Inline-Skripte mehr
- Migration `0002_phase3_tables.sql` gelöscht: Die Kernschleife hat noch keine Go-Handler,
  die das Schema nutzen. Nach MVP-Regel wird es bereitgestellt, wenn es reif ist. Alte
  Test-DBs werden zurückgesetzt (keine Nutzer-DBs existieren noch).
- Registrierung sendet Bestätigungslink statt direkten Login; `Authenticate` lehnt unverifizierte Accounts ab

- Spielwelt: vier NRW-Editionen (Köln, Dortmund, Duisburg, Düsseldorf) statt einer
  fiktiven Stadt; zum Launch sind nur zwei offen
- Architektur vereinfacht: keine eigene Persistenzschicht, SQL steht direkt in den
  Aktionen; kein Highscore-Snapshot, kein Idempotenz-Token, keine Golden Files
- Datenbank-Migrationssystem hinzugefügt (`internal/db/migrations`) mit Embed-FS
- HTTP-Server mit Timeouts und Graceful Shutdown
- CSP zur Verstärkung der Sicherheit
