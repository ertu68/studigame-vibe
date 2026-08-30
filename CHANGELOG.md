# Changelog

Alle nennenswerten Änderungen an diesem Projekt.

Format nach [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
Versionierung nach [SemVer](https://semver.org/lang/de/).

Bis zum ersten öffentlichen Launch steht alles unter `[Unreleased]`.

## [Unreleased]

### Hinzugefügt

- **Skillbaum (Weiterbildung)**: Drei Skills — Lerneffizienz (mehr Leergut),
  Auftreten (bessere Job-Bezahlung), Konzentration (schnellere
  Weiterbildungen). Ein Weiterbildungs-Slot pro Charakter, Kosten werden vorab
  abgebucht, Abbruch erstattet 80 %. Die Kostenkurve ist überlinear
  (`Basis × Stufe^1.8`), Studienfächer differenzieren Kosten und Dauer.
  Der Skill-Bonus wird beim Start einer Aktion eingefroren. Erreichbar über
  „Lernen“ in der Bottom-Navigation.
- **BAföG als passives Grundeinkommen**: nach dem Beantragen (Knopf auf dem
  Dashboard) trifft alle 6 Stunden ein fester Betrag (30,00 €) in den Behälter
  — ohne Schicht, ohne Laune-Kosten. Überlauf geht wie jedes Einkommen in den
  Spendentopf. Ein 4-Perioden-Grenzwert verhindert, dass ein vernachlässigter
  Account eine Wochen-Auszahlung anspart; die Auszahlungs-Phase bleibt an der
  ursprünglichen Kadenz verankert (Booking, Ledger und Pointer in einer
  Transaktion, sodass ein Neuladen nicht doppelt zahlt).
- **Campus-Kiosk: Kaffee-Boost**: Ein Kaffee (2,00 €) auf dem Dashboard
  halbiert für zwei Stunden die Dauer einer Weiterbildung (temporärer
  Konzentrations-Boost, Mechanik nach dem Plan „Kaffee statt Promille").
  Mehrere Käufe stapeln das Fenster; Ledger-Buchung und Boost-Pointer liegen in
  einer Transaktion. Verbrauchter Boost ist die erste Verwendung des
  `caffeine_until`-Felds.
- **Dritte Aktion: Nebenverdienst (Verbrechen)**: Eine Hausarbeit ghostwriten
  oder einen Spickzettel verticken — höherer Stundenlohn als der Job (15 €/h
  statt 8 €/h), dafür 30 % Risiko, erwischt zu werden. Ein drohender Schlag
  kürzt die Kohle auf ein Viertel und kostet doppelt Laune; sauber läuft die
  Auszahlung voll. Risiko wird erst beim Auflösen gewürfelt, Skill-neutral im
  MVP (Taschendiebstahl/Stressresistenz sind aufgeschoben).

### Geändert

- **Registrierungsfeld „Uni-Stadt" → „Heimatstadt"**: Das Feld, in dem
  Spieler ihre echte Herkunftsstadt angeben (reine Anzeige, keine
  Spielfunktion), heißt jetzt „Heimatstadt". Der alte Name kollidierte mit
  der fiktiven Spielwelt „Studistadt" (`editions`).

- **Karten-Abstand und Skill-Werte auf dem Dashboard**: Die Karten auf
  `/actions` und `/skills` haben mehr Luft (einheitlicher `space-y-8`-Abstand),
  und das Dashboard zeigt im Kartenblock „Dein Stand“ die aktuellen
  Skill-Stufen (z. B. `Lerneffizienz: 1/20`).
- **Alpine.js und Live-Countdown für laufende Aktionen**: Die Restzeit einer
  offenen Schicht zählt auf `/actions` im Browser sekündlich runter (Alpine.js,
  vendored, kein Build-Schritt) und lädt beim Ablauf automatisch neu, damit der
  Server auszahlt. Der Meta-Refresh bleibt als No-JS-Fallback, der manuelle
  Neu-laden-Knopf ist dafür entfallen.
- **Werkstudenten-Job als zweite zeitbasierte Aktion**: Neben dem
  Flaschensammeln gibt es jetzt einen Job mit festem Stundenlohn (kein Glück,
  aber verlässlich). Die Auflösung läuft über einen `kind`-Dispatch
  (`collect`/`job`) mit gemeinsamem Buchungspfad für Geld, Laune, Ledger und
  Charity — Geld fließt weiterhin in denselben Behälter. Klausur/Studium folgt
  erst mit dem Skillsystem.
- **Behälter-Upgrade als erster Geld-Sink**: Der Behälter lässt sich auf dem
  Dashboard sofort aufrüsten (ein Tap, eine Transaktion; Ledger-Debit mit
  `reason='container_upgrade'`). Überlineare Kostenkurve `2000 × Level^1.8` Cent,
  Max-Stufe 4 (Bausparvertrag), danach blendet die UI den Knopf aus. Ohne diese
  Ausgabe-Möglichkeit war der Charity-Überlauf spannungslos.
- **Rebrand Pennergame → Studigame**: Go-Modul, alle Imports und der
  DB-Dateiname heißen jetzt `studigame`; Backup-Skript, docker-compose und
  IDE-Dateien nachgezogen
- **Bezirke durch Studienfächer ersetzt**: Die `districts`-Tabelle
  (Ertrags-/Risiko-Multiplikator) ist entfernt, `subjects` (Studienfach)
  eingeführt; die Charaktererstellung fragt jetzt das Studienfach ab statt
  einer Stadt. Bis das Skillsystem steht, ist das Studienfach rein Flavor
  (siehe `studigame-plan.md` §3b)
- **Spielwelt**: statt vier paralleler NRW-Editionen startet das Spiel mit
  einer fiktiven Uni-Stadt (**„Studistadt“**); die `editions`-Tabelle
  bleibt technisch für spätere Uni-Städte
- **Landingpage**: Teaser-Text neu formuliert – beschreibt das
  Studentenleben-Spiel auf eigene Faust statt als Pennergame-Hommage
- **Behälter-Überlauf**: Was nicht mehr in den Geld-Behälter passt, geht
  automatisch in den Spendentopf der Uni-Stadt statt zu verfallen
- **Dokumentation**: Plan, Roadmap und README auf den aktuellen Stand gebracht.
  Der Entwicklungsplan ist final aufs MVP geschnitten (siehe §3b) – klar gestaffelte
  Kernschleife: Sammeln + Behälter stehen, Behälter-Upgrade als erster Geld-Sink
  kommt als Nächstes, dann Balancing, dann scharf spielen.
  DaisyUI als bewusst verworfen gekennzeichnet, tote Plan-Verweise entfernt.
  Der Plan ist **implementierungsfertig**: alle MVP-Entscheidungen sind getroffen
  (Behälter-Upgrade als Geld-Sink: sofortiger Kauf, überlineare Kostenkurve
  `Basiswert × Level^1.8` mit Basiswert 2000 ct, Max-Stufe 4; Uni-Stadt
  „Studistadt“ und die 4 Studienfächer final; Kosten und Formeln in §3b).
  Offen bleiben nur Pre-Launch-Ops (Domain, VPS, Mailversand, Rechtliches) –
  die erledigt der Entwickler später beim Deploy

### Hinzugefügt

- Landingpage mit Teaser-Text und Warteliste (`GET /`, `POST /waitlist`),
  Absenden per htmx ohne Reload
- Registrierung, Login, Logout mit bcrypt-Passworthashes, Passwort-Reset und
  Double-Opt-In per E-Mail-Verifikation
- Session-Auth über httpOnly-Cookies mit serverseitiger Session-Tabelle
- CSRF-Schutz (Double-Submit-Cookie) auf allen POST-Routen; IP-Rate-Limiting
  auf Login, Registrierung und Passwort-Reset
- SQLite-Anbindung (WAL, Foreign Keys) mit eingebettetem Schema und
  Migrations-System (`internal/db/migrations`)
- Einheiten und Aktionen im Schema: `editions`, `subjects`, `characters`,
  `ledger`, `actions`
- Charaktererstellung unter `/character/new`: Studienfach wählen, Name setzen,
  10 € Startkapital; ein Charakter pro Edition, Name je Edition eindeutig
- Dashboard unter `/` mit Kasse, Laune und Studienfach des Charakters
- `internal/game`: reine Spielregeln für Sammelertrag, Behältergrößen und
  Charity-Überlauf, ohne Datenbank und ohne eigene Uhr
- Sammeln-Aktion unter `/actions`: Schichtlänge wählen (15 min bis 8 h), das
  Ergebnis wird erst beim Zurückkommen ausgewürfelt; eine Schicht pro Charakter,
  keine Doppelauszahlung
- Laufende Schicht zeigt Restzeit und Rückkehrzeit; die Seite lädt sich per
  Meta-Refresh genau einmal neu, wenn die Schicht fällig ist
- `Mailer`-Interface mit Konsolen- und SMTP-Implementierung (SMTP via Env-Vars)
- Deploy-Setup: Dockerfile, docker-compose, Caddy als TLS-Reverse-Proxy
- Unit-Tests für Auth, Rate-Limiter und die Spielregeln
- Structured Logging mit `log/slog`, Panik-Handling-Middleware,
  Security-Header, lokales htmx-Serving (CSP-kompatibel), `/healthz`,
  Session-Sweeper, Backup-Skript, CI/CD-Pipeline (`go vet`, `go test -race`)
- **Design-System**: Tailwind mit benutzerdefinierten Farben, mobile-first
  Layout mit Bottom-Navigation, Sticky-Statusleiste, wiederverwendbare
  Komponenten (button, card, form_field, flash, progress, statusbar, bottom_nav)

### Behoben

- **Live-Countdown brach unter CSP**: Die Content-Security-Policy
  (`script-src 'self'`) blockierte Alpine.js – es wertet `x-data`/`x-show`
  zur Laufzeit aus und braucht `'unsafe-eval'`; der Registrierungs-Code lag
  als Inline-Script. Beides gelöst: Komponente nach `static/js/actions.js`
  ausgelagert (`'self'`-Datei, keine Inline-Skripte) und `script-src` um
  `'unsafe-eval'` ergänzt.

- Dockerfile auf Go 1.27 (D1)
- CSRF + Rate-Limit auf `POST /waitlist` (D3)
- Sessions beim Passwort-Reset invalidieren (D4)
- `Register` nur UNIQUE-Verletzung als „E-Mail vergeben“ werten (D5)
- Migrations-System (D11)
- Tailwind-Build lud `tailwind.config.js` nie (fehlende `@config`-Direktive),
  dadurch waren Custom-Farben und Komponentenklassen nicht im CSS – Seiten
  liefen unstyled aus
- Fehlender Logout-Button in der UI; doppelt gerenderte Flash-Meldung auf
  Login/Register/Reset
- Docker-Build schlug fehl, weil das mitgelieferte `tailwindcss`-Binary nur auf
  macOS lief; Build und `make css` laden jetzt die passende Version selbst
- Überschriften/Fließtext nutzten Schrift-Klassen, die es im CSS nicht gab
- Nach dem Passwort-Reset führt ein normaler Link zum Login statt Inline-JS;
  die Content-Security-Policy erlaubt keine Inline-Skripte mehr
