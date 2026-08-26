# Pennergame Remake (Arbeitstitel)

Siehe [pennergame-remake-plan.md](./pennergame-remake-plan.md) fuer Konzept und Phasenplan.

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
