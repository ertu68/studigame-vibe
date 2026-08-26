# Roadmap

Was gebaut wird, in welcher Reihenfolge, und was bewusst noch nicht.
Detailbegründungen stehen in [pennergame-remake-plan.md](./pennergame-remake-plan.md).

Legende: 🙋 = manuelle Aufgabe (nicht per Code lösbar) · 🚧 = blockiert etwas

---

## Jetzt

Kleine Fixes und das Fundament, das später teuer nachzurüsten wäre.

- [ ] `Dockerfile` auf Go 1.27 (aktuell 1.23, Build schlägt fehl) 🚧
- [ ] CSRF + Rate-Limit auf `POST /waitlist`
- [ ] Sessions beim Passwort-Reset invalidieren
- [ ] `Register`: nur UNIQUE-Verletzung als „E-Mail vergeben“ werten
- [ ] Migrations-System (nummerierte SQL-Dateien + `schema_migrations`) 🚧 vor der ersten echten Nutzer-DB
- [ ] Server-Timeouts, Graceful Shutdown, Panic-Recovery
- [ ] `GET /healthz`, Security-Header, strukturiertes Logging
- [ ] htmx lokal ausliefern statt per CDN
- [ ] Sweeper: abgelaufene Sessions und Reset-Tokens löschen
- [ ] CI: `go vet` + `go test -race` bei jedem Push
- [ ] Backup-Skript für die SQLite-Datei
- [ ] 🙋 Name + Domain festlegen 🚧 blockiert alles Öffentliche
- [ ] 🙋 VPS bestellen, DNS setzen, `.env` befüllen
- [ ] 🙋 Mailversand-Anbieter wählen, SPF/DKIM setzen
- [ ] 🙋 Impressum + Datenschutzerklärung
- [ ] 🙋 Backup-Restore einmal testen

## Als Nächstes

- [ ] Double-Opt-In für Warteliste und Registrierung
- [ ] Echter SMTP-Versand statt Konsolen-Mailer
- [ ] Design-System: Tailwind-Tokens, Basiskomponenten, Statusleiste
- [ ] Geld-Behälter als sichtbares Fortschritts-Element im Header
- [ ] Bottom-Navigation, mobile-first
- [ ] Bestehende Seiten auf das finale Design umstellen

## Danach — Kernschleife

Der eigentliche Kern. Alles darunter hängt am Fundament aus „Jetzt“.

- [ ] `characters`, `editions`, `districts`, `ledger`, `actions`
- [ ] Reine Spiellogik in `internal/game` (keine DB, keine Uhr, testbar)
- [ ] Sammeln-Aktion mit Dauerwahl und Lazy-Tick-Auflösung
- [ ] Geld-Behälter mit Kappung, Überlauf fließt in den Charity-Pool
- [ ] **Hier anhalten und spielen.** Sammeln + Behälter sind bereits ein Spiel. Wenn sich der Kern langweilig anfühlt, retten ihn die Punkte darunter nicht.
- [ ] Skillsystem mit Warteschlange (2 Slots, 80 % Erstattung bei Abbruch)
- [ ] Shop: Behälter-Upgrades, Ausrüstung, Instrumente, Wohnung
- [ ] Verbrechen-Aktion mit Risiko und Ertrag
- [ ] Tabellentests für jede Formel, Integrationstests für jede Aktion
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
| Düsseldorf | Kontrast-Stadt: teuer, riskant, belohnt Kapital | geschlossen |

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
