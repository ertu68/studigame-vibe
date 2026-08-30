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
  Komponenten sind handgebaute Tailwind-Partials, bewusst ohne DaisyUI (neue Dependency für
  funktionierenden Code wäre Rückschritt).
- **Phase 3** (Kernschleife): Editionen, Studienfächer, Charaktere, Kassenbuch und
  Aktionen im Schema; Charaktererstellung, Dashboard und die Aktionen stehen.
  Schichten von 15 Minuten bis 8 Stunden, Ergebnis wird beim Zurückkommen ausgewürfelt,
  Überlauf aus dem vollen Behälter geht in den Spendentopf. Zwei Aktionen: das
  Flaschensammeln (Glück + Behälter) und der **Werkstudenten-Job** (fester Stundenlohn),
  aufgelöst über einen `kind`-Dispatch mit gemeinsamem Buchungspfad. Der erste Geld-Sink
  steht: **Behälter-Upgrade** als sofortiger Kauf (überlineare Kostenkurve
  `2000 × Level^1.8` Cent, Max-Stufe 4), Knopf auf dem Dashboard, Ledger-Debit in einer
  Transaktion.

Der Code ist auf dem Studigame-Stand: ein Go-Modul `studigame`, eine fiktive
Uni-Stadt („Studistadt“) als Edition, Studienfach statt Bezirk als
Klassensystem. Details zum Konzept in `studigame-plan.md` (Abschnitt 3, 3a, 3b).

Spielregeln leben infra-frei in `internal/game`, Aktionen in
`internal/handlers/*_store.go` — der Plan-Entwurf von getrenntem `internal/actions`
ist verworfen, „eine Datei pro Aktion“ hat sich in `handlers` niedergeschlagen.

Aktueller Kern = **Sammeln + Werkstudenten-Job mit Geld-Behälter und
Charity-Überlauf plus Behälter-Upgrade als erster Geld-Sink**. Charaktererstellung
fragt das Studienfach (bisher rein Flavor, ohne Spieleffekt). Als Nächstes: die
laufende Aktion live im Browser runterzählen (Alpine.js) und dann das Balancing
spielen und bewerten (siehe `studigame-plan.md` §3b und ROADMAP).

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

Spielregeln (Sammelertrag, Behältergrößen, Charity-Split) liegen rein in
`internal/game` — ohne DB und ohne Uhr, tabellengetestet.

## Mitarbeiten

Diese README ist die Dokumentation des Projekts und wird **mit jeder Änderung
mitgezogen**, nicht hinterher. Konkret gehört zu jeder Aufgabe:

1. `go build ./... && go vet ./... && go test ./...` grün
2. README anpassen, wenn sich Struktur, Kommandos, Env-Vars oder Funktionsumfang ändern
3. Eintrag in `CHANGELOG.md` unter `[Unreleased]`
4. erledigte Punkte in `ROADMAP.md` abhaken

Nach Abschluss einer Phase zusätzlich: Abschnitt „Stand“ auf die neue Phase heben.
Die ausführlichen Regeln stehen in [CLAUDE.md](./CLAUDE.md).
