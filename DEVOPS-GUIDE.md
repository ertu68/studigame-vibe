# Studigame DevOps-Guide (für dich)

Dein Lernpfad, während der Agent den MVP-Code fertigstellt. Du machst **alles
außer dem Spielcode**: Git-Prozess, CI/CD, Container, Server, Deployment,
Monitoring, Security.

Dieses Dokument ist der **Umsetzungs-Fahrplan**. Die ausführliche Theorie und
das „warum" steht in [STUDIGAME-GUIDE.md](./STUDIGAME-GUIDE.md) (Abschnitte 5–13).
Hier steht die **Reihenfolge**: was du wann tust, und warum gerade jetzt.

Legende: 🟢 jetzt sofort · 🟡 sobald MVP-Code grün · 🔵 Lern-Kür (nicht
launch-kritisch) · 🙋 manuell, kostet Geld oder Rechtsbezug

---

## 0. Deine drei Fragen, direkt beantwortet

**„Soll ich jetzt schon einen Server mieten?"**
Ja — bei dir lohnt es sich sofort. Für reines Code-Lernen bräuchtest du ihn nicht
(CI/CD und Docker laufen lokal, k8s via kind/k3d umsonst). Aber dein Ziel ist
**DevOps lernen**, und „auf echtem Blech betreiben" ist eine eigene Fähigkeit:
Linux härten, DNS, TLS, SSH, systemd, einen laufenden Dienst beobachten — das
übst du lokal gar nicht. Bei 5 €/Monat als Lern-Budget schlägt der Übungswert
die Kosten klar. **Empfehlung:** Hetzner CAX11 (ARM, ~4 €/Monat, dein Image baut
ohnehin multi-arch), monatlich kündbar. Miete ihn **in Phase A**, parallel zum
Agent. Wichtig: Server mieten und aufsetzen ist erlaubt, **bevor** das Recht
steht — nur das öffentliche Schalten des E-Mail-Formulars wartet auf Impressum +
Datenschutz (Phase B).

**„Kann ich mit CI/CD schon anfangen, obwohl nichts live ist?"**
Ja — genau richtig. CI ist in der Entwicklung am wertvollsten, nicht erst beim
Launch. Es gibt sogar schon `.github/workflows/go.yml` (Build + vet + `go test
-race`). Das ist dein Startpunkt, den du ausbaust. CD (Deployment) kommt später,
CI baust du **ab heute**.

**„Kann ich die ‚Kommt bald'-Seite jetzt live schalten um E-Mails zu sammeln?"**
Ja, aber mit zwei Vorbedingungen, die nicht verhandelbar sind:

1. 🙋 **Impressum + Datenschutzerklärung** müssen stehen (Pflicht ab der ersten
   öffentlichen Seite mit E-Mail-Feld, §5 DDG + Art. 13 DSGVO).
2. 🙋 **Double-Opt-In** für die Warteliste (Code steht schon) + ein echter
   Mailversender mit SPF/DKIM, sonst landen die Bestätigungsmails im Spam.

Technisch kann die Landing-Seite sofort deployt werden — der Code (`GET /`,
`POST /waitlist`) läuft. Es ist also kein Code-, sondern ein reines Rechts-Thema.
Der Server und der Deploy stehen schon in Phase A; in Phase B kommt nur das
Rechtliche dazu und dann schaltest du das Formular öffentlich.

---

## Der Fahrplan in vier Phasen

```
Phase A  Fundament: Git, CI, Image, Server, Deploy 🟢 jetzt, parallel zum Agent
Phase B  "Kommt bald" öffentlich schalten          🟡 wenn Recht + Mail stehen
Phase C  Kubernetes-Lernpfad (k3s)                 🔵 Kür, nach dem MVP-Launch
Phase D  Security, Monitoring, Betrieb vertiefen   🔵 Dauerlauf
```

Server steckt bewusst in Phase A: er ist dein wichtigstes Lern-Objekt, nicht erst
ein Launch-Schritt. Du betreibst die App privat (nur-du, per IP oder
Test-Subdomain), lange bevor sie öffentlich beworben wird.

Du musst nicht auf den Agent warten. Phase A läuft komplett parallel: der Agent
schreibt Spielcode, du baust die Pipeline drumherum.

---

## Phase A — Fundament (🟢 jetzt, parallel zum Agent, ~5 €/Monat)

Ziel: sauberer Git-Prozess, grüne CI-Pipeline, ein Image in der Registry **und**
ein echter Server, auf dem die App privat läuft. Alles außer „öffentlich
beworben" passiert hier.

### A1 — Git-Prozess aufsetzen

Der Code liegt schon in Git, aber noch ohne Prozess. Richte ein:

- [ ] **Branch-Schutz auf `main`** (GitHub → Settings → Rules → Rulesets):
  PR erzwingen, CI-Check als „required", linear history, kein force-push.
- [ ] **Trunk-based Workflow** üben: kurze Branches `feat/…`, `ci/…`, `docs/…`,
  → PR → squash-merge. `main` ist immer deploybar. (ADR-1 im Guide.)
- [ ] **Conventional Commits**: `feat(ci): add trivy scan`, `fix(deploy): …`.
- [ ] **Signierte Commits** (SSH-Key): `git config commit.gpgsign true` +
  `gpg.format ssh`. GitHub zeigt dann „Verified".
- [ ] `.gitattributes` mit `* text=lf`, `.editorconfig` anlegen.
- [ ] **Niemals** committen: `*.db`, `.env`, `tailwindcss`-Binary. Prüfen:
  `git check-ignore studigame.db .env tailwindcss` muss alle drei nennen.

> Lern-Übung (Guide §5): Schreibe die Issues für den nächsten Milestone **bevor**
> du anfängst, schätze S/M/L, vergleiche hinterher mit der Realität.

### A2 — CI ausbauen (die bestehende `go.yml` ist der Keim)

Baue die Pipeline in **einzelnen PRs** aus, jeder mit Changelog-Eintrag. So
lernst du jede Stufe isoliert:

- [ ] **PR 1:** `setup-go` mit Cache (`cache: true`), damit Builds schnell sind.
- [ ] **PR 2:** `golangci-lint` dazu (staticcheck, gosec, errcheck). Legt
  `.golangci.yml` an.
- [ ] **PR 3:** `govulncheck` (bekannte CVEs in Dependencies) + `gitleaks`
  (keine Secrets im Code).
- [ ] **PR 4:** `gofmt -l`-Check — bricht, wenn ungeformatter Code kommt.
- [ ] **PR 5:** `docker build` im PR (ohne Push) + **Trivy**-Scan des Images.
- [ ] **PR 6:** Jobs parallelisieren + `concurrency`-Gruppe (alter Run wird
  abgebrochen, wenn du neu pushst).
- [ ] **PR 7:** `permissions: contents: read` als Default setzen, Third-Party-
  Actions auf Commit-SHA pinnen (Security-Hardening, Guide §11).

> Faustregel: billige Checks zuerst (fmt, vet), teure danach (Tests, Image-Scan).
> PR-Feedback soll unter 5 Minuten kommen.

### A3 — Makefile als einzige Tür

Lokal und CI sollen **dieselben** Befehle laufen. Erweitere das `Makefile`:

- [ ] `make ci` = fmt-check + vet + lint + test + govulncheck, ein Ziel.
- [ ] `make help` als Übersicht.
- [ ] Die CI ruft dann nur noch `make ci` auf — kein Drift zwischen lokal und CI.

Aktuell existieren `run`, `build`, `test`, `css`, `watch`. Env-Vars:
`PORT`, `DATABASE_PATH`, `APP_ENV`, `SITE_NAME`, `SITE_DOMAIN`, `MAIL_FROM`,
`SMTP_HOST/PORT/USER/PASS` (siehe `internal/config/config.go`). Lege eine
`.env.example` an (committed) mit allen Keys und Dummy-Werten.

### A4 — Image bauen und in die Registry pushen

Das `Dockerfile` ist schon multi-stage und `CGO_ENABLED=0` (winziges Image).
Jetzt automatisieren:

- [ ] **PR 8:** Workflow `main` → baut multi-arch Image (`linux/amd64,arm64`
  via buildx) → pusht nach **GHCR** (`ghcr.io/<user>/studigame`), getaggt
  `sha-<short>` + `main`. Nutzt `docker/metadata-action` +
  `docker/build-push-action`, Auth über `GITHUB_TOKEN`.
- [ ] Kernregel verinnerlichen: **einmal bauen, denselben Digest überall
  deployen.** Nie `latest` in Produktion.

### A5 — Docs-as-Code anfangen

- [ ] `docs/adr/0001-trunk-based.md` schreiben (Vorlage im Guide §15). Danach
  ADRs 2–7 nach und nach (htmx, SQLite, Sessions, Monolith, Lazy-Resolve,
  Cents+Ledger).
- [ ] `docs/architecture.md` mit einem Mermaid-C4-Diagramm (Context + Container).
- [ ] `markdownlint-cli2` + `lychee` (toter-Link-Check) in die CI.

### A6 — 🙋 VPS mieten + härten + privat deployen

Dein zentrales Lern-Objekt. Der Server läuft, die App ist erreichbar — aber nur
für dich, nicht öffentlich beworben.

- [ ] **Hetzner CAX11** (ARM, ~4 €/Monat) bestellen. SSH-Key **beim Anlegen**
  hinterlegen, Passwort-Login nie aktivieren.
- [ ] Härten: `ufw` nur 22/80/443, root-SSH aus, unprivilegierter User,
  `fail2ban`, `unattended-upgrades`. Das ist der Linux/Networking-Working-Depth
  aus dem Guide — am echten System geübt, nicht in der Theorie.
- [ ] Docker + Compose-Plugin installieren.
- [ ] **Privat deployen:** `docker-compose.yml` + `deploy/Caddyfile` existieren
  schon (App + Caddy, Auto-HTTPS). `.env` mit `APP_ENV=production` und einer
  **Test-Subdomain** (z. B. `test.deinedomain.de`) befüllen, `docker compose
  up -d`. Caddy zieht automatisch ein Let's-Encrypt-Zertifikat.
- [ ] **Go-Port nie veröffentlichen:** Die App vertraut `X-Forwarded-For` (Rate-Limit).
  `docker-compose.yml` nutzt dafür `expose`, nicht `ports` — so bleibt es. Wer
  Port 8080 direkt erreicht, könnte das Limit mit gefälschten Headern umgehen.
  Außerhalb von Docker: `PORT` nur hinter Caddy, Firewall zu. Caddy überschreibt
  den Header von außen (`trusted_proxies` nicht setzen).
- [ ] **Smoke-Test:** `/healthz` liefert 200, Seiten laden über HTTPS.
  (`/readyz` mit DB-Ping fehlt noch — guter kleiner Issue für den Agent.)
- [ ] Solange keine Werbung läuft und das Formular nicht öffentlich verlinkt ist,
  sammelst du hier noch keine fremden Daten — die Rechtspflicht greift erst,
  wenn du die Seite öffentlich bewirbst (Phase B). Für dich zum Testen reicht
  der private Betrieb.

**Ende Phase A:** Jeder Push läuft durch eine echte Pipeline, ein signiertes
Image landet in GHCR, `main` ist geschützt, **und die App läuft per HTTPS auf
deinem Server** — privat, zum Üben. Kosten ~4 €/Monat + Domain.

---

## Phase B — „Kommt bald" öffentlich schalten (🟡 wenn Recht + Mail stehen)

Server und Deploy stehen schon aus Phase A. Hier kommt nur noch das dazu, was
die Seite von „läuft privat" zu „darf öffentlich E-Mails sammeln" bringt: das
Rechtliche und der echte Mailversand. Danach ist der Schritt zum Live-Schalten
nur noch ein DNS-Eintrag und ein Link.

### B1 — 🙋 Rechtliches (blockiert alles Öffentliche)

- [ ] **Domain registrieren** (z. B. bei INWX/Namecheap, ~10 €/Jahr).
- [ ] **Impressum** schreiben (Name, Anschrift, Kontakt — §5 DDG).
- [ ] **Datenschutzerklärung** (Art. 13 DSGVO: was wird gespeichert, warum,
  wie lange, Rechte). Für eine Warteliste mit einem E-Mail-Feld überschaubar.
- [ ] **Kein Cookie-Banner nötig**, solange kein Tracking/Analytics läuft und
  Fonts selbst gehostet sind (§25 TDDDG). Nicht aus Reflex einbauen.
- [ ] Einmaliger Rechts-Check vor Go-Live lohnt sich (Bußgeldrahmen bis 50 k€).

### B2 — 🙋 Mailversand

- [ ] Anbieter wählen: **Brevo** (kostenloses Kontingent), Postmark oder SES.
- [ ] **SPF, DKIM, DMARC** im DNS setzen — ohne die landet die Bestätigungsmail
  im Spam und die Warteliste bleibt leer.
- [ ] `SMTP_HOST/PORT/USER/PASS` in `.env` auf dem Server befüllen. Leerer
  `SMTP_HOST` = Konsolen-Mailer (nur für Dev).
- [ ] **Übung (Guide §11):** rotiere die SMTP-Credentials einmal komplett durch
  und schreibe ein Runbook `docs/runbooks/rotate-smtp.md` dabei.

### B3 — Öffentlich schalten

- [ ] `SITE_DOMAIN` in der `.env` von der Test-Subdomain auf die echte Domain
  umstellen, A/AAAA-Record setzen, `docker compose up -d`.
- [ ] Impressum- und Datenschutz-Links auf der Seite sichtbar (Footer).
- [ ] Erst **jetzt** die Seite bewerben / verlinken — ab hier sammelst du echte
  fremde Daten, und ab hier muss das Recht stehen.

### B4 — CD: Deployment automatisieren

- [ ] **PR:** Workflow, der bei einem Release-Tag per SSH auf den Server geht und
  `docker compose pull && up -d` ausführt. Credentials als GitHub-Environment-
  Secret, Prod-Environment mit **required reviewer** (du bestätigst den Deploy).
- [ ] **Backup:** `scripts/`-Backup existiert schon — aber ein `cp` auf eine
  WAL-SQLite ist unsicher. Umstellen auf `VACUUM INTO` oder `.backup`. Nightly
  per Cron + **einmal echten Restore testen** (Runbook!).

**Ende Phase B:** „Kommt bald"-Seite ist öffentlich, sammelt bestätigte E-Mails,
ein Tag-Push deployt automatisch. Server und HTTPS liefen schon seit Phase A.

---

## Phase C — Kubernetes-Lernpfad (🔵 Kür, nach dem MVP-Launch)

Hier lernst du den Teil, den Jobanzeigen wollen. **Wichtig:** Das Compose-Setup
aus Phase A bleibt in Produktion, bis k3s wirklich sitzt. Kaputt-deployen übst du
auf Staging, nicht auf der Live-Warteliste.

- [ ] **Lokal zuerst:** `kind` oder `k3d`-Cluster, App als Deployment rein.
- [ ] **Kustomize** `base/` + `overlays/{staging,prod}`. Gleiches Image, nur
  Config unterscheidet sich (12-Factor).
- [ ] **Die eine Regel für SQLite:** `replicas: 1`, `strategy: Recreate`, RWO-PVC.
  SQLite hat genau einen Writer — ein Rolling-Update würde einen zweiten Pod auf
  dieselbe DB starten. Das ist die wichtigste Design-Entscheidung hier.
- [ ] **k3s** single-node auf einem (zweiten) VPS.
- [ ] `cert-manager` + Let's-Encrypt (ersetzt Caddy), Traefik-Ingress (bei k3s dabei).
- [ ] **Probes:** liveness `/healthz`, readiness `/readyz`.
- [ ] **securityContext:** `runAsNonRoot`, `readOnlyRootFilesystem`, Capabilities
  droppen.
- [ ] **Secrets:** SOPS+age oder Sealed Secrets (nie Klartext-Secrets in Git).
- [ ] **Litestream-Sidecar:** kontinuierliches DB-Backup in Object-Storage.
- [ ] **GitOps mit Argo CD:** der Cluster zieht sich den Zustand aus Git. CI hält
  dann keine Cluster-Credentials mehr — das ist der „pull statt push"-Ansatz.
- [ ] **NetworkPolicy:** default-deny, nur DNS/SMTP/Object-Storage erlauben.

> Reihenfolge-Tipp (Guide §14, M6): cert-manager → Secrets → Litestream → Argo CD
> → Staging/Prod-Approval. Nicht alles auf einmal.

---

## Phase D — Security, Monitoring, Betrieb (🔵 Dauerlauf)

Läuft begleitend ab Phase B, vertieft sich in C. Nichts davon ist „einmal fertig".

**DevSecOps (Guide §11):**

- [ ] CodeQL + gosec im PR.
- [ ] SBOM (Syft/Trivy) + **cosign** keyless-Signatur beim Image-Build.
- [ ] 1-Seiten-Threat-Model (`docs/threat-model.md`) + `SECURITY.md`.
- [ ] Kyverno im Cluster: nur signierte Images, kein `latest`, Limits Pflicht.
- [ ] Dependabot für `gomod`, `docker`, `github-actions`.

**Observability (Guide §12):**

- [ ] slog-JSON läuft schon → Grafana Alloy → Loki → Grafana.
- [ ] `/metrics` (Prometheus): RED-Metriken pro Route + Business-Counter
  (`actions_started_total{kind}`, `charity_overflow_cents_total`, …).
  **Keine** High-Cardinality-Labels (keine User-IDs).
- [ ] kube-prometheus-stack (Prometheus, Alertmanager, Grafana).
- [ ] Ein Dashboard: Traffic, Fehler, Latenz, Aktionen pro Art, Geld rein vs.
  Charity-Überlauf, Pod-Restarts, PVC-Füllstand.
- [ ] **SLOs** definieren (z. B. 99,5 % Dashboard-Requests non-5xx, p95 < 300 ms,
  letztes Backup < 26 h) + Alerts, die auf dein Handy gehen (ntfy/Discord).

**Betrieb (Guide §13):**

- [ ] **Game-Days auf Staging:** Pod mitten in einer Schicht killen, PVC löschen
  und restoren (echte RTO messen), kaputtes Image deployen, TLS brechen, Disk
  volllaufen lassen, **Cluster komplett aus Git + Backup neu bauen**.
- [ ] Nach jedem Vorfall: blameless **Postmortem** mit Action-Items als Issues.
- [ ] **DORA-Metriken** monatlich (`docs/metrics.md`): Deploy-Frequenz, Lead-Time,
  Change-Failure-Rate, Time-to-Restore.

---

## Was der Agent macht vs. was du machst

| Agent (Spielcode) | Du (DevOps) |
| --- | --- |
| Restliche MVP-Features, Balancing-Hooks | Git-Prozess, Branch-Schutz |
| `/readyz` mit DB-Ping (kleiner Issue für ihn) | CI-Pipeline Stufe für Stufe |
| Version-Info ins Binary (`-ldflags -X`) | Image-Build + GHCR + Signatur |
| Spielregeln, Tests der Formeln | Server, DNS, TLS, Deploy, Backup |
| — | Monitoring, Alerts, Security-Scans |
| — | 🙋 Domain, Impressum, Mailversand, Rechts-Check |

Schnittstelle zwischen euch: **Issues**. Wenn dir beim Deploy etwas fehlt
(`/readyz`, Version-Endpoint, ein Env-Var), schreib ein Issue statt selbst im
Spielcode zu wühlen — das ist genau die saubere Trennung, die du üben willst.

---

## Heute anfangen — die ersten drei Schritte

1. **Branch-Schutz auf `main`** aktivieren (A1). 5 Minuten, sofort Wirkung.
2. **CI-PR 1** (setup-go-Cache) öffnen — dein erster bewusster CI-Commit (A2).
3. **VPS bestellen** (A6) — dein Lern-Objekt. Härten und die App per
   Test-Subdomain privat deployen, parallel zu CI/Image.

Danach arbeitest du Phase A von oben nach unten ab. Öffentlich bewerben erst in
Phase B (wenn Recht + Mail stehen), Kubernetes erst in Phase C.
