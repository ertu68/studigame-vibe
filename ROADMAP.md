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

- [x] `characters`, `editions`, `districts`, `ledger`, `actions`
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

## Editionen

Vier NRW-Städte als getrennte Spielwelten. Alle vier sind ab Tag 1 in Datenbank
und Code, freigeschaltet wird per Flag.

| Stadt | Twist | Launch |
| --- | --- | --- |
| Köln | Event-Stadt: Feiertage bringen Spitzen | offen |
| Dortmund | Spieltag-Stadt: belohnt Timing | offen |
| Duisburg | Schicht-Stadt: konstant, belohnt Ausdauer | geschlossen |
| Düsseldorf | Kontrast-Stadt: teuer, riskant, belohnt Kapital | offen |

Zwei Städte zum Start, weil eine geteilte Spielerbasis sich toter anfühlt als
eine volle Stadt. Öffnen ist ein `UPDATE`, kein Release.

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
- Nostalgie-Quiz beim Onboarding, rein kosmetisch

**Technik**

- Volltextsuche über Nachrichten
- Admin-Oberfläche statt SQL-Konsole
- Highscore-Snapshot statt Live-Sortierung — erst, wenn das `ORDER BY` messbar bremst
- Metriken-Endpunkt für Retention und aktive Spieler
- Postgres — erst, wenn SQLite gemessen an eine Grenze stößt
- Web-Push — bewusst verworfen, passt nicht zum ruhigen Ton
