# Studigame: Rebuild & Learning Guide

A guide for rebuilding Studigame from scratch, by hand, to learn **Go** and
then the **SDLC/DevOps** side (Git, docs, testing, CI/CD, DevSecOps,
observability, Kubernetes). It summarises the first, AI-built version
(`studigame-plan.md`, `ROADMAP.md`, `README.md`, `CHANGELOG.md`, `CLAUDE.md`,
code). Architecture is given as hints only, so you design it yourself.

**Depth legend:** **Deep** = design and explain trade-offs · **Working** = set up and
use properly · **Awareness** = know what it is and when you'd need it.

**Core DevOps skill set (most job ads expect these):**

| Skill | Tools used here | Depth |
| --- | --- | --- |
| CI/CD pipelines | GitHub Actions (concepts transfer to GitLab CI/Jenkins) | Deep |
| Containers | Docker, multi-stage builds, registries (GHCR) | Deep |
| Kubernetes | k3s, Kustomize, Helm (consume), probes, RBAC, NetworkPolicy | Deep |
| GitOps | Argo CD (or Flux) | Working → Deep |
| Infrastructure as Code | OpenTofu/Terraform, cloud-init or Ansible | Working → Deep |
| Monitoring & alerting | Prometheus, Grafana, Alertmanager, SLOs | Working → Deep |
| Logging & tracing | Loki + Alloy, OpenTelemetry | Working |
| DevSecOps | Trivy, govulncheck, cosign, SBOM, Kyverno, secrets (SOPS) | Working → Deep |
| Testing & QA | Go testing, fuzzing, k6, ZAP | Deep |
| Git & docs | trunk-based, SemVer, ADRs, runbooks, changelog | Deep |
| Linux/networking basics | SSH, systemd, DNS, TLS, firewalls (you'll hit these on the VPS) | Working |
| Planning & process | Issues, board, stories, DoD | Working, light |
| Artifact management | GHCR, Go proxy. Artifactory/Nexus/Harbor | Awareness |
| UML / diagrams | C4 + deployment diagram in Mermaid (you know UML already) | Awareness |

Skip or keep at awareness: Jenkins, service meshes, multi-cluster,
microservices (a monolith is right here), deep Artifactory admin, full UML.

---

## 1. The game

### 1.1 Pitch & tone

A text-based browser game inspired by *Pennergame* (2007), but its own thing:
**student life instead of homelessness.** You're a broke student in the
fictional uni town **Studistadt**. You collect bottles after WG parties, work
Werkstudent jobs, ghostwrite term papers, train skills and upgrade your wallet
from *Portemonnaie* to *Bausparvertrag*.

- **Tone:**
  - Dry, darkly humorous satire of student life (being broke, exam stress, WG
    chaos), with German insider humour (BAföG, Mensa, ECTS bureaucracy).
  - It mocks no real vulnerable group, and almost everyone can relate.
  - No exclamation marks, no emoji in the UI.
  - Test for every text: *would I show this to a Studierendenwerk?*
- **Content:** fictional town, universities and people. Subjects differ through
  numbers and affectionate clichés, never real contempt for a field. No dated
  2000s shock humour.
- **Name:** **Studigame**. It sounds close to "Pennergame" on purpose (slang
  word + "game").
- **Legal:**
  - The landing page says honestly that it's inspired by the 2000s cult browser
    games (incl. Pennergame) and has no connection to Farbflut Entertainment.
  - Never use their name or assets. The company is still active.
- **UX:**
  - Mobile-first, server-rendered, minimal text-based UI.
  - "One tap = start an action": a check-in between two lectures takes about
    2 minutes.
- **Money:** donations + cosmetic "Alumni/Förder" supporter status (ad-free,
  badge). **Never pay-to-win.**

### 1.2 The original, as a reference for later features

- Started with 10 € and the goal of becoming a "Schlossbesitzer". Several parallel
  city editions. Premium was cosmetic only. The backend was Python + MySQL.
- **Collecting:**
  - You chose a duration between 10 min and 12 h.
  - Shorter shifts were more efficient per hour, longer ones paid more in total.
  - Yield depended on the district, the dexterity skill and a daily bottle price.
- **Container:** a cup → … → container ladder. Money beyond the cap was **lost**.
- **Skills:**
  - Cost was `base × next level` (**linear**, which caused the "overpowered too
    fast" snowball). Cancelling refunded 80 %, and two skills could be queued.
  - The skills: ATT, DEF, dexterity (more bottles, from level 20 you see incoming
    attacks), speaking (prerequisite for most), education (unlocks begging
    spots), music (instruments pay passively every 6 h), social contacts (pets,
    homes), concentration (faster training), pickpocketing (slightly more income).
- **PvP:**
  - The same target could be attacked only every 36 h.
  - Fight duration depended on how far apart the districts were. There was a
    fight log.
  - **Pet fights** with bets and stances were the only way to send money to
    another player.
- **Gangs:**
  - Name + password to join.
  - Gang property (armory, house, training, account) with 5 levels each and
    % bonuses for all members.
  - Gang chat, alliances and wars, a separate gang highscore.
- **Social:** friends and block list, messages, forum signatures (every click
  on one = an in-game donation).
- **Growth engine:** a personal referral link (`/change_please/{id}`, roughly
  "got some change?"). A click gave the owner an in-game donation, so viral
  growth was built into the begging fantasy.

### 1.3 Curation: what to keep, change or drop

| Element | Decision | Why |
| --- | --- | --- |
| Core loop (collect → sell → invest → risk → climb) | Keep | Timeless |
| Container with cap | Keep, but overflow → **charity pot** | Same tension, feels good instead of punishing |
| Alcohol (Promille) system | Keep as **coffee** | Fits student life |
| Cleanliness stat | Park | A second stat to maintain doubles UI and balancing work |
| Referral link | Keep and extend | Growth built into the fantasy |
| Cosmetic-only membership | Keep | Fits the monetisation stance |
| Real cities/districts | Fictional town + **subjects as classes** | Creative freedom, distance from the original |
| Shock humour names | Drop | Dated, was borderline even then |
| "End your life" = delete account | "Account löschen" | Reads like self-harm |
| 36 h cooldown per target | Keep, tune | Anti-griefing |
| No direct money transfer | Keep | Prevents farming and real-money trading |
| Linear skill costs | **Superlinear** `base × n^1.8` + horizontal goals | Anti-snowball |

### 1.4 Core loop

1. **Collect / work.** Pick a shift. The result is **rolled when it ends**, not at start.
2. **Keep it.** Your container has a cap. What doesn't fit goes to the town's
   **charity pot**. (Turning that into a real donation is unsolved: who pays,
   at what rate, accounting, tax.)
3. **Reinvest.** Container upgrades first (the first money sink), then skills,
   equipment, housing.
4. **Risk.** Ghostwriting, selling Spickzettel, grade competition.
5. **Climb.** Semesters, ECTS, highscore.

**MVP = steps 1–3. Then stop and play.** If collecting, the container and upgrades
are boring, more features won't fix it.

### 1.5 Subjects (class system)

The subject replaces the original's districts completely. Campus places
(Mensa, Bibliothek, Kneipenviertel) are only flavour text labels: no
multiplier, no unlock. Start with 4 subjects, since each extra one is balancing
work.

| Subject | Bonus | Malus / flavour |
| --- | --- | --- |
| Informatik | Konzentration cheaper ("Debug-Modus"), good Werkstudent jobs | – |
| BWL | Bigger starting container, cheaper networking | Pricier "look" equipment needed for status |
| Jura | Bonus in grade competition / PvP (cutthroat) | More expensive, longer trainings |
| Geisteswissenschaften | Cheap, versatile skills, social/music bonus | Lower income ceiling (self-ironic running gag) |

- Until skills exist, the subject is pure identity.
- **Fachwechsel** (later) costs time and money and resets some progress. It
  mirrors the real pain and gives a natural "big decision".
- **Zweitstudium** (later) is prestige: restart from zero with a small permanent
  bonus, maybe in another subject.

### 1.6 Mechanic mapping (original → Studigame)

| Original | Studigame |
| --- | --- |
| Deposit bottles | Leergut after the WG party |
| Selling | Automatic on return (no separate step) |
| Container | Portemonnaie → Jutebeutel → Rucksack → IKEA-FRAKTA-Tasche → **Bausparvertrag** (endgame joke) |
| Education level | ECTS points / semester count |
| ATT | Durchsetzungsvermögen (grade competition, group work) |
| DEF | Stressresistenz (less mood/money loss) |
| Dexterity | Lerneffizienz (more yield per action) |
| Speaking | Präsentationstechnik (prerequisite for many skills) |
| Concentration | Konzentration (faster study phases) |
| Music | Street / WG-hallway concerts (passive income) |
| Pickpocketing | Rabattcode-Jäger (max out student discounts) |
| Social contacts | Networking (unlocks WGs/flats, later gangs) |
| Crimes | Ghostwrite a Hausarbeit, sell Spickzettel, top tier: steal exam questions |
| Passive begging | Ask parents (modest), later BAföG / Stipendium |
| PvP | Notenkonkurrenz in curved exams, sabotaging group work |
| Gangs | WGs: WG rivalry, faculty feuds (BWL vs. Informatik vs. Geisteswiss.) |
| Gang property | WG equipment: WLAN speed, coffee machine, bookshelf, WG-Kasse |
| Home ladder | WG-Zimmer → Wohnheim-Einzelzimmer → 1-Zimmer → Altbau → Eigentumswohnung |
| Honorary membership | Alumni status / Fördermitglied, no gameplay advantage |
| Alcohol | Coffee (temporary concentration boost) |
| Referral link | `/kleingeld/{code}` with click de-duplication |

### 1.7 What v1 built (numbers are starting points, tune by playing)

| Knob | Value |
| --- | --- |
| Start | 10 € (booked via the ledger), mood 100 (0–100), one character per edition, name unique per edition |
| Registration | Also asks for a **Heimatstadt** (real home town, display only. Idea: city leaderboards later) |
| Mood | Every shift costs 5 per hour (min 2). Output × (0.6 + 0.4 × mood/100) |
| Shifts | 15 min, 1 h, 4 h, 8 h. One open action per character |
| Collect | 40 bottles/h × 25 ct × luck 0.8–1.2 × (1 + 0.15 × container level) × skill |
| Job (Werkstudent) | 8 €/h × skill, no luck |
| Crime (Nebenverdienst) | 15 €/h. 30 % caught → 25 % payout and double mood cost. Skill-neutral for now |
| Klausur | 0.25 ECTS/h × skill. **Open: nothing consumes ECTS yet** |
| Container capacity | 20 / 80 / 200 / 1000 / 5000 € (level 0 must exceed the 10 € start) |
| Container upgrade | `2000 ct × (level+1)^1.8`, rounded to euros → 20 / 70 / 144 / 243 €. Instant buy, one tap, ledger reason `container_upgrade`, max level 4 hides the button. **Invariant: capacity ≥ next upgrade cost** (v1 had a softlock: Jutebeutel 50 € < 70 €) |
| Skills | Lerneffizienz (+5 %/lvl collect), Auftreten (+5 %/lvl job), Konzentration (−5 %/lvl training time, floor 50 %, +5 %/lvl ECTS). Cost `base × (lvl+1)^1.8` (base 10 €, Konz. 12 €). Max level 20/20/10. 2 h training. **One slot** (original had two). 80 % refund on cancel. Bonus **frozen at action start** |
| Subject modifiers | Geisteswiss. cost × 0.8 · Jura cost × 1.2 and time × 1.2 · Informatik Konz. cost × 0.9 |
| BAföG | Apply once, then 30 € every 6 h, no mood cost. Catch-up capped at 4 periods. Pointer stays on the original cadence |
| Coffee (Campus-Kiosk) | 2 € → training time halved for 2 h. Purchases extend the window |
| Shop (draft) | Equipment (Studentenausweis, Notebook, Fahrrad), instruments (Gitarre, Keyboard → passive income), housing (WG-Zimmer, Wohnheim, 1-Zimmer → container bonus). **Least thought-through, redesign it** |
| Charity pot | One per edition. Every overflow is added |

**Deferred skills** (Bildungsstufe, Sozialkontakte, Musik, Rabattcode-Jäger,
ATT, DEF) are postponed, not cancelled. The housing ladder and most subject
bonuses depend on them.

**Open design questions:**
- Passive BAföG (5 €/h) is almost as good as the active job (8 €/h). Widen the gap or
  keep it on purpose?
- **Endgame hole:** all container upgrades cost about 477 € in total. After that,
  nothing absorbs money until the shop exists. What's the long-term sink?
  (Housing ladder, prestige?)
- What does ECTS buy? (Bildungsstufe, Stipendium unlock, semester progress,
  highscore.)
- Do the deferred skills come back, or are they gone for good?
- Charity pot → real donation (needs a partner, accounting, tax).
- Stipendium: rarer and better than BAföG, unlocked from an education threshold.

### 1.8 Later features (specs from the plan and roadmap)

- **Notenkonkurrenz (PvP):** ATT/DEF, 36 h cooldown per target, **newbie
  protection**, fight log.
- **WGs (gangs):** create/join, shared Kasse, upgrades, chat, two roles
  (Chef/Mitglied). Later: alliances and wars, finer roles.
- **Social:** messages, friends, block list, **report button**. Moderation rules
  must exist before launch.
- **Highscore** per edition (players + WGs).
- **Referral** `/kleingeld/{code}` with click de-duplication. Build it early,
  since it's the best growth lever.
- **Seasonal content** from the academic calendar: *Klausurenphase*
  (high risk/reward), *Semesterferien* (calmer, job focus).
- **Community:**
  - Honest "inspired by" note and familiar terms for old players.
  - Fast mobile onboarding for new ones.
  - Optional **nostalgia quiz** (badge only).
  - **"Urgestein" badge** for beta testers.

### 1.9 Phases

| Phase | Content | v1 status |
| --- | --- | --- |
| 0 Landing | One page, e-mail waitlist (htmx, double opt-in), Impressum + privacy, self-hosted fonts, name | done (domain open) |
| 1 Foundation | Schema, register/login/logout/reset, double opt-in, subject choice, sessions, CSRF, rate limits, layout, Docker + Caddy deploy | done |
| 2 Design system | Tailwind tokens, components (button, card, form, status bar, bottom nav, flash, progress) | done |
| 3 Core loop | Collect, container + overflow, upgrade, job, skills, BAföG, coffee, crime, Klausur, shop | built, **balancing never played** |
| 4 Social & PvP | See 1.8 | open |
| 5 Mobile/PWA | Manifest + icons (service worker optional), performance and a11y pass. **No web push** (calm tone, avoids addiction mechanics) | open |
| 6 Monetisation | Cosmetic shop + supporter status, Stripe/Paddle with **idempotent webhooks**, donate button, real charity link | open |
| 7 Launch | Closed beta via the waitlist, public changelog, Discord from day one (also the feedback channel, no own forum), transparent Reddit posts, TikTok/Reels nostalgia, campus posters with QR codes | open |

---

## 2. Design system

- **v1 look:** modern-minimal. Indigo primary + lime accent on slate neutrals,
  plus success/danger/warning.
  - The plan's original palette was dropped because it was too close to
    Pennergame: Asphalt `#1C1B1A`, Pfandglas-Grün `#7A9B57`,
    Verkehrsorange `#D9711F`, Münzgold `#C9A24B`, Pappe `#E8E1D3`,
    Backstein `#A83E32`, Textmarker `#E8D630`.
  - Pick your own.
- **Tokens:**
  - CSS variables in Tailwind v4 `@theme`.
  - **Light/dark** via `[data-theme="dark"]`. A toggle in the status bar stores
    the choice in `localStorage` and falls back to the system preference.
- **Typography:**
  - Headlines: a bold condensed grotesque (Archivo Black / Barlow Condensed),
    "cardboard sign" feel.
  - Body: Inter or Source Sans.
  - Money and stats: monospace (JetBrains Mono).
  - System fonts are fine, but self-host anything else.
- **Layout:**
  - Mobile-first, one column.
  - **Sticky status bar** at the top: money, mood, running activity badge.
  - **Bottom nav** with icons and an active state (Übersicht / Aktionen / Lernen
    / Shop, later Stadt / WG).
- **Signature element:** the money container as a **progress bar in the
  header**. It fills up, shows the level and next tier, and changes shape on upgrade.
- **Components:** hand-written Tailwind partials (button, card, badge, form
  field, flash, stat, progress, skeleton). **No DaisyUI.**
- **Icons:** Lucide SVGs inline via a Go template func `{{icon "wallet"}}`.
  No icon font, no npm.
- **Countdown:** a small Alpine component counts down and reloads when the shift
  ends. `<meta refresh>` stays as the no-JS fallback.

---

## 3. Tech stack

| Area | Choice | Alternatives / notes |
| --- | --- | --- |
| Language/HTTP | Go (latest), `net/http` with 1.22+ routing (`"POST /actions/{kind}"`, `r.PathValue`) | No framework needed |
| Templates | `html/template` + `embed` | `templ` if you want type safety (codegen step) |
| DB | SQLite (WAL) via `modernc.org/sqlite` (pure Go, `CGO_ENABLED=0`, tiny static images) | `mattn/go-sqlite3` needs CGO. Postgres only if measured or for k8s scaling practice |
| SQL / migrations | `database/sql`, hand-written SQL, own runner (numbered `.sql` embedded + `schema_migrations`) | `sqlc`, `goose` |
| Auth | Server-side sessions: random 32-byte token in an HttpOnly cookie, **SHA-256 hash** in the DB, bcrypt (`golang.org/x/crypto`) | No JWT: one web app, instant revocation. JWT only for a future stateless API |
| CSRF | Go 1.25 `http.CrossOriginProtection` or double-submit cookie (v1) | |
| Logging | `log/slog` JSON to stdout | |
| Config | Env vars only | Maps to k8s ConfigMap/Secret |
| Mail | `net/smtp` behind a `Mailer` interface + console fallback for dev | Provider (Brevo, Postmark, SES…) with SPF/DKIM/DMARC |
| Time / luck | `func() time.Time` and `func() float64` injected | `testing/synctest` |
| Frontend | **htmx** + **Alpine.js CSP build**, both vendored, **Tailwind v4 standalone CLI** (one binary, downloaded per platform, git-ignored), Lucide SVGs | No Node, no JS build |
| Tests | stdlib `testing`, `httptest`, `cookiejar` | No testify, no mocks |
| Hosting | v1: 1 VPS (Hetzner) + Caddy (auto-HTTPS) via Compose | Learning path: k3s (§9) |
| Payments (phase 6) | Stripe or **Paddle** (Merchant of Record, handles EU VAT) | |

Runtime dependencies: **only** `modernc.org/sqlite` and `golang.org/x/crypto`.

**Env vars (v1):** `PORT` (dev 8410), `DATABASE_PATH`,
`APP_ENV` (`production` → secure cookies), `SITE_NAME`, `SITE_DOMAIN`,
`MAIL_FROM`, `SMTP_HOST/PORT/USER/PASS` (empty → console mailer).
Commit `.env.example`. `make run` must load `.env` (v1 forgot this at first).

**Dev commands (v1):** `make run`, `make test`, `make css` (downloads Tailwind
on first run), `make watch`, `docker compose up -d --build`.

---

## 4. Architecture hints

**Principles (from `CLAUDE.md`):**
- It's a solo MVP: too much abstraction is the most expensive mistake. When in
  doubt, write the shorter solution.
- No interface with only one implementation (Mailer has two, so it's fine).
- No dependency for what the stdlib can do. No config option for a value that
  never changes.
- No code "for later". **Deleting is an improvement.** The one sanctioned
  exception: the `editions` table, so more towns need no schema change.

**Non-negotiable rules:**
- The `game` package is pure: stdlib only, **no DB, no HTTP, no `time.Now()`, no
  `rand`**. Time and luck are passed in.
- Money = `int` cents, never float. Every movement goes through **one booking
  function** that writes a **ledger** row.
- One player action = one DB transaction. Handlers stay thin: parse → call
  action → render.
- Roll outcomes at **resolve** time, otherwise the result sits in the DB in
  advance. Never take results, prices or amounts from the form.
- Every query about other players filters by `edition_id`.

**Possible layout:**
- `cmd/server`: wiring only
- `internal/config`
- `internal/db`: open, migrations, sweeper
- `internal/game`: pure rules + tests
- `internal/play`: use-cases, TX, SQL
- `internal/auth`: sessions, passwords, tokens, mailer
- `internal/httpx`: middleware
- `internal/handlers`
- `internal/web`: templates, components, render, funcs
- `static/`
- `deploy/`

v1 put the SQL "store" code in `handlers`, which blurred the layers.

**Key patterns:**
- **Timed actions, resolved lazily** (no background ticker):
  - Start: insert an `actions` row with a **frozen** `payload_json` (mood,
    container level, skill factor).
  - Enforce one open action per character with
    `UNIQUE INDEX … ON actions(character_id) WHERE resolved_at IS NULL`. A
    unique violation means "busy".
  - Page load: resolve in one TX. Roll now, compute via `game`, book, then a
    conditional `UPDATE … AND resolved_at IS NULL` with a `RowsAffected` check.
  - Dispatch on `kind` with a plain `switch`, no resolver framework. Use a
    **typed outcome struct**.
  - Show the result once as a flash message.
- **Booking path:** `book(tx, char, delta, reason, refID)` applies the cap
  (`SplitAtCapacity`), updates the balance, writes the ledger
  (`delta_cents`, `balance_after`, `reason`, `ref_id`) and adds overflow to
  `charity_pool`. The ledger holds money only. ECTS gets its own column.
- **Debits:** `UPDATE … SET money = money - ? WHERE id = ? AND money >= ?`.
  0 rows means "not enough money".
- **Passive income** (BAföG, instruments):
  - Store a `next_payout_at` pointer.
  - On load, pay `min(periods, cap) × amount`.
  - Advance the pointer by whole periods (not to `now`).
  - Condition the update on the old pointer value.
- **Catalogs** (skills, tiers, items) live in Go code: versioned, typed, testable.
  The DB stores only what a character owns or has reached (by slug).
  Editions and subjects are seeded via migration.
- **HTTP:**
  - Middleware is `func(http.Handler) http.Handler`. Order: recovery → logging →
    security headers → rate limit → CSRF → auth.
  - Pass the user via `context` or a `RequireAuth(func(u, w, r))` adapter.
  - Use POST → redirect → GET with a flash cookie. Requests with an `HX-Request`
    header get a partial instead.
  - Typed view structs with an embedded `Base` (site, user, CSRF, active path,
    status bar) instead of `map[string]any`.
  - Map domain errors to German flash messages. Unknown errors → log + 500.
- **Templates:**
  - `{{define "content"}}` blocks share one namespace (the last one wins), so
    parse **one set per page** (layout + components + page) at startup with
    `template.Must`.
  - Add a test that parses all templates (v1 had parse panics).
- **SQLite:**
  - Pragmas: `WAL`, `foreign_keys`, `busy_timeout`, `_txlock=immediate`.
  - One writer (v1: `SetMaxOpenConns(1)`. Better: 1 write conn + a read pool).
  - Store times one way (Unix ints or UTC RFC3339).
  - Never edit an applied migration.
- **Lifecycle:**
  - Use `signal.NotifyContext` and pass `ctx` to every goroutine (sweeper every
    minute).
  - `server.Shutdown` on SIGTERM (k8s sends it).
  - Read/Write/Idle timeouts, `MaxBytesReader`.
  - `/healthz` (process up) + `/readyz` (DB ping).
  - Embed `time/tzdata` (minimal images have no tz database).

**Routes (v1):**

```
GET  /healthz                    GET  /  (landing)        POST /waitlist
GET|POST /register   GET /verify?token=   GET|POST /login   POST /logout
GET|POST /reset-password         GET|POST /reset-password/confirm
GET  /dashboard  (resolves due actions + BAföG)   GET|POST /character/new
GET  /actions    POST /actions/{collect|job|crime}   POST /actions/upgrade-container
GET  /klausur    POST /klausur/start
GET  /skills     POST /skills/train   POST /skills/cancel
POST /bafog/toggle   POST /kiosk/coffee   GET /shop   POST /shop/purchase
```

Rate limits: auth routes 10/min/IP, action routes 30/min/IP. Session TTL 30 days,
reset token TTL 60 min.

**Data model sketch:**

```
users(id, name, email UNIQUE, password_hash, verified, city, created_at)
sessions / password_resets / verification_tokens(token_hash PK, user_id, expires_at)
waitlist(id, email UNIQUE, verified)
editions(id, slug, name, active)            subjects(id, edition_id, name)
characters(id, user_id, edition_id, subject_id, name, money_cents, container_level,
           mood, ects, bafog_next_at, caffeine_until)  UNIQUE(user,edition) UNIQUE(edition,name)
ledger(id, character_id, delta_cents, balance_after, reason, ref_id, created_at)
charity_pool(edition_id PK, cents)
actions(id, character_id, kind, started_at, ends_at, resolved_at, payload_json, result_json)
skills(character_id, skill, level)          skill_trainings(character_id UNIQUE, skill, level, ends_at, cost_cents)
character_items(character_id, item_slug, quantity)
```

Use random public IDs for anything shown in URLs (prevents enumeration).
`ON DELETE CASCADE` from `users` makes "Account löschen" (GDPR) simple.

**Lessons from v1:**

| Kept (worked well) | Do differently |
| --- | --- |
| Pure `game` + injected time/luck + table tests | ECTS was parsed out of a display string and booked into the money ledger → typed outcome |
| Frozen payload, roll at resolve, partial unique index | Upgrade checked money on a stale struct outside the TX → conditional debit |
| One booking path for all kinds | BAföG double payout under concurrency → conditional updates from day one |
| Hashed tokens, bcrypt, reset kills all sessions | Shared `signal.Notify` channel could hang shutdown → `NotifyContext` |
| Non-revealing auth errors | Alpine needed `unsafe-eval` → CSP build |
| `httptest` game-loop walk with fixed clock | `X-Forwarded-For` trusted blindly → trust only the proxy |
| No Node toolchain, `CGO_ENABLED=0` | `cp` backup of a WAL DB → `.backup` / `VACUUM INTO` / Litestream |
| | Mixed time formats, `map[string]any`, probes logged, shop items in DB with stringly `effect_type` |
| | Features added faster than played. Docs drifted (repetitive README, dead § refs) |

---

## 5. Planning & Git

**Process (Working, light):**
- GitHub Issues for everything (one issue ≈ one PR).
- **Projects board:** Backlog → Ready → In progress → In review → Done.
- Milestones = the roadmap steps.
- Labels: `type:*`, `area:*`, `security`.
- **User stories** with Given/When/Then acceptance criteria (they become tests).
- **Definition of Done** in the PR template.
- Exercise: write and size (S/M/L) the issues for a milestone before coding,
  then compare with reality.

**Git (Deep):**

| Strategy | Use when |
| --- | --- |
| **Trunk-based / GitHub Flow**: short-lived branches → PR → squash merge, `main` always releasable | **Use this** (ADR-1) |
| GitFlow (`develop`, `release/*`, `hotfix/*`) | Scheduled releases of shipped software. Know it to explain why not |
| Release branches | Several supported versions. Practise once for a hotfix |
| Environment branches | Anti-pattern. Promote artifacts, not merges |

- **Conventions:**
  - Branches like `feat/…`, `fix/…`, `ci/…`, `docs/…`.
  - **Conventional Commits** (`feat(game): add container upgrade`,
    `fix(auth): …`, breaking: `feat!:` or a `BREAKING CHANGE:` footer). One-line
    imperative subject, body only for the *why*.
  - Squash merge, linear history.
  - **SemVer** annotated and signed tags (`v0.x` until launch).
  - Never commit with red tests.
- **Protection:**
  - Rulesets on `main`: PR + required checks + linear history, no force
    push/delete.
  - SSH-signed commits, CODEOWNERS.
  - `.gitattributes` (`eol=lf`). Never commit `*.db`, `.env` or binaries.
- **Practise on purpose:**
  - `commit --fixup` + `rebase -i --autosquash`
  - `bisect run go test ./internal/game/...` on a deliberately planted bug
  - `reflog` recovery after `reset --hard`
  - `cherry-pick` to a release branch
  - `worktree`
  - `switch` / `restore` / `stash`
  - `blame -w -C`, `log -S`, `log -L`
  - a deliberate merge conflict
  - `revert` vs `reset` (never rewrite shared history)
  - hooks in `.githooks/` (`core.hooksPath`): pre-commit gofmt + gitleaks,
    commit-msg check

---

## 6. Documentation (Deep)

Docs-as-code: in the repo, reviewed in PRs, part of the DoD, linted in CI
(`markdownlint-cli2`, `lychee` for links, optional Vale). Sort content with
**Diátaxis**: tutorial, how-to, reference, explanation.

| File | Purpose |
| --- | --- |
| `README.md` | What it is, screenshot, quickstart (3 commands), test, status, links. Short |
| `CONTRIBUTING.md` | Workflow, conventions, checks, DoD |
| `CHANGELOG.md` | Keep a Changelog, `[Unreleased]` until launch, one user-facing sentence per entry (Added/Changed/Fixed/Removed). **By hand first**, then automate (release-please / git-cliff) and compare what got lost |
| `ROADMAP.md` or board | The single to-do list |
| `docs/adr/NNNN-*.md` | Context, decision, options, consequences, "revisit when". Never edit after acceptance, supersede instead |
| `docs/architecture.md` | Mermaid: C4 context, C4 container, deployment view (VPS → k3s → pod + sidecar → PVC, Ingress, object storage), key invariants |
| `docs/operations.md` | Environments, env vars, SLOs, alerts, backups, RPO/RTO |
| `docs/runbooks/` | Restore DB, rotate a secret, rollback, … (with "last verified" date) |
| `docs/postmortems/`, `docs/playtests/` | Incident reviews, balancing notes |
| `docs/threat-model.md`, `SECURITY.md` | Security design, reporting + vuln SLA |
| Go doc comments | Every exported identifier. `go doc ./internal/game` should read well |
| In-game help (German) | Mechanics explained for players (collect, job, container, skills) |

**First ADRs:**
1. Trunk-based development
2. htmx instead of an SPA
3. SQLite instead of Postgres
4. Sessions instead of JWT
5. Monolith
6. Lazy resolution
7. Cents + ledger
8. Single replica in k8s
9. GitOps
10. Automated releases
11. SemVer meaning for the game
12. Secrets tool

**Writing habits:**
- Know your reader and their task.
- Lead with the answer.
- Date anything that goes stale.
- If you explain something twice, write it down.
- Free course: Google *Technical Writing One/Two*.

---

## 7. Testing & QA (Deep)

| Level | What / tool |
| --- | --- |
| Static | `gofmt`, `go vet`, golangci-lint v2 (staticcheck, gosec, errcheck) |
| Unit | table tests for every formula at its boundaries (level 0/max, mood 0/100, roll 0/0.999, overflow exactly at cap) |
| Invariants | capacity ≥ next upgrade cost, costs strictly increasing, catch-up clamped |
| Fuzz | `go test -fuzz` on parsers/formatters (short in PR, long nightly) |
| Integration | one test per action against SQLite (`:memory:` / `t.TempDir()`): happy path + critical edge (double resolve, overflow, busy, not enough money) |
| Concurrency | two parallel resolves/payouts → exactly one booking. Always `-race` |
| HTTP journeys | `httptest` + `cookiejar`: auth (register → verify → login → logout) and game loop (create → shift → advance clock → resolve → upgrade). Assert behaviour, not markup |
| Templates | test that all templates parse |
| Time logic | injected clock, `testing/synctest` |
| Benchmarks | `go test -bench` + `benchstat` on resolve and dashboard render |
| Load | k6 with ~50 virtual players + `pprof`. Find the SQLite write ceiling |
| Smoke | after every deploy: `/healthz`, `/readyz`, login page |
| DAST / a11y | ZAP baseline on staging. Lighthouse (mobile) before releases |
| E2E browser | optional: Playwright for 1–3 flows, nightly |
| Manual | exploratory charters + playtest log (`docs/playtests/DATE.md`: what you did, numbers, what felt off, tuning) |

**QA practices:**
- **Quality gates** in CI block the merge.
- Measure and publish coverage, but **don't gate on a percentage**.
- Flaky test → quarantine + issue the same day.
- Bug → failing test first, and it stays as a regression test.
- `make seed` for a demo account.
- Optional: mutation testing (`gremlins`) once on `game`.
- New logic with a branch gets a test, a one-liner doesn't.

---

## 8. CI/CD (Deep)

**Concepts to master:**
- CI vs. continuous delivery vs. continuous deployment.
- **Build once, promote the same image digest.** Never `latest` in prod.
- Pipeline as code. Fail fast (cheap checks first), PR feedback under 5 min,
  caching, parallel jobs, matrix builds.
- **Push deploy** (CI runs kubectl) vs. **GitOps pull** (the cluster pulls from
  Git, so CI holds no cluster credentials).
- Environments with approvals, concurrency groups (no parallel deploys), OIDC.
- Deployment strategies: recreate / rolling / blue-green / canary (Argo Rollouts,
  awareness).
- **Expand/contract migrations** so the old version works with the new
  schema. Back up before migrating.
- Image rollback is easy, but a schema rollback isn't. Know when to roll back and when to
  roll forward.
- Version info baked in: `-ldflags "-X main.version=$GITHUB_SHA"`, shown on
  `/healthz` or in the footer.

**Pipelines (GitHub Actions):**
- **PR:**
  - gofmt / vet / lint → tests with `-race` → govulncheck + gitleaks →
    markdownlint / lychee / kubeconform → `docker build` + Trivy (no push).
- **main:**
  - same checks → multi-arch image (buildx, `linux/amd64,arm64`, Hetzner
    CAX = ARM) to **GHCR**, tagged `sha-<short>` + `main`, via
    `docker/metadata-action` + `build-push-action` with `GITHUB_TOKEN` →
    SBOM + cosign + provenance → commit the digest to the staging overlay.
- **Release (tag `v*`):**
  - release-please PR merged → tag → retag the same digest as `vX.Y.Z` →
    GitHub Release → PR bumping the prod overlay → approval.
- **GitOps:**
  - Argo CD syncs staging/prod → smoke test → notify (mail/Discord/ntfy) on
    failure.
- **Nightly (`schedule`):**
  - long fuzz, ZAP on staging, Trivy rescan of the prod image, backup restore
    check.

**Build-up order (one PR each, with a changelog entry):**
1. vet + test, required check
2. setup-go cache, `-race`, golangci-lint
3. parallel jobs + `concurrency` cancel
4. `make ci` as the single entry point for local and CI
5. image build on PR, push on `main`
6. multi-arch + BuildKit cache (`type=gha`)
7. security stages (§11)
8. release automation
9. first CD: SSH + `docker compose pull && up -d`
10. Argo CD
11. prod promotion with environment approval
12. post-deploy smoke + alert
13. reusable workflows vs composite actions
14. nightly workflow
15. *optional:* preview environment per PR (ephemeral namespace)
16. *optional:* `act` locally

**Pitfalls:**
- Fork PRs get no secrets, and `pull_request_target` with a PR checkout is dangerous.
- Cache key design and cache poisoning.
- Tool version drift between local and CI. Pin via `go.mod` toolchain + `tool`
  directives.

GitLab CI (stages, runners) and Jenkins (Jenkinsfile, agents) use the same
concepts.

---

## 9. Docker, Kubernetes & IaC (Working → Deep)

**Docker / Compose (first deploy):**
- Multi-stage: `golang:alpine` → `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"`
  → `distroless/static:nonroot` (or alpine + non-root user).
- Tailwind is built inside the image (download the linux binary there).
- `.dockerignore`.
- Compose: app + Caddy (`{$SITE_DOMAIN} { reverse_proxy app:8080 }`), named
  volume for `/data`.
- Ops tasks: domain, VPS, DNS, `.env`, mail provider + SPF/DKIM/DMARC.

**Kubernetes:** kind/k3d locally, then single-node **k3s** on a VPS.

| Resource | Notes |
| --- | --- |
| Namespace | `studigame`, `studigame-staging` |
| Deployment | **`replicas: 1`, `strategy: Recreate`**. SQLite = one writer on an RWO volume. A rolling update would start a second pod on it |
| PVC | RWO, holds the DB |
| Service + Ingress / Gateway API | Traefik ships with k3s. ingress-nginx was announced for retirement, so check its status first |
| cert-manager + ClusterIssuer | Let's Encrypt, replaces Caddy |
| ConfigMap / Secret | Non-secret env / SMTP creds via **SOPS + age** or **Sealed Secrets** (awareness: External Secrets Operator) |
| Probes | liveness `/healthz`, readiness `/readyz`, startup if migrations are slow |
| resources, securityContext | requests/limits, `runAsNonRoot`, `readOnlyRootFilesystem`, drop capabilities, `fsGroup` |
| Litestream sidecar | Continuous backup to object storage (Hetzner, B2, MinIO). Restore via an init container |
| NetworkPolicy | default deny. Allow DNS, SMTP, object storage |

- **Kustomize** `base/` + `overlays/{staging,prod}`. Use Helm to install charts
  (cert-manager, kube-prometheus-stack, Argo CD). Writing your own chart is
  optional.
- **Argo CD** app-of-apps (or Flux), so the whole cluster is declared in Git.
- In-process state (rate limiter, template cache) is per pod. That's correct
  with 1 replica.
- **Optional scale-out track (learning only):**
  - Postgres (CloudNativePG, driver `pgx/v5/stdlib`, `?` → `$1`).
  - `replicas: 2+`, RollingUpdate, PDB, HPA.
  - Rate limiter shared or per pod.
  - Sweeper as a CronJob or behind an advisory lock.

**IaC:**
- **OpenTofu** + `hetznercloud/hcloud`: server, firewall, SSH key, volume, DNS.
  Remote state with locking (S3 backend). `fmt`/`validate`/`plan` on PR (as a
  comment), `apply` on merge with approval.
- Bootstrap with **cloud-init** (simple) or **Ansible** (inventory, playbooks,
  roles, idempotency, common in jobs).
- Validate with `kubeconform`, `kube-linter`, `tflint`, Trivy config.

**Environments:**

| Env | Where | Deployed by | Data |
| --- | --- | --- | --- |
| local | `make run` / kind | you | throwaway + `make seed` |
| staging | k3s namespace | GitOps on every `main` merge | seeded, wipeable |
| prod | k3s namespace (later its own VPS) | GitOps after release + approval | real, backed up |

Same image digest everywhere. Config only via env (12-factor).

---

## 10. Artifacts & dependencies (Awareness)

- **Concepts:**
  - An artifact is an immutable, versioned build output (image, binary, SBOM,
    chart).
  - **Promote, don't rebuild.**
  - Retention policies.
  - Proxy/cache repositories (rate limits, availability, audit).
  - Provenance.
- **Here:**
  - Images in GHCR (with cleanup of old `sha-*` tags).
  - Go modules via `GOPROXY` + `GOSUMDB` (Go's built-in checksummed artifact
    repo, worth reading about).
  - Optional GoReleaser binaries in GitHub Releases.
  - SBOMs and signatures stored as OCI artifacts next to the image.
  - Vendored JS with version + SHA-256 in `static/js/VERSIONS`.
- **Enterprise:** Artifactory and Nexus are universal repositories, Harbor is the
  CNCF OCI registry with scanning. Knowing what they do and why companies use them
  is enough.
- **Dependency updates:** Dependabot (or Renovate) for `gomod`, `docker` and
  `github-actions`. Group minor/patch updates weekly, review majors by hand.

---

## 11. DevSecOps (Working → Deep)

| Stage | Control |
| --- | --- |
| Design | 1-page threat model (data flow + STRIDE), OWASP Top 10 + a few ASVS L1 items mapped to the app |
| Commit / repo | gitleaks hook, GitHub secret scanning + push protection, Dependabot alerts |
| PR | CodeQL, gosec, **govulncheck** (reachability-aware), Trivy config scan (Dockerfile, k8s, tofu) |
| Build | Trivy image scan, SBOM (Syft/Trivy), **cosign** keyless signing, SLSA provenance (`actions/attest-build-provenance`) |
| Deploy | Pod Security Admission `restricted`, **Kyverno** (signed images only, no `latest`, limits required), NetworkPolicy |
| Runtime | ZAP baseline, nightly rescan of the deployed image, OpenSSF Scorecard. Awareness: Falco |

**Hardening GitHub Actions:**
- `permissions: contents: read` by default, more per job only where needed.
- Pin third-party actions to a commit SHA (Dependabot updates them).
- Pass untrusted input (`github.event.*.title`) via `env:`, never inline in `run:`.
- No `pull_request_target` with a PR checkout.
- OIDC (`id-token: write`) instead of stored tokens.
- Prod environment with required reviewers.

**Secrets:**
- `.env` is git-ignored. `.env.example` is committed.
- CI secrets are scoped to environments. k8s uses SOPS or Sealed Secrets.
- Rotate SMTP creds once end-to-end and write the runbook.

**App security checklist:**
- Auto-escaping on, never `template.HTML` with user input.
- CSRF on every POST, rate limit on every auth and action route.
- No price, result or amount from the form.
- Tokens are 32 random bytes, only the SHA-256 is stored, TTL, single use. Reset
  invalidates all sessions.
- Cookies `HttpOnly` / `Secure` (prod) / `SameSite=Lax`.
- Strict CSP (`script-src 'self'`, no inline, **no `unsafe-eval`**), `nosniff`,
  `frame-ancestors 'none'`, Referrer-Policy. HSTS once on a real domain.
- Login, reset and register never reveal whether an e-mail exists.
- Only a UNIQUE violation means "e-mail taken". Other DB errors → 500.
- Panic recovery, timeouts, `MaxBytesReader`. Trust XFF only from the proxy.
- Non-root container.

**Vuln handling:** critical within 48 h, high within 7 days. Every suppression
(`.trivyignore`, …) needs a reason and an expiry date.

---

## 12. Logging & observability (Working → Deep)

- **Logs:**
  - slog JSON to stdout.
  - Fixed fields: `request_id`, `user_id`, `character_id`, `route`, `status`,
    `duration_ms`, `version`.
  - Request-ID middleware (accept or generate `X-Request-ID`, put it in
    `context`, echo it in the response).
  - Levels: ERROR = a human must act, WARN = unexpected but handled, INFO =
    business events, DEBUG off in prod.
  - Never log passwords, tokens, cookies or full e-mails. Don't log probes.
  - The **ledger** is the money audit trail, logs are for ops.
  - Pipeline: stdout → **Grafana Alloy** (Promtail's successor) → **Loki** →
    **Grafana**.
- **Metrics:**
  - Prometheus `/metrics`, not exposed via Ingress.
  - **RED** (rate, errors, duration histogram) per route *pattern*. **USE** /
    Go runtime metrics come for free.
  - Business counters: `actions_started_total{kind}`,
    `actions_resolved_total{kind}`, `money_booked_cents_total{reason}`,
    `charity_overflow_cents_total`, `registrations_total`.
  - **No high-cardinality labels** (no user IDs, e-mails, raw paths).
  - Stack: kube-prometheus-stack (Prometheus, Alertmanager, Grafana).
- **Dashboard:** traffic, errors, latency, actions per kind, money in vs.
  charity overflow, pod restarts, PVC usage.
- **Traces (optional):** OpenTelemetry → Tempo. Spans for request, DB TX,
  template render.
- **SLOs:**
  - Examples: 99.5 % of dashboard requests non-5xx over 30 days, p95 < 300 ms,
    last backup < 26 h.
  - Error budget: if it's burnt, fix reliability before features.
- **Alerting:**
  - Alert on symptoms, and only if actionable. Every alert links a runbook.
  - Route to your phone (mail, Discord or ntfy).
  - Plus an external uptime check.
- **Profiling:** `net/http/pprof` on an internal port. Read a CPU and heap
  profile during the k6 test.

---

## 13. Releases & operations (Working)

- **SemVer for a game** (write it as an ADR): MAJOR = world reset / incompatible
  data, MINOR = new mechanics, PATCH = fixes and balancing. `0.x` until the
  closed beta, `1.0.0` at launch.
- **Release notes** (curated, German, in-game "Neuigkeiten") ≠ **changelog**
  (complete).
- **Release checklist** (issue template):
  - milestone done
  - release PR reviewed
  - staging runs the same digest + smoke + playtest
  - migrations backward compatible, backup < 1 h old
  - tag + release
  - prod approval, smoke, watch dashboards for 30 min
  - release notes published
- **Hotfix:** fix on `main` → patch tag. If `main` has unreleased work, cut
  `release/x.y` from the tag and cherry-pick.
- **Feature flags** (env or DB table) let you merge unfinished work to trunk and
  release it later. Awareness: OpenFeature.
- **Backups/DR:**
  - Define RPO (e.g. 1 min with Litestream) and RTO (e.g. 30 min).
  - Nightly automated restore check (`PRAGMA integrity_check` + row count).
  - Do a real restore drill before launch.
- **Incidents:**
  - SEV1 down / SEV2 core feature broken / SEV3 degraded.
  - Keep a timeline in an issue.
  - Write a **blameless postmortem** with action items as issues.
- **Game days on staging:**
  - kill the pod mid-shift
  - delete the PVC and restore (measure the real RTO)
  - deploy an image that fails readiness
  - break TLS
  - revoke the SMTP secret
  - fill the disk
  - **destroy and rebuild the whole cluster from Git + tofu + backup**
- **DORA metrics** (monthly in `docs/metrics.md`): deployment frequency, lead
  time for changes, change failure rate, time to restore.

**Dev experience:**
- `make help` as the single entry point (`run`, `test`, `lint`, `css`, `seed`,
  `docker`, `k8s-validate`, `ci`). CI calls the same targets.
- Pin tools with `go get -tool`.
- `.editorconfig`, `.golangci.yml`.
- Optional: devcontainer, Tilt/Skaffold.

---

## 14. Roadmap: build + learn together

Each step ends running, tested, documented and committed.

| # | Build | Add SDLC/DevOps practices |
| --- | --- | --- |
| 0 | Repo setup | README, LICENSE, `.gitignore`, `.editorconfig`, rulesets, signed commits, issue/PR templates, board, CHANGELOG, ADR-1 |
| M0 | Hello server: config, slog, `/healthz` + `/readyz`, graceful shutdown | CI steps 1–4, Makefile, git hooks, logging conventions |
| M1 | Templates + embed, landing + waitlist (htmx), SQLite + migration runner | ADRs 2–3, architecture doc, Dependabot, gitleaks, govulncheck |
| M2 | Auth: register/login/logout, sessions, CSRF, rate limiter, double opt-in, reset, sweeper | Threat model, SECURITY.md, CodeQL, fuzzing, auth journey test, ADR-4 |
| M3 | Pure `game` rules, character creation (subject, name, start capital via ledger), dashboard + status bar | Table + invariant tests, coverage report, Go doc comments, ADR-7 |
| M4 | Core loop: collect, booking + ledger + charity, container upgrade, countdown → **playable** | Integration + concurrency tests, Docker + Compose deploy, image to GHCR, `v0.1.0`, playtest log, backup runbook. **Play it for a week** |
| M5 | Job + crime, typed outcomes | release-please, feature flags, benchmarks, k6 + pprof |
| M6 | Skills, BAföG, coffee, decide ECTS → Klausur, shop/housing redesign | k3s, Kustomize, cert-manager, secrets, Litestream, Argo CD, staging → prod approval, smoke tests |
| M6+ | Balancing pass | Supply chain (SBOM, cosign, Kyverno), PSA, NetworkPolicy, ZAP, OpenTofu, metrics/Loki/Grafana, SLOs + alerts, first game day + postmortem |
| M7 | Social: highscore, public IDs/profiles, Notenkonkurrenz, WGs, messages/friends/block/report, referral | Hotfix drill, tracing, Scorecard, DORA tracking, cluster rebuild |
| Launch | PWA manifest, perf/a11y pass, in-game help, closed beta via waitlist | Docs review, restore drill, legal checklist, `v1.0.0` |

**Target repo layout:**

```
.github/   workflows/ (ci, release, nightly, deploy), ISSUE_TEMPLATE/, PULL_REQUEST_TEMPLATE.md, CODEOWNERS, dependabot.yml
.githooks/ cmd/ internal/ static/
deploy/    compose/ (compose + Caddyfile), k8s/base/, k8s/overlays/{staging,prod}/, argocd/
infra/     OpenTofu, cloud-init or ansible/
docs/      adr/, architecture.md, operations.md, runbooks/, postmortems/, playtests/, threat-model.md, metrics.md
README.md CONTRIBUTING.md CHANGELOG.md SECURITY.md LICENSE Makefile Dockerfile .golangci.yml .editorconfig .dockerignore
```

---

## 15. Template skeletons

- **PR template:** What & why (closes #) · How to verify · DoD checklist: tests,
  `make ci` green, docs/ADR updated, CHANGELOG entry, no new secrets/findings,
  migration backward compatible.
- **User story:** As a … I want … so that … · acceptance criteria
  (Given/When/Then, including the concurrency case) · out of scope.
- **Bug report:** observed · expected · steps · environment + version + time
  (UTC) + request ID · logs.
- **ADR:** title · status/date/supersedes · context · decision · considered
  options · consequences (+/−) · revisit when.
- **Runbook:** title · last verified · duration · severity · when to use ·
  preconditions/access · steps · verify · if this fails.
- **Postmortem:** impact · severity · detected by · timeline (UTC) · root cause
  (5 whys) · went well / badly / lucky · action items (issue, owner, date).

---

## 16. Conventions, legal & backlog

**Conventions:**
- **Language:** player-facing text in German, "du", dry. Code, identifiers,
  comments and commits in English. Repo docs in your choice of language.
- Keep player strings out of deep logic (a lookup function), but **don't build
  i18n**. The humour doesn't translate.
- **After every task:**
  - `go build ./... && go vet ./... && go test ./...` green.
  - README updated if structure, commands, env vars or scope changed.
  - CHANGELOG entry.
  - Roadmap updated.
  - After a whole phase, update the README status.
- Ask (or decide consciously) on balancing numbers, naming, cost-causing and
  legal topics.

**Legal & pre-launch (Germany, not legal advice):**
- Impressum (§5 DDG) + privacy policy (Art. 13 GDPR), required from the
  waitlist onwards.
- **No cookie banner** as long as there's no tracking/analytics (§25 TDDDG) and
  fonts are self-hosted, with no CDNs.
- Double opt-in for waitlist and registration. "Account löschen" really deletes.
- AGB, withdrawal terms and tax only once you sell something. Prefer a Merchant
  of Record (Paddle).
- Choose and contact a charity partner before promising real donations.
- A one-time legal check before go-live is worth it (fines up to 50 k€).
- Moderation rules before chat and messages go live.

**Parked (not MVP):**
- cleanliness stat
- Erasmus event
- English / i18n
- own forum (Discord instead)
- more towns as editions (once Studistadt gets boring)

**Idea backlog** (default: won't be built until a player or a number justifies it):
- **Gameplay:**
  - seasons/leagues with resets and cosmetic rewards
  - daily small tasks
  - weather affecting yield
  - pets and pet fights
  - instruments with a mini-game
  - WG territory control
  - campus places as action labels
  - Zweitstudium, Fachwechsel
- **Community:**
  - public profile with shareable signature
  - weekly e-mail recap (opt-in)
  - WG wars and alliances, finer WG roles
  - Heimatstadt leaderboards, WG search by city, city flavour texts (never
    shorten the city list)
  - nostalgia quiz
- **Tech:**
  - admin UI instead of an SQL console
  - highscore snapshots (only when `ORDER BY` is measurably slow)
  - retention metrics
  - full-text search on messages
  - Postgres (only when SQLite hits a measured limit)

**Rejected:**
- web push (calm tone, no addiction mechanics)
- DaisyUI
- JWT
- English localisation
- pay-to-win
- separate `internal/actions` package (v1 decision, revisit it yourself)

---

## 17. Resources

- **Git:** Pro Git (git-scm.com/book), trunkbaseddevelopment.com,
  conventionalcommits.org, semver.org
- **Docs:** keepachangelog.com, diataxis.fr, adr.github.io (MADR), c4model.com,
  Google Technical Writing courses
- **Go:** go.dev/doc (Effective Go, testing/fuzzing tutorials), the Go blog
- **CI/CD & DevOps:** *Accelerate*, *Continuous Delivery*, dora.dev,
  12factor.net, GitHub Actions docs + "Security hardening for GitHub Actions"
- **SRE:** sre.google (SRE Book + Workbook, free), Prometheus instrumentation
  best practices, OpenTelemetry docs
- **Security:** OWASP Top 10 / ASVS / Cheat Sheets, slsa.dev, OpenSSF Scorecard,
  Sigstore docs
- **Kubernetes:** kubernetes.io (tasks), k3s, Argo CD and Kustomize docs, later
  *Kubernetes the Hard Way*
- **Certifications (optional):** GitHub Actions, CKAD → CKA
