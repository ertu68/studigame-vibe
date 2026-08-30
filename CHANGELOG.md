# Changelog

Alle nennenswerten Änderungen an diesem Projekt.

Format nach [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
Versionierung nach [SemVer](https://semver.org/lang/de/).

Bis zum ersten öffentlichen Launch steht alles unter `[Unreleased]`.

## [Unreleased]

### Geändert

- **Rebrand Pennergame → Studigame**: Go-Modul, alle Imports und der
  DB-Dateiname heißen jetzt `studigame`; Backup-Skript, docker-compose und
  IDE-Dateien nachgezogen
- **Bezirke durch Studienfächer ersetzt**: Die `districts`-Tabelle
  (Ertrags-/Risiko-Multiplikator) ist entfernt, `subjects` (Studienfach)
  eingeführt; die Charaktererstellung fragt jetzt das Studienfach ab statt
  einer Stadt. Bis das Skillsystem steht, ist das Studienfach rein Flavor
  (siehe `studigame-plan.md` §3b)
- **Spielwelt**: statt vier paralleler NRW-Editionen startet das Spiel mit
  einer fiktiven Uni-Stadt (**„Studistadt“**); die `editions`-Tabelle
  bleibt technisch für spätere Uni-Städte
- **Landingpage**: Teaser-Text neu formuliert – beschreibt das
  Studentenleben-Spiel auf eigene Faust statt als Pennergame-Hommage
- **Behälter-Überlauf**: Was nicht mehr in den Geld-Behälter passt, geht
  automatisch in den Spendentopf der Uni-Stadt statt zu verfallen
- **Dokumentation**: Plan, Roadmap und README auf den aktuellen Stand gebracht.
  Der Entwicklungsplan ist final aufs MVP geschnitten (siehe §3b) – klar gestaffelte
  Kernschleife: Sammeln + Behälter stehen, Behälter-Upgrade als erster Geld-Sink
  kommt als Nächstes, dann Balancing, dann scharf spielen.
  DaisyUI als bewusst verworfen gekennzeichnet, tote Plan-Verweise entfernt.
  Der Plan ist **implementierungsfertig**: alle MVP-Entscheidungen sind getroffen
  (Behälter-Upgrade als Geld-Sink: sofortiger Kauf, überlineare Kostenkurve
  `Basiswert × Level^1.8` mit Basiswert 2000 ct, Max-Stufe 4; Uni-Stadt
  „Studistadt“ und die 4 Studienfächer final; Kosten und Formeln in §3b).
  Offen bleiben nur Pre-Launch-Ops (Domain, VPS, Mailversand, Rechtliches) –
  die erledigt der Entwickler später beim Deploy

### Hinzugefügt

- Landingpage mit Teaser-Text und Warteliste (`GET /`, `POST /waitlist`),
  Absenden per htmx ohne Reload
- Registrierung, Login, Logout mit bcrypt-Passworthashes, Passwort-Reset und
  Double-Opt-In per E-Mail-Verifikation
- Session-Auth über httpOnly-Cookies mit serverseitiger Session-Tabelle
- CSRF-Schutz (Double-Submit-Cookie) auf allen POST-Routen; IP-Rate-Limiting
  auf Login, Registrierung und Passwort-Reset
- SQLite-Anbindung (WAL, Foreign Keys) mit eingebettetem Schema und
  Migrations-System (`internal/db/migrations`)
- Einheiten und Aktionen im Schema: `editions`, `subjects`, `characters`,
  `ledger`, `actions`
- Charaktererstellung unter `/character/new`: Studienfach wählen, Name setzen,
  10 € Startkapital; ein Charakter pro Edition, Name je Edition eindeutig
- Dashboard unter `/` mit Kasse, Laune und Studienfach des Charakters
- `internal/game`: reine Spielregeln für Sammelertrag, Behältergrößen und
  Charity-Überlauf, ohne Datenbank und ohne eigene Uhr
- Sammeln-Aktion unter `/actions`: Schichtlänge wählen (15 min bis 8 h), das
  Ergebnis wird erst beim Zurückkommen ausgewürfelt; eine Schicht pro Charakter,
  keine Doppelauszahlung
- Laufende Schicht zeigt Restzeit und Rückkehrzeit; die Seite lädt sich per
  Meta-Refresh genau einmal neu, wenn die Schicht fällig ist
- `Mailer`-Interface mit Konsolen- und SMTP-Implementierung (SMTP via Env-Vars)
- Deploy-Setup: Dockerfile, docker-compose, Caddy als TLS-Reverse-Proxy
- Unit-Tests für Auth, Rate-Limiter und die Spielregeln
- Structured Logging mit `log/slog`, Panik-Handling-Middleware,
  Security-Header, lokales htmx-Serving (CSP-kompatibel), `/healthz`,
  Session-Sweeper, Backup-Skript, CI/CD-Pipeline (`go vet`, `go test -race`)
- **Design-System**: Tailwind mit benutzerdefinierten Farben, mobile-first
  Layout mit Bottom-Navigation, Sticky-Statusleiste, wiederverwendbare
  Komponenten (button, card, form_field, flash, progress, statusbar, bottom_nav)

### Behoben

- Dockerfile auf Go 1.27 (D1)
- CSRF + Rate-Limit auf `POST /waitlist` (D3)
- Sessions beim Passwort-Reset invalidieren (D4)
- `Register` nur UNIQUE-Verletzung als „E-Mail vergeben“ werten (D5)
- Migrations-System (D11)
- Tailwind-Build lud `tailwind.config.js` nie (fehlende `@config`-Direktive),
  dadurch waren Custom-Farben und Komponentenklassen nicht im CSS – Seiten
  liefen unstyled aus
- Fehlender Logout-Button in der UI; doppelt gerenderte Flash-Meldung auf
  Login/Register/Reset
- Docker-Build schlug fehl, weil das mitgelieferte `tailwindcss`-Binary nur auf
  macOS lief; Build und `make css` laden jetzt die passende Version selbst
- Überschriften/Fließtext nutzten Schrift-Klassen, die es im CSS nicht gab
- Nach dem Passwort-Reset führt ein normaler Link zum Login statt Inline-JS;
  die Content-Security-Policy erlaubt keine Inline-Skripte mehr
