# Roadmap

Was gebaut wird, in welcher Reihenfolge, und was bewusst noch nicht.
Detailbegründungen stehen in [pennergame-remake-plan.md](./pennergame-remake-plan.md).

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
- [ ] 🙋 Name + Domain festlegen 🚧 blockiert alles Öffentliche
- [ ] 🙋 VPS bestellen, DNS setzen, `.env` befüllen
- [ ] 🙋 Mailversand-Anbieter wählen, SPF/DKIM setzen
- [ ] 🙋 Impressum + Datenschutzerklärung
- [ ] 🙋 Backup-Restore einmal testen
- [ ] uuid für öffentliche IDs (z. B. Charaktere) statt fortlaufender Integer
  erwägen — verhindert Enumeration. `github.com/google/uuid` ist bereits im
  `go.sum` (Transitiv-Dependency von `modernc.org/sqlite`), aber unbenutzt;
  müsste von `// indirect` zu direkter Dependency werden
- [x] Tailwind vs. handgeschriebenes CSS entschieden: Tailwind bleibt

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

## Danach — Kernschleife

Der eigentliche Kern. Alles darunter hängt am Fundament aus „Jetzt“.

- [x] `characters`, `editions`, `districts`, `ledger`, `actions` (Schema wird
  im Studigame-Umbau angepasst: `districts` → `subjects`, siehe Rebrand-Abschnitt)
- [x] Reine Spiellogik in `internal/game` (keine DB, keine Uhr, testbar)
- [x] Charaktererstellung: Stadt wählen, Name setzen, Startkapital über das Ledger
- [x] Sammeln-Aktion mit Dauerwahl und Lazy-Tick-Auflösung
- [x] Geld-Behälter mit Kappung, Überlauf fließt in den Charity-Pool
- [ ] **Hier anhalten und spielen.** Sammeln + Behälter sind bereits ein Spiel. Wenn sich der Kern langweilig anfühlt, retten ihn die Punkte darunter nicht.
- [ ] 🙋 Balancing der Sammeln-Aktion abnehmen: 40 Flaschen/h, 25 ct je Flasche,
  Behälterstufen 20/50/200/1000/5000 €, Laune kostet 5 Punkte je Stunde
- [ ] Skillsystem mit Warteschlange (2 Slots, 80 % Erstattung bei Abbruch)
- [ ] Shop: Behälter-Upgrades, Ausrüstung, Instrumente, Wohnung
- [ ] Verbrechen-Aktion mit Risiko und Ertrag
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
- [ ] 🙋 Charity-Partner auswählen und anschreiben
- [ ] 🙋 AGB, Widerruf, Steuerfragen klären
- [ ] Closed Beta an die Warteliste
- [ ] Öffentliches Changelog

---

## Rebrand: Pennergame → Studigame

Konzeptwechsel, Details und Begründung in `studigame-plan.md` (Abschnitt 3, 3a).
Noch nicht umgesetzt — eigener Umsetzungsplan, bevor Code angefasst wird.

- [ ] Go-Modul `pennergame` → `studigame` (go.mod, alle Imports, Docker-Image-
  und Binary-Name) — rein mechanisch, aber repo-weit
- [ ] `districts`-Tabelle (`yield_factor`/`risk_factor`/`unlock_at_cents`/
  `distance_rank`) entfernen, `subjects`-Tabelle (Studienfach) einführen;
  `characters.district_id` → `characters.subject_id`
- [ ] Campus-Orte (Mensa/Bibliothek/Kneipenviertel) als reines Text-Label in
  Aktionen, ohne eigene Stats-Tabelle
- [ ] Editionen-Bootstrap auf eine fiktive Uni-Stadt reduzieren (statt vier
  NRW-Städte); `editions`-Tabelle bleibt technisch für spätere Erweiterung
- [ ] Migrationen `0003_core_gameplay.sql`/`0005_open_duesseldorf.sql` direkt
  anpassen statt additiv überschreiben (noch keine echten Nutzerdaten)
- [ ] Skill-/Mechanik-Namen aus `studigame-plan.md` Abschnitt 3 übernehmen
  (ECTS-Punkte, Rabattcode-Jäger, Notenkonkurrenz, …), sobald die jeweilige
  Mechanik gebaut wird
- [ ] `CLAUDE.md`/README-Verweise auf Bezirke/Städte final auf Studienfach/
  Edition umziehen, sobald der Code-Umbau steht

## Editionen

Eine fiktive Uni-Stadt zum Start (MVP), statt der vier NRW-Städte aus der
Pennergame-Version — siehe Rebrand-Abschnitt oben. Weitere Uni-Städte sind als
zusätzliche Editionen technisch vorbereitet, aber erst relevant, wenn eine
volle erste Stadt langweilig wird (`studigame-plan.md`, Abschnitt 7b).

- [ ] 🙋 Name der Uni-Stadt festlegen

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
