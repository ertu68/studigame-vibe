# Studigame

Browser-Game im Geist von Pennergame, aber eigenständig: Studentenleben statt
Obdachlosigkeit. Go, server-rendered, SQLite, kein JS-Build. Spielwelt ist eine
fiktive Uni-Stadt, Studienfach übernimmt die Rolle des Klassensystems.

| Dokument | Inhalt |
| --- | --- |
| [studigame-plan.md](./studigame-plan.md) | Konzept, Architektur, Phasenplan |
| [ROADMAP.md](./ROADMAP.md) | was als Nächstes ansteht |
| [CHANGELOG.md](./CHANGELOG.md) | was sich geändert hat |
| [CLAUDE.md](./CLAUDE.md) | Arbeitsanweisungen für Coding-Agents |

## Stand

- **Phase 0** (Landingpage/Warteliste): `GET /`, `POST /waitlist` — htmx-Formular ohne Reload.
- **Phase 1** (Grundgerüst): SQLite-Schema, Registrierung/Login/Logout, Passwort-Reset
  (SMTP vorkonfigurierbar, fallback auf Konsolenlogging), Double-Opt-In mit E-Mail-Verifikation,
  Session-Cookies, CSRF (double-submit cookie), Rate-Limiting auf Login/Register/Reset, Basis-Layout.
- **Phase 2** (Design-System): Tailwind CSS mit benutzerdefinierten Farben, responsive
  mobile-first Layouts, wiederverwendbare Komponentenbibliothek, Statusleiste, Bottom-Navigation.
- **Phase 3** (Kernschleife, angefangen): Editionen, Bezirke, Charaktere, Kassenbuch und
  Aktionen im Schema; Charaktererstellung, Dashboard und die Sammeln-Aktion stehen.
  Schichten von 15 Minuten bis 8 Stunden, Ergebnis wird beim Zurückkommen ausgewürfelt,
  Überlauf aus dem vollen Behälter geht in den Spendentopf.

Der aktuelle Code ist noch auf dem Pennergame-Stand (vier NRW-Städte als Editionen,
Bezirke als Ertrags-/Risiko-Multiplikator). Der Umbau auf Studigame (eine fiktive
Uni-Stadt, Studienfach statt Bezirk) ist geplant, siehe `studigame-plan.md`
Abschnitt 3a und die entsprechenden Punkte in `ROADMAP.md`.

Name ist entschieden (Studigame), Domain ist noch offen — siehe `.env.example`.

## Entwicklung

```
cp .env.example .env
make run          # startet auf :8080, SQLite-Datei landet lokal
make test
make css          # baut CSS mit Tailwind
make watch        # beobachtet CSS-Änderungen
```

Die Tailwind-Standalone-CLI liegt nicht im Repo. `make css` lädt beim ersten Aufruf
das passende Binary für die eigene Plattform nach `./tailwindcss` (git-ignoriert);
der Docker-Build holt sich die Linux-Variante selbst.

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
internal/game/        reine Spielregeln, ohne DB und ohne Uhr
internal/web/         html/template-Rendering (embed)
internal/web/components/  Wiederverwendbare UI-Komponenten
internal/web/templates/  Server-gerenderte Seiten
static/css/           Tailwind CSS und kompilierte Stylesheets
static/js/            htmx Bibliothek
static/fonts/         Platz für selbst gehostete Schriften (aktuell System-Fonts)
deploy/               Caddyfile
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
