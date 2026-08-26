# Changelog

Alle nennenswerten Änderungen an diesem Projekt.

Format nach [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
Versionierung nach [SemVer](https://semver.org/lang/de/).

Bis zum ersten öffentlichen Launch steht alles unter `[Unreleased]`.

## [Unreleased]

### Hinzugefügt

- Landingpage mit Teaser-Text und Warteliste (`GET /`, `POST /waitlist`), Absenden per htmx ohne Reload
- Registrierung, Login, Logout mit bcrypt-Passworthashes
- Passwort-Reset per E-Mail-Token (gehasht gespeichert, 60 min gültig, keine User-Enumeration)
- Session-Auth über httpOnly-Cookies mit serverseitiger Session-Tabelle
- CSRF-Schutz nach dem Double-Submit-Cookie-Muster auf allen POST-Routen
- IP-Rate-Limiting auf Login, Registrierung und Passwort-Reset
- SQLite-Anbindung (WAL, Foreign Keys) mit eingebettetem Schema
- Basis-Layout und Navigation (ungestylt, Design folgt in Phase 2)
- `Mailer`-Interface mit Konsolen-Implementierung als Platzhalter
- Deploy-Setup: Dockerfile, docker-compose, Caddy als TLS-Reverse-Proxy
- Unit-Tests für Auth und Rate-Limiter
- Entwicklungsplan, Architektur und Roadmap als Dokumente im Repo

### Bekannte Lücken

Siehe `ROADMAP.md`, Abschnitt „Jetzt“.

### Geändert

- Spielwelt: vier NRW-Editionen (Köln, Dortmund, Duisburg, Düsseldorf) statt einer
  fiktiven Stadt; zum Launch sind nur zwei offen
- Architektur vereinfacht: keine eigene Persistenzschicht, SQL steht direkt in den
  Aktionen; kein Highscore-Snapshot, kein Idempotenz-Token, keine Golden Files
