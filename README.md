# Pennergame Remake (Arbeitstitel)

Browser-Game im Stil von Pennergame: Go, server-rendered, SQLite, kein JS-Build.
Spielwelt sind vier NRW-Städte als getrennte Editionen.

| Dokument | Inhalt |
| --- | --- |
| [pennergame-remake-plan.md](./pennergame-remake-plan.md) | Konzept, Architektur, Phasenplan |
| [ROADMAP.md](./ROADMAP.md) | was als Nächstes ansteht |
| [CHANGELOG.md](./CHANGELOG.md) | was sich geändert hat |
| [CLAUDE.md](./CLAUDE.md) | Arbeitsanweisungen für Coding-Agents |

## Stand

- **Phase 0** (Landingpage/Warteliste): `GET /`, `POST /waitlist` — htmx-Formular ohne Reload.
- **Phase 1** (Grundgerüst): SQLite-Schema, Registrierung/Login/Logout, Passwort-Reset
  (Mailer als Platzhalter, loggt statt zu versenden), Session-Cookies, CSRF (double-submit
  cookie), Rate-Limiting auf Login/Register/Reset, Basis-Layout, Deploy-Setup (Docker + Caddy).

Name, Domain und echter Mailversand sind noch offen — siehe `.env.example`.

## Entwicklung

```
cp .env.example .env
make run          # startet auf :8080, SQLite-Datei landet lokal
make test
```

## Deploy

```
docker compose up -d --build
```

`docker-compose.yml` baut die App und startet Caddy als TLS-terminierenden Reverse Proxy.
`SITE_DOMAIN` in `.env` setzen, bevor es auf einem echten Server läuft.

## Struktur

```
cmd/server/          main() — Routing, Wiring
internal/config/      Env-Var-Konfiguration
internal/db/          SQLite-Verbindung + Schema
internal/auth/        Sessions, CSRF, Registrierung/Login/Reset, Mailer-Interface
internal/httpx/       generisches HTTP-Middleware (Rate-Limiter)
internal/handlers/    HTTP-Handler, verbindet auth/db mit Templates
internal/web/         html/template-Rendering (embed)
internal/web/templates/  Server-gerenderte Seiten (noch ungestyled, Phase 2 macht Design)
static/                Statische Assets (CSS-Platzhalter)
deploy/                Caddyfile
```

Ab Phase 3 kommen `internal/game` (reine Spielregeln, ohne DB und Uhr) und
`internal/actions` (eine Datei pro Spieleraktion, je eine Transaktion) dazu.

## Mitarbeiten

Diese README ist die Dokumentation des Projekts und wird **mit jeder Änderung
mitgezogen**, nicht hinterher. Konkret gehört zu jeder Aufgabe:

1. `go build ./... && go vet ./... && go test ./...` grün
2. README anpassen, wenn sich Struktur, Kommandos, Env-Vars oder Funktionsumfang ändern
3. Eintrag in `CHANGELOG.md` unter `[Unreleased]`
4. erledigte Punkte in `ROADMAP.md` abhaken

Nach Abschluss einer Phase zusätzlich: Abschnitt „Stand“ auf die neue Phase heben.
Die ausführlichen Regeln stehen in [CLAUDE.md](./CLAUDE.md).
