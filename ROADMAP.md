# Roadmap

Was gebaut wird, in welcher Reihenfolge, und was bewusst noch nicht.
Detailbegründungen stehen in [studigame-plan.md](./studigame-plan.md).

Legende: 🙋 = manuelle Aufgabe (nicht per Code lösbar) · 🚧 = blockiert etwas

---

## Jetzt

Kleine Fixes und das Fundament, das später teuer nachzurüsten wäre.

- [x] `Dockerfile` auf Go 1.27 (aktuell 1.23, Build schlägt fehl) 🚧
- [x] CSRF + Rate-Limit auf `POST /waitlist`
- [x] Sessions beim Passwort-Reset invalidieren
- [x] `Register`: nur UNIQUE-Verletzung als „E-Mail vergeben“ werten
- [x] Migrations-System (nummerierte SQL-Dateien + `schema_migrations`) 🚧 vor der ersten echten Nutzer-DB
- [x] Server-Timeouts, Graceful Shutdown, Panic-Recovery
- [x] `GET /healthz`, Security-Header, strukturiertes Logging
- [x] htmx lokal ausliefern statt per CDN
- [x] Sweeper: abgelaufene Sessions und Reset-Tokens löschen
- [x] CI: `go vet` + `go test -race` bei jedem Push
- [x] Backup-Skript für die SQLite-Datei
- [x] Name entschieden: Spiel = Studigame, Uni-Stadt = Studistadt, Studienfächer
  = Informatik/BWL/Jura/Geisteswissenschaften
- [ ] 🙋 Domain registrieren 🚧 blockiert alles Öffentliche – erledigt der
  Entwickler später beim Deploy
- [ ] 🙋 VPS bestellen, DNS setzen, `.env` befüllen
- [ ] 🙋 Mailversand-Anbieter wählen, SPF/DKIM setzen
- [ ] 🙋 Impressum + Datenschutzerklärung
- [ ] 🙋 Backup-Restore einmal testen
- [x] uuid für öffentliche IDs: entfällt, keine Route trägt eine Charakter-ID
  (alle Routen laufen über die Session). Erst bei öffentlichen Profilen relevant.
  Vorher erwog: verhindert Enumeration. `github.com/google/uuid` ist bereits im
  `go.sum` (Transitiv-Dependency von `modernc.org/sqlite`), aber unbenutzt;
  müsste von `// indirect` zu direkter Dependency werden
- [x] Tailwind vs. handgeschriebenes CSS entschieden: Tailwind bleibt

## Code-Review-Funde (vor Beta abarbeiten)

Aus dem Review des Codestands nach dem Shop-Commit. Reihenfolge = Abarbeitung.

**1. Geld-Races und BAföG**
- [x] Lost Update bei Geld: Abbuchungen relativ (`money_cents - ? WHERE money_cents >= ?`) statt stale Absolutwert — Shop, Container-Upgrade, Kaffee, Skills, Schichtabrechnung
- [x] `UpgradeContainer`: Level atomar hochzählen (Doppelklick kauft sonst zwei Stufen zum Preis von einer)
- [x] `PurchaseItem`: Mengenprüfung innerhalb der Transaktion
- [x] `CancelSkillTraining`: Doppel-Abbruch zahlt zweimal zurück
- [x] `StartSkillTraining`: Level vor dem Auflösen fälliger Trainings gelesen — Spieler zahlt für ein Level, das er schon hat
- [x] BAföG-Clamp: Zeiger bei Kappung bis `now` vorziehen, sonst holt jeder Reload 4 weitere Perioden nach

**2. Stimmung und Shop-Wirkung**
- [x] Stimmung regeneriert (+5 pro Stunde ohne Schicht, Startwert, 🙋 Balancing — `game.MoodRegenPerHour`)
- [x] Shop-Effekte wirken: `container_bonus` erhöht Kapazität, `passive_income` zahlt stündlich, `mood_bonus` senkt Erschöpfung in Prozent (Startwerte, 🙋 Balancing)
- [x] `subject_required` im Shop prüfen oder Spalte streichen (Spalte gestrichen, Migration 0013)

**3. Auth und Logs**
- [x] Login: gleiche Antwort und bcrypt-Zeit für unbekannt / unbestätigt / falsches Passwort
- [x] „Bestätigungsmail erneut senden“ (Token läuft nach 1 h ab, Account sonst tot)
- [x] Rohe Reset-Tokens und E-Mail-Adressen nicht ins Log
- [x] `ResetPassword` / `VerifyEmail` atomar (Transaktion, Einmal-Token)
- [x] Passwort-Obergrenze (72 Byte, bcrypt) mit sauberer Meldung, `MaxBytesReader` auf Formularen
- [x] Zeitzone: DB-Zeiten immer UTC, unabhängig von `TZ` des Servers
- [x] 🙋 Go-Port nur auf 127.0.0.1 binden (`X-Forwarded-For` wird vertraut) — in `DEVOPS-GUIDE.md` vermerken

**4. ECTS sauber trennen**
- [x] ECTS nicht aus Anzeigetext zurückparsen und nicht ins Geld-Ledger buchen
- [x] ECTS-Stand anzeigen (Klausur-Seite, Dashboard)
- [x] Auskommentierten Code in `klausur_store.go` löschen

**5. Aufräumen**
- [x] `gofmt -w .` (`game/klausur.go`, `klausur_test.go`, `shop_handlers.go`)
- [x] Leerzeichen im Nebenverdienst-Text („…€ übrig“)
- [x] Shop-Code verschlanken (drei gleiche Schleifen in `game/shop.go`, ungenutztes `CanPurchase`)
- [x] Zeichen statt Bytes bei der Namenslänge (`character_handlers.go`)

---

## UI-Redesign: Studigame-Look

Abstand zum Original-Pennergame-Look (asphalt/pfandglas/verkehrsorange/muenzgold),
modern-minimal, mobile-first. UI-Probleme nur hier, kein neuer Spielinhalt.
Ein Bastel-Kosmetik-Reset, den ich mir mit dem Nutzer vorher als Richtung abgesteckt
habe (nach Abstimmung in der Session).

### Entscheidungen (festgehalten, damit der nächste Agent nicht nachbaut)

- [x] **daisyUI nicht verwenden** — neu bestätigt: Phase-2-Komponenten sind bereits
  handgebaute Tailwind-Partials in `internal/web/components/`; eigene Seite im
  Ideen-Halde sagt dasselbe. Neue Dependency nur um funktionierenden Code zu
  ersetzen wäre Rückschritt (CLAUDE.md-Regeln). Light/Dark reichen als Themes,
  daisyUI würde dafür einen kompletten Framework-Umzug brauchen.
- [x] **Light/Dark-Thema nativ** über CSS-Variablen + `data-theme`-Attribut
  (dasselbe Prinzip wie daisyUI-Themes, nur ohne deren Framework). Default: Light.
  Toggle-Button in der Statusleiste, Auswahl im `localStorage`.
- [x] **Icons ohne JS und ohne Font**: Lucide-SVGs inline als Go-Template-Helper
  `icon "name"` (vendored, gleiches Muster wie htmx/alpine). Kein Icon-Font,
  kein npm, kein `<img>`-Netzwerkzug.
- [x] Palette-Richtung: frisches Uni-Vokabular (Indigo-Primär + Lime/Lime-Akzent),
  neutrale Slate-Neutrals statt brauner Pennergame-Töne. Zustandsfarben
  success/danger/warning bleiben.

### Aufgaben (Reihenfolge)

- [x] Neues Farb-/Token-System in `tailwind.config.js` und `static/css/input.css`
  (semantische Tokens statt `asphalt`/`pappe`/`pfandglas`/`muenzgold`/`backstein`)
- [x] Light/Dark-Variablen in `:root` + `[data-theme="dark"]`, `theme.js`-Toggle
  (vanilla, `localStorage`)
- [x] `icon`-Helper in Go (`internal/web/templates.go`) + SVG-Partials für die
  benötigten Icons (Haus, Geldbörse, Buch, Kaffee, Abschluss, Blitz, …)
- [x] Basiskomponenten auf neues Design: Button, Card, Badge, Formularfelder,
  Flashes, Stat/Stat-Wert, Progress-Bar, Skeleton, Loading
- [x] Statusleiste: Geld-Behälter als sichtbarer **Progress-Bar + Stat** im Header
  (Füllstand %, nächste Stufe, pot), Thema-Toggle, Icons statt Text
- [x] Bottom-Navigation: Icons + Label, aktiver Zustand, (Logout bleibt)
- [x] Richtung/Abstimmung ansehen, dann Seiten einzeln umstellen:
  Dashboard (Stats + Progress), Aktionen (Shift-Cards), Weiterbildung,
  Landing/Warteliste, Login/Register
- [x] `static/css/app.css` aufräumen: weggeworfene Pennergame-Klassen gelöscht,
  Rest auf Tokens umgezogen
- [x] CSS gebaut, `go test ./...` grün, README/CHANGELOG nachgezogen

## Als Nächstes

- [x] Uni-Stadt bei der Registrierung erfassen (feste Liste, reine Anzeige,
  keine Spielfunktion — später ggf. mit Gameplay verknüpfen)
- [x] Double-Opt-In für Warteliste und Registrierung
- [x] Echter SMTP-Versand statt Konsolen-Mailer
- [x] Design-System: Tailwind-Tokens, Basiskomponenten, Statusleiste
- [x] Geld-Behälter als sichtbares Fortschritts-Element im Header
- [x] Bottom-Navigation, mobile-first
- [x] Bestehende Seiten auf das finale Design umstellen
- [x] Docker-Build reparieren: `tailwindcss`-Binary wird im Build für die Zielplattform
  geladen statt aus dem Repo genommen
- [x] `tailwindcss`-Binary in `.gitignore` aufnehmen (80 MB, nicht plattformportabel)
- [x] Emoji aus `bottom_nav.html`/`statusbar.html` entfernen — widerspricht der eigenen
  Regel „keine Emoji in der Oberfläche" (CLAUDE.md)
- [x] Plan-Checkbox „Fonts selbst hosten" korrigiert — System-Fonts sind der aktuelle Stand
- [x] Phase-3/4-Tabellen (`0002_phase3_tables.sql`) gelöscht — kein Code nutzte sie, MVP-Regel "kein Code für später"

## Bewertung & Nacharbeiten (aus der Doku-Analyse)

Bewusst von der Kernschleife getrennt: hier geht es nicht um neuen Spielinhalt,
sondern um dokumentierte Lücken und Balance-Fälle, die ein späterer Leser sonst
übersieht. Jeder Punkt bezieht sich auf eine konkrete Fundstelle.

- [x] BAföG-Passivertrag: Auszahlung von 30 € auf 20 € je 6 h gesenkt
  (≈3,33 €/h, `internal/game/bafog.go`), Abstand zum Job 8 €/h. Startwert,
  Feinjustierung mit dem Balancing 🙋
- [x] Endgame-Loch nach Behälter-Maxl: Campus-Shop (Wohnungen, Ausrüstung,
  Instrumente) ist der Geld-Sink nach dem Bausparvertrag. Feintuning der Preise 🙋
- [x] Schicksal der gestrichenen 6 Skills: **aufgeschoben, nicht gestrichen**.
  Kommen als „Später"-Umfang zurück (Bildungsstufe, Sozialkontakte, Musik,
  Taschendiebstahl, ATT, DEF). Wohnungsleiter und Fach-Boni hängen daran.
  Festgehalten in `studigame-plan.md` §3
- [x] Kaffee-Boost (`0009_coffee.sql`, `shop.go`) zum Abschluss in CHANGELOG &
  ROADMAP nachziehen — uncommitted WIP, nicht vergessen beim Commit

## Danach — Kernschleife

Der eigentliche Kern. Alles darunter hängt am Fundament aus „Jetzt".

- [x] `characters`, `editions`, `subjects`, `ledger`, `actions` (Studigame-Umbau:
  `districts` → `subjects`, siehe Rebrand-Abschnitt)
- [x] Reine Spiellogik in `internal/game` (keine DB, keine Uhr, testbar)
- [x] Charaktererstellung: Studienfach wählen, Name setzen, Startkapital über das Ledger
- [x] Sammeln-Aktion mit Dauerwahl und Lazy-Tick-Auflösung
- [x] Geld-Behälter mit Kappung, Überlauf fließt in den Charity-Pool
- [x] **Behälter-Upgrade als erster Geld-Sink** (kein weiterer Bar-Punkt!): Kauf
  "Behälter aufrüsten", `container_level` +1, Ledger-Debit, überlineare Kosten-
  kurve (`Basiswert × Level^1.8`) in `internal/game`, tabellengetestet.
  Begründung: ohne Ausgabe-Möglichkeit ist der Charity-Überlauf spannungslos,
  man grindet ins Leere. (Details: `studigame-plan.md` §3b.)
- [x] Softlock behoben: Jutebeutel (Stufe 1) fasste 50 €, Upgrade auf Rucksack
  kostete aber 69,64 € — unerreichbar ohne BAföG. Jutebeutel fasst jetzt 80 €,
  Upgrade-Preise runden auf volle Euro. Neuer Test
  `TestContainerCapacityAffordsNextUpgrade` verhindert die Wiederkehr.
- [x] 🙋 Balancing der Sammeln-Aktion abnehmen: 40 Flaschen/h, 25 ct je Flasche,
  Behälterstufen 20/80/200/1000/5000 €, Upgrade-Kosten, Laune 5 Punkte/h
- [x] Alpine.js einbinden und die laufende Aktion live runterzählen: auf
  `/actions` der offenen Schicht einen echten Countdown im Browser statt des
  Meta-Refresh (Meta-Refresh als Fallback behalten). Dafür ist Alpine.js als
  Frontend-JS neben htmx eingebunden (vendored, kein Build-Schritt).
- [ ] **Hier anhalten und spielen.** Sammeln + Behälter + Upgrade sind bereits ein
  Spiel. Wenn sich der Kern langweilig anfühlt, retten ihn die Punkte darunter nicht.
- [x] Zweite zeitbasierte Aktion: Werkstudenten-Job (fester Stundenlohn, kein
  Glück) neben dem Flaschensammeln. Auflösung per `kind`-Dispatch erweitert
  (ein `switch` über `collect`/`job`, gemeinsamer Buchungspfad für Geld/Laune/
  Ledger/Charity)
- [x] Klausur/Studium als dritte Aktion (ECTS-Produktion) — das Skillsystem
  steht, danach schlüssig als Nächstes. Blocker: ECTS hat bisher keinen
  Abnehmer — der Konsument wäre die aufgeschobene Bildungsstufe (Skill) und
  das Stipendium; erst entscheiden, was ECTS bezahlt, sonst steht ungenutzter
  Fortschritt in der DB
- [x] Skillsystem (reduzierter MVP): drei Skills (Lerneffizienz, Auftreten,
  Konzentration), ein Weiterbildungs-Slot pro Charakter (bewusst statt zweien),
  80 % Erstattung bei Abbruch, überlineare Kostenkurve; Studienfächer
  differenzieren Kosten/Dauer (Geisteswissenschaften günstiger, Jura teurer und
  langsamer, Informatik loyal bei Konzentration). Der Skill-Bonus wird beim
  Start einer Aktion eingefroren, damit ein Level-Up mitten in der Schicht
  nichts rückwirkend aufbläht
- [x] Passives **BAföG** als Grundeinkommen: nach dem Beantragen (Dashboard-Knopf)
  trifft alle 6 h ein fester Betrag (30,00 €) in den Behälter, ohne Schicht und
  ohne Laune-Kosten. Booking, Ledger und Auszahlungs-Zeiger in einer Transaktion
  (kein Doppelzahlen beim Neuladen); ein 4-Perioden-Grenzwert begrenzt die
  Nachzahlung für vernachlässigte Accounts; Überlauf geht in den Spendentopf.
  Abrechnung in `dashboard.go` und `collect_handlers.go`, reine Regeln in
  `internal/game/bafog.go`
- [x] **Campus-Kiosk: Kaffee-Boost** (erster Teil des breiten Shops): ein Kaffee
  (2 €) auf dem Dashboard halbiert für 2 h die Weiterbildungsdauer; Käufe
  stapeln das Fenster. Mechanik „Kaffee statt Promille" aus dem Plan. Reine
  Regeln in `internal/game/shop.go`, Kauf in `coffee_store.go`, Route
  `POST /kiosk/coffee`
- [x] Breiter Shop (Ausrüstung, Instrumente, Wohnung) — erst nach weiteren
  Einkommensquellen, sonst kauft nichts zusammen
- [x] Verbrechen-Aktion mit Risiko und Ertrag (Nebenverdienst): eine Hausarbeit
  ghostwriten oder einen Spickzettel verticken, `kind "crime"` im selben
  Dispatch-Schalter wie `collect`/`job`, gemeinsamer Buchungspfad. Höherer
  Stundenlohn als der Job (15 €/h statt 8 €/h), dafür 30 % Risiko `roll ≤
  0.30` erwischt zu werden (Kohle fällt auf ein Viertel, Laune doppelt).
  Skill-neutral im MVP — Taschendiebstahl/Stressresistenz sind aufgeschoben.
  Reine Regeln in `internal/game/crime.go`, Start in `crime_store`→
  `StartCrime`, Route `POST /actions/crime`. Zahlen sind Startwerte, mit dem
  Sammeln-/Job-Balancing gemeinsam spielen und bewerten 🙋
- [x] Tabellentests für jede Formel, Integrationstests für jede Aktion
- [ ] 🙋 Balancing spielen und bewerten

## Später — Sozial & PvP

- [ ] Kampfsystem mit 36-h-Cooldown pro Ziel und Neulingsschutz
- [ ] Banden: Kasse, Ausbauten, Chat, zwei Rollen (Chef/Mitglied)
- [ ] Nachrichten, Freunde, Blockliste, Meldefunktion
- [ ] Highscore je Edition
- [ ] Referral-Link `/kleingeld/{code}` mit Klick-Deduplizierung
- [ ] 🙋 Moderationsregeln festlegen

## Vor dem Launch

- [ ] PWA: Manifest und Icons (Service Worker nur, falls danach noch Lust da ist)
- [ ] Performance- und Accessibility-Pass
- [ ] Unterstützer-Status: werbefrei, Badge, Kosmetik — kein Gameplay-Vorteil
- [ ] Zahlungsanbieter mit idempotenten Webhooks
- [ ] Spenden-Button und echte Charity-Kopplung
- [ ] 🙋 Einfache Wiki-/Hilfeseite für Spieler anlegen (Mechaniken kurz erklärt:
  Sammeln, Job, Behälter, Weiterbildung)
- [ ] 🙋 Charity-Partner auswählen und anschreiben
- [ ] 🙋 AGB, Widerruf, Steuerfragen klären
- [ ] Closed Beta an die Warteliste
- [ ] Öffentliches Changelog

---

## Rebrand: Pennergame → Studigame

Konzeptwechsel, Details und Begründung in `studigame-plan.md` (Abschnitt 3, 3a).
Der Code-Umbau ist umgesetzt und grün (`go test ./...`).

- [x] Go-Modul `pennergame` → `studigame` (go.mod, alle Imports, DB-Dateiname,
  Backup-Skript, docker-compose)
- [x] `districts`-Tabelle (`yield_factor`/`risk_factor`/`unlock_at_cents`/
  `distance_rank`) entfernt, `subjects`-Tabelle (Studienfach) eingeführt;
  `characters.district_id` → `characters.subject_id`; Charaktererstellung
  fragt das Studienfach ab statt einer Stadt
- [x] Campus-Orte (Mensa/Bibliothek/Kneipenviertel) als reines Text-Label in
  Aktionen, ohne eigene Stats-Tabelle
- [x] Editionen-Bootstrap auf eine fiktive Uni-Stadt reduziert (statt vier
  NRW-Städte); `editions`-Tabelle bleibt technisch für spätere Erweiterung
- [x] Migrationen `0003_core_gameplay.sql` direkt angepasst statt additiv
  überschrieben, `0005_open_duesseldorf.sql` gelöscht (keine echten
  Nutzerdaten)
- [ ] Skill-/Mechanik-Namen aus `studigame-plan.md` Abschnitt 3 übernehmen
  (ECTS-Punkte, Rabattcode-Jäger, Notenkonkurrenz, …), sobald die jeweilige
  Mechanik gebaut wird
- [x] `CLAUDE.md`/README-Verweise auf Bezirke/Städte final auf Studienfach/
  Edition umziehen, sobald der Code-Umbau steht

## Editionen

Eine fiktive Uni-Stadt zum Start (MVP), statt der vier NRW-Städte aus der
Pennergame-Version — siehe Rebrand-Abschnitt oben. Weitere Uni-Städte sind als
zusätzliche Editionen technisch vorbereitet, aber erst relevant, wenn eine
volle erste Stadt langweilig wird (`studigame-plan.md`, Abschnitt 7b).

- [x] Name der Uni-Stadt festgelegt: **Studistadt**

---

## Ideen-Halde

Nicht eingeplant, nicht versprochen. Hier landen Einfälle, damit sie nicht im
Kopf herumliegen. Der Standard ist „wird nicht gebaut“, bis sich jemand
beschwert oder eine Zahl es rechtfertigt.

**Gameplay**

- Saison-/Ligasystem mit Reset und kosmetischen Belohnungen
- Tägliche Kleinaufgaben („bring 20 Flaschen in Ehrenfeld“)
- Wetter, das den Ertrag beeinflusst
- Haustiere und Haustierkämpfe (Original-Feature, für den MVP zu groß)
- Instrumente mit Minispiel statt reinem Timer
- Stadtteil-Kontrolle durch Banden

**Community**

- Öffentliches Spielerprofil mit teilbarer Signatur
- Wochenrückblick per E-Mail (opt-in)
- Bandenkriege und Bündnisse — erst sinnvoll, wenn genug Banden existieren
- Feinkörnige Bandenrollen jenseits von Chef und Mitglied
- Mehrwert für die bei der Registrierung erfasste Uni-Stadt (bisher reine
  Anzeige): Stadt-Leaderboard zusätzlich zum globalen Highscore; Freunde-/
  WG-Suche nach Stadt filterbar, sobald es Freunde/Banden gibt; Flavor-Text
  je Stadt (Mensa-Insider-Jokes o. Ä.) — günstig, viel Charme. Bewusst keine
  Städte aus der Liste streichen: das Risiko, dass jemandes Stadt fehlt,
  wiegt schwerer als der Pflegeaufwand einer langen Liste
- Nostalgie-Quiz beim Onboarding, rein kosmetisch

**Technik**

- Volltextsuche über Nachrichten
- Admin-Oberfläche statt SQL-Konsole
- Highscore-Snapshot statt Live-Sortierung — erst, wenn das `ORDER BY` messbar bremst
- Metriken-Endpunkt für Retention und aktive Spieler
- Postgres — erst, wenn SQLite gemessen an eine Grenze stößt
- Web-Push — bewusst verworfen, passt nicht zum ruhigen Ton
- DaisyUI — verworfen: Phase-2-Komponenten sind bereits handgebaute Tailwind-Partials
  in `internal/web/components`. Neue Dependency nur um funktionierenden Code zu ersetzen
  wäre Rückschritt (siehe CLAUDE.md-Regeln)
