# Entwicklungsplan: Pennergame-Nachbau (Arbeitstitel)

*Stand: 26.08.2026 – lebendes Dokument, wird während der Entwicklung laufend angepasst.*

> **Legende für Aufgaben**
>
> - `[ ]` / `[x]` – Code-Arbeit (mache ich)
> - 🙋 **DU** – manuelle Aktion, die nur du erledigen kannst (Account anlegen, Geld ausgeben, Entscheidung treffen, Text schreiben)
> - 🚧 **BLOCKER** – blockiert echte Weiterarbeit
> - ⚠️ **RISIKO** – nicht blockierend, aber vor Launch zu erledigen

---

## 0. Status-Check (26.08.2026)

### Ist-Zustand im Repo

```
cmd/server/main.go            Routing + Wiring          55 LoC
internal/config/              Env-Konfiguration         35
internal/db/                  SQLite Open + schema.sql  57
internal/auth/                Session, CSRF, Login,     429
                              Register, Reset, Mailer
internal/httpx/               IP-Rate-Limiter           103
internal/handlers/            HTTP-Handler              237
internal/web/                 html/template + embed     59
internal/web/templates/       7 Seiten + 1 Partial
static/, deploy/, Dockerfile, docker-compose.yml, Makefile
≈ 1.100 LoC · `go build ./...` ✅ · `go test ./...` ✅ · `go vet` ✅
```

### Phase 0 – Landingpage/Warteliste: **technisch fertig, betrieblich nicht live**

| Punkt | Status |
| --- | --- |
| 1-Seiter mit Teaser-Text | [x] `templates/landing.html`, Formulierung wie in Abschnitt 3 |
| E-Mail-Feld, htmx ohne Reload | [x] `POST /waitlist` → `partials/waitlist_success.html` |
| Waitlist-Tabelle | [x] `waitlist` mit `UNIQUE(email)`, `INSERT OR IGNORE` |
| Domain + Name final | 🙋 **DU** – offen, `SITE_NAME=TODO-Name` |
| Auf Kanälen posten | 🙋 **DU** – offen |
| Seite tatsächlich erreichbar | 🙋 **DU** – nie deployt, läuft nur lokal |

### Phase 1 – Grundgerüst: **~80 % fertig**

| Punkt | Status |
| --- | --- |
| Repo, Go-Module, Ordnerstruktur | [x] – aber **0 Commits**, alles untracked 🙋 |
| SQLite-Schema `users`, `sessions` | [x] + `password_resets`, `waitlist` |
| Registrierung / Login / Logout | [x] |
| **Bestätigungsmail (Double-Opt-In)** | [ ] **fehlt komplett** – im Plan gefordert |
| Passwort-Reset | [x] Token gehasht, TTL 60 min, keine User-Enumeration |
| bcrypt, httpOnly-Cookie, CSRF, Rate-Limit | [x] |
| Basis-Layout + Navigation | [x] |
| Deploy-Pipeline einmal durchgespielt | [ ] 🚧 **BLOCKER, siehe unten** |

### Gefundene Defekte (aus dem Code-Review)

| # | Fund | Schwere | Wo |
| --- | --- | --- | --- |
| D1 | `Dockerfile` nutzt `golang:1.23-alpine`, `go.mod` verlangt `go 1.27.0` → **Docker-Build schlägt fehl** | 🚧 Blocker | `Dockerfile:1` |
| D2 | `docker-compose.yml` hat `env_file: .env`, die Datei existiert nicht → `compose up` bricht ab | 🚧 Blocker | 🙋 `cp .env.example .env` |
| D3 | `POST /waitlist` ohne CSRF **und** ohne Rate-Limit → offener Spam-Endpunkt | ⚠️ hoch | `main.go:37` |
| D4 | Nach Passwort-Reset bleiben alle alten Sessions gültig | ⚠️ hoch (Security) | `auth.ResetPassword` |
| D5 | `auth.Register` mappt **jeden** DB-Fehler auf `ErrEmailTaken` → echte Fehler unsichtbar | ⚠️ mittel | `auth.go:27` |
| D6 | `http.ListenAndServe` ohne Timeouts (Slowloris) und ohne Graceful Shutdown | ⚠️ mittel | `main.go:54` |
| D7 | Abgelaufene `sessions`/`password_resets` werden nie gelöscht → Tabellen wachsen | ⚠️ mittel | – |
| D8 | htmx per CDN (`unpkg.com`) → externe Abhängigkeit zur Laufzeit, keine SRI, kein CSP möglich | ⚠️ mittel | `layout.html:9` |
| D9 | Kein `/healthz`, keine Security-Header, kein strukturiertes Logging (`log` statt `log/slog`) | ⚠️ niedrig | – |
| D10 | `static/` enthält `.DS_Store` und wird ungefiltert per `FileServer` ausgeliefert | ⚠️ niedrig | `static/` |
| D11 | Kein Migrations-Mechanismus – `schema.sql` kann nur `CREATE IF NOT EXISTS`, keine `ALTER`-Historie | 🚧 Blocker **ab Phase 3** | `internal/db` |
| D12 | Kein CI (kein `go vet`/`test` bei Push) | ⚠️ niedrig | – |

### Antwort auf die Frage: „ist irgendwas blockiert?"

**Feature-Arbeit: nein.** Phase 2 und Phase 3 könnten sofort starten.
**Deploy: ja, aber nur trivial** – D1 (eine Zeile im Dockerfile) und D2 (`cp .env.example .env`).
**Ab Phase 3: D11** – ohne Migrationen wird jedes Schema-Update zum Handbetrieb auf der Prod-DB. Muss vor der ersten echten Nutzer-DB stehen.

Inhaltlich blockiert dich genau **eine** Sache, und die kann ich nicht für dich lösen:

> 🙋 **DU: Name + Domain festlegen.** Daran hängen: Landingpage-Text, `SITE_NAME`, `SITE_DOMAIN`, TLS-Zertifikat, Mail-Absenderdomain (SPF/DKIM), Discord-Servername, Stripe-Produktname. Solange das `TODO-Name` bleibt, kann Phase 0 nicht wirklich live gehen.

### 🙋 Deine Aufgabenliste – jetzt sofort

| # | Aufgabe | Warum jetzt | Kosten |
| --- | --- | --- | --- |
| A1 | **Name entscheiden** (Kartonkönig / Bahnhofsbaron / Pfandpirat / NullkommaNichts / eigener) | blockiert alles Öffentliche | 0 € |
| A2 | **Domain registrieren** (`.de` über INWX/Netcup) | Voraussetzung TLS + Mail | ~10 €/Jahr |
| A3 | **VPS bestellen** (Hetzner CX22, Falkenstein) + DNS A-Record auf die IP | Deploy-Durchstich | ~4 €/Monat |
| A4 | `cp .env.example .env`, Werte eintragen | D2 | 0 € |
| A5 | **Git-Remote anlegen** (GitHub, privat) – Repo hat 0 Commits | Backup + CI | 0 € |
| A6 | **Mailversand-Anbieter** wählen (Empfehlung: Postmark ~15 $/Mo oder Brevo Free-Tier 300/Tag) | Double-Opt-In + Passwort-Reset brauchen echten Versand | 0–15 €/Mo |
| A7 | Markenrecherche „Penner…"-Wortmarke bei DPMA (kostenloses Online-Register), optional Kurz-Anwaltstermin | Namensrisiko vor dem Bekanntmachen ausschließen | 0 € / ~150 € |
| A8 | **Impressum + Datenschutzerklärung** (Pflicht in DE, ab der ersten öffentlichen Seite mit E-Mail-Formular) | rechtlich verpflichtend, DSGVO | 0 € (Generator) |

*A1–A5 reichen für einen Live-Deploy der Landingpage. A6 und A8 sind Pflicht, bevor du die Seite öffentlich bewirbst.*

---

## 1. Funktionsanalyse: Das Original

Die offizielle Anleitung (Hamburg-, Berlin-, Atlantis-Edition) und ein ausführlicher Fan-Guide von 2009 sind noch über den Suchindex erschlossen. Dadurch lässt sich das Spiel fast vollständig rekonstruieren.

### 1.1 Grundgerüst

- Start: 10 € Startkapital, Ziel: vom Pennerleben zum "Schlossbesitzer"
- Mehrere parallele "Editionen" pro Stadt (Hamburg, Berlin, Köln, München, Mallorca, dazu ein Fun-Server "Atlantis/Vatikan")
- Finanzierung nur über Spenden + "Ehrenmitgliedschaft" – **wichtig für deine Monetarisierungsfrage: Premium gab laut eigener Aussage des Originals keine echten Gameplay-Vorteile**, nur Kosmetik/Komfort (eigenes Haustier-Design, größeres Bandenavatar, eigene E-Mail-Adresse, kein Werbebanner). Das validiert die Empfehlung von letztem Mal ziemlich direkt.

### 1.2 Die Kernschleife

1. **Ressourcen sammeln** – Pfandflaschen sammeln (Dauer wählbar 10 Min–12 h; kürzer = effizienter pro Zeiteinheit, aber absolute Menge steigt mit Dauer; Ertrag hängt von Stadtteil, Geschicklichkeit und tagesaktuellem Pfandpreis ab)
2. **Verkaufen/Betteln** – Flaschen verkaufen, dazu passives Bettel-/Spenden-Einkommen
3. **Investieren** – in Skills, Ausrüstung, Wohnung
4. **Risiko eingehen** – Verbrechen begehen (mehr Ertrag, aber Erwischt-Risiko) oder andere Spieler angreifen
5. **Aufsteigen** – bessere Stadtteile, Ausrüstung, Bande, Highscore-Rang

Ein Geld-"Behälter" begrenzt, wie viel du auf einmal tragen kannst (Trinkbecher → Tüte → Beutel → Einkaufswagen → Container). Wer sein Limit nicht rechtzeitig upgradet, verliert überschüssiges Geld einfach – kleine, aber richtig clevere Mechanik, zwingt zum Reinvestieren und passt perfekt zum Thema.

### 1.3 Skillsystem ("Weiterbildung")

Kosten = Basiswert × nächstes Level. Abbruch erstattet 80 %. Zwei Skills gleichzeitig in der Warteschleife möglich.

| Skill | Effekt |
| --- | --- |
| Angriff (ATT) | Kampfstärke |
| Verteidigung (DEF) | weniger Geldverlust bei Niederlage |
| Geschicklichkeit | mehr Flaschen pro Suche; ab Stufe 20 siehst du eingehende Angriffe |
| Sprechen | Voraussetzung für die meisten anderen Skills |
| Bildungsstufe | schaltet neue Bettelplätze frei |
| Musik | schaltet Instrumente frei (passives Einkommen alle 6 h) |
| Sozialkontakte | schaltet Haustiere/Wohnungen frei |
| Konzentration | beschleunigt Weiterbildungen (4 Stufen, mit Trade-offs ggü. anderen Aktionen) |
| Taschendiebstahl | erhöht alle Einnahmen leicht |

### 1.4 Soziales & PvP

- **Kampf**: gleicher Gegner nur alle 36 h angreifbar, Kampfdauer abhängig von der "Entfernung" der Wohnbezirke, Kampf-Log mit Gewinn/Verlust-Historie
- **Haustierkämpfe**: separates Wett-System mit Einsatz und Stances (Aggressiv/Defensiv/Neutral) – nebenbei der einzige Weg, Geld direkt an andere Spieler zu senden
- **Banden**: Name+Passwort, Bandeneigentum (Waffenkammer/Bandenhaus/Training/Konto, je 5 Stufen mit %-Boni für alle Mitglieder), Bandenchat, Bündnisse/Kriege, eigener Bandenhighscore
- **Freunde/Blockliste, Nachrichten, Forensignaturen** (jeder Klick auf die Signatur triggert eine Spende!)
- **Highscore** für Spieler und Banden getrennt

### 1.5 Der eingebaute Wachstumsmotor

Hochrelevant für deine Community-Frage: Das Original hatte einen **Referral-Mechanismus direkt in der Spielfantasie eingebaut**. Jeder Spieler bekommt einen persönlichen Link (`.../change_please/{RefID}/` – "haste mal Kleingeld?"), den man in Forensignaturen oder der eigenen Profil-Homepage postet. Klickt jemand drauf, gibt es eine In-Game-Spende. Virales Wachstum, direkt an die Kernfantasie (Betteln) gekoppelt, ohne sich wie ein aufgesetzter "Invite 3 friends"-Hack anzufühlen.

### 1.6 Technischer Befund

- Backend laut einer alten Stellenanzeige von Farbflut Entertainment: **Python + MySQL**
- Die Domain läuft technisch noch, Neuregistrierung ist aber kaputt, keine Weiterentwicklung mehr
- Farbflut Entertainment GmbH (Hamburg) ist laut Handelsregister weiterhin wirtschaftsaktiv – der Markenname "verwaist" also nicht, deshalb weiterhin: eigenen Namen verwenden

---

## 2. Kuration: Übernehmen, ändern, weglassen

| Element | Entscheidung | Warum |
| --- | --- | --- |
| Grundschleife (sammeln → verkaufen → investieren → Risiko → aufsteigen) | **Übernehmen** | Zeitlos, funktioniert einfach |
| Geld-Behälter mit Kappung | **Übernehmen, leicht verändert** | Statt "Geld weg für immer" bei Überlauf: automatische Spende an eine echte Obdachlosenhilfe. Gleiche Spielspannung, fühlt sich aber gut statt frustrierend an |
| Promille-System | **Optional / abschwächen** | Charmant und thematisch stimmig, aber ggf. kein Spiel, das Alkoholkonsum als Statuswert incentiviert. Alternative: Mechanik behalten, aber als reinen "Laune/Glück"-Wert umlabeln statt explizit Alkohol |
| Sauberkeit sinkt durchs Sammeln, erhöht Spenden | **Übernehmen** | Nette Nebenressource, unproblematisch |
| Referral-Link als "Spendenlink" | **Übernehmen & ausbauen** | Genialer eingebauter Wachstumsmotor, siehe 1.5 |
| Ehrenmitgliedschaft ohne Gameplay-Vorteil | **Übernehmen** | Deckt sich mit der Monetarisierungs-Empfehlung von letztem Mal – hat sich schon damals bewährt |
| Reale Städte/Stadtteile (Hamburg, Berlin, St. Pauli, …) | **Ändern → vier NRW-Städte** | Andere Region als das Original (kein Hamburg/Berlin), aber echte Orte statt Fantasie: Ruhrgebiet + Rheinland tragen die Pfand-/Malocher-Fantasie von selbst und erzeugen echte Lokalrivalität. Details in Abschnitt 3.1 |
| Shock-Humor bei manchen Item-/Haustiernamen | **Weglassen** | Harter, dated 2000er-Forenhumor, der schon damals grenzwertig war |
| "Dein Leben beenden" als Label für Account löschen | **Ändern** | Klingt nach Selbstverletzung, auch wenn nur der Charakter gemeint ist – einfach "Account löschen" nennen |
| 36h-Cooldown pro Angriffsziel | **Übernehmen, Werte tunen** | Guter Anti-Griefing-Mechanismus |
| Kein direkter Geldtransfer (nur Bandenkasse/Haustierkämpfe) | **Übernehmen** | Verhindert simples Farming/RMT |
| Textbasiert, schlichtes UI | **Übernehmen, modernisieren** | Genau dein Wunsch: gleicher Minimalismus, aber mobile-first & schnell |

---

## 3. Konzept fürs Remake

- **Ton**: gleiche trocken-schwarzhumorige "vom Tellerwäscher"-Satire, ohne den ranzigen 2000er-Forenhumor
- **Setting**: vier echte NRW-Städte als getrennte Editionen (siehe 3.1) statt Hamburg/Berlin

### 3.1 Setting: vier NRW-Editionen

Statt einer fiktiven Stadt vier echte NRW-Städte als **getrennte Spielwelten** („Editionen“, exakt wie im Original). Jede Edition hat eigene Spieler, eigene Banden, eigenen Highscore. Kein Kampf und kein Handel über Editionsgrenzen.

| Edition | Charakter | Startbezirke (Beispiele) | Mechanischer Twist |
| --- | --- | --- | --- |
| **Köln** | Rheinländisch, Karneval, Touristen, Kioskkultur | Innenstadt, Ehrenfeld, Nippes, Kalk, Chorweiler | **Event-Stadt**: an Karnevals-/Feiertagen deutlich mehr Flaschen, dafür höhere Konkurrenz und mehr Kontrollen |
| **Dortmund** | Ruhrpott, Fußball, Bierkultur | City, Nordstadt, Kreuzviertel, Hörde, Hombruch | **Spieltag-Stadt**: an Spieltagen Ertragsspitze rund ums Stadion, ansonsten ruhig — belohnt Timing |
| **Duisburg** | Hafen, Industrie, Schichtbetrieb | Innenhafen, Ruhrort, Hochfeld, Meiderich, Homberg | **Schicht-Stadt**: gleichmäßiger, planbarer Ertrag rund um die Uhr, keine Spitzen — belohnt Ausdauer |
| **Düsseldorf** | Landeshauptstadt, Kö-Schickeria, Altstadt-Kneipenmeile | Altstadt, Friedrichstadt, Oberbilk, Bilk, Garath | **Kontrast-Stadt**: höchster Ertrag pro Flasche (teure Gegend, viel Altstadt-Leergut), aber die härtesten Freischaltschwellen und das höchste Risiko in den Nobelbezirken — belohnt Kapital statt Zeit |

**Der Rhein-Ruhr-Bogen:** die vier decken bewusst unterschiedliche Spielertypen ab — Düsseldorf (Kapital) und Duisburg (Ausdauer) sind mechanisch gegenpolig, obwohl sie 25 km auseinanderliegen. Dazu kommt die vorhandene reale Rivalität Köln–Düsseldorf, die sich als Bandenkrieg-Fantasie von selbst erzählt.

**Warum vier und nicht eine:** Editionen kosten mechanisch fast nichts (eine Spalte `edition_id`, siehe 6.4), geben aber vier Startpunkte für Community-Bildung, echte Lokalrivalität („Köln gegen Düsseldorf“, „Dortmund gegen Duisburg“) und einen natürlichen Grund für Neustarts, ohne Fortschritt zu löschen. Der Twist pro Stadt ist **ein einziger Multiplikator** in der Ertragsformel, kein Sonder-Code je Stadt.

> ⚠️ **Aber: Editionen teilen die Spielerbasis.** Vier halbleere Städte fühlen sich toter an als eine volle — PvP und Banden brauchen Mitspieler. Deshalb: **zum Launch nur Köln und Dortmund offen** (`editions.active = 0` für die anderen), Duisburg und Düsseldorf werden freigeschaltet, sobald die ersten beiden tragen. Alle vier sind ab Tag 1 in der DB und im Code — das Freischalten ist ein `UPDATE`, kein Release.

**Warum NRW:** das Original war Hamburg/Berlin-zentriert — eine andere Region schafft Abstand, ohne die Fantasie zu verlieren. Pfandsammeln, Industriebrache und Kioskkultur sind im Ruhrgebiet stärker verankert als irgendwo sonst.

**Bezirksnamen sind echt, Beschreibungen sind es nicht.** Bezirke unterscheiden sich rein über Zahlen (Ertragsfaktor, Freischaltschwelle, Distanz für Kampfdauer). Es gibt **keine** Texte, die einen realen Stadtteil oder seine Bewohner abwerten.

> ⚠️ **Sensibilität:** Manche dieser Stadtteile haben in der Presse einen Ruf, der real existierende Menschen betrifft. Ein Bezirk darf spielmechanisch „schwierig“ sein (höheres Risiko, höherer Ertrag), aber nie textlich herabgesetzt werden. Konkret: kein Bezirk bekommt eine Beschreibung, die man nicht der lokalen Stadtverwaltung zeigen könnte. Bei Zweifel: Zahl statt Prosa.

🙋 **DU:** Die vier Städte final bestätigen — und welche zwei zum Launch offen sind (Vorschlag: Köln + Dortmund). Weitere Kandidaten für später: Essen, Bochum, Gelsenkirchen, Wuppertal.

### 3.2 Weiteres

- **Name**: noch offen. "Penner" allein ist aller Wahrscheinlichkeit nach nicht eigenständig markenrechtlich geschützt (nur "Pennergame"/"Pennershop" als Ganzes) – lohnt sich aber, beim späteren Kurz-Anwaltstermin zu verifizieren, falls du es doch verwenden willst. Ein paar Ideen zum Weiterspinnen, nichts in Stein gemeißelt: *Kartonkönig*, *Bahnhofsbaron*, *Pfandpirat*, *NullkommaNichts*.
- **Wartungs-/Landingpage**: *"Bald verfügbar: [Name] – inspiriert vom Kult-Browsergame Pennergame (2007), aber kein offizielles Nachfolgeprojekt und ohne Verbindung zu Farbflut Entertainment."* Diese Formulierung referenziert das Original ehrlich (rechtlich unproblematische sogenannte nominative Nutzung), macht aber unmissverständlich klar, dass es keine offizielle Fortsetzung ist – nimmt Verwechslungsrisiko raus. Dazu ein simples E-Mail-Waitlist-Formular.

---

## 4. Tech-Stack (final)

| Bereich | Wahl | Warum |
| --- | --- | --- |
| Backend | Go, `net/http` (1.22+) oder `chi` | passt zu einem textbasierten, formularbasierten Spiel |
| Templating | `templ` oder `html/template` | typsicher, kompiliert, kein JS-Build nötig |
| Interaktivität | htmx | Server-rendered Partial-Updates statt SPA-Overhead |
| Styling | Tailwind (Standalone-CLI, kein Node-Build) | genau dein Vorschlag – Primärfarben & Co. als Tokens, siehe Abschnitt 5 |
| DB | SQLite (WAL-Modus) zum Start | 1 Datei, kein DB-Service zu pflegen; Migration auf Postgres nur falls wirklich nötig |
| Auth | **Sessions statt JWT** (Begründung unten) | einfacher & sicherer für diese Architektur |
| Game-Ticks | goroutines + `time.Ticker` | kein Queue-System nötig |
| Hosting | 1 VPS (z. B. Hetzner) + Caddy (Auto-HTTPS) | kein Kubernetes, auch wenn's aus dem Job vertraut ist |
| Payments | Stripe oder Paddle (Merchant of Record, übernimmt EU-USt.) | |

**Zu deiner JWT-Idee, kurz und ehrlich:** Für eine klassische, server-gerenderte htmx-App ist JWT die unnötig komplizierte Wahl. JWTs lohnen sich, wenn du eine zustandslose API für mehrere Clients brauchst (z. B. später eine native Mobile-App). Für "eine Web-App, ein Server" ist eine simple Server-Session (bcrypt-Hash + httpOnly-Secure-Cookie + Session-Tabelle in deiner DB) einfacher, sicherer (kein Token-Diebstahl über localStorage, kein Revoke-Problem vor Ablauf) und schneller gebaut. Brauchst du später eine echte externe API, kannst du JWT gezielt für diesen Teil nachrüsten. Zwei Dinge dabei nicht vergessen: CSRF-Tokens auf jedem htmx-POST-Formular, und Rate-Limiting auf Login/Registrierung.

---

## 5. Corporate Design – Tokens

Ausgangspunkt: Straße, Improvisation, schwarzer Humor – kein generischer SaaS-Look.

**Farben**

- Asphalt (dunkler Hintergrund): `#1C1B1A`
- Pfandglas-Grün (Primär/Akzent): `#7A9B57`
- Verkehrsorange (Sekundär/CTA): `#D9711F`
- Münzgold (Währung, Premium, Erfolge): `#C9A24B`
- Pappe/Papier (helle Flächen, Karten): `#E8E1D3`
- Backstein-Rot (Warnung, Angriff): `#A83E32`

**Typografie**

- Display/Headlines: kräftige, leicht kondensierte Groteske (z. B. Archivo Black oder Barlow Condensed Bold) – wirkt wie ein Pappschild, ohne kitschig zu sein
- Fließtext: gut lesbare humanistische Sans (z. B. Inter oder Source Sans)
- Zahlen/Stats (Geld, ATT/DEF): Monospace (z. B. JetBrains Mono) – gibt Werten einen "Ticker"-Charakter, klar abgesetzt vom Fließtext

**Layout**

- Mobile-first, eine Spalte, Bottom-Nav mit den Hauptbereichen (Übersicht/Aktionen/Stadt/Bande/Kampf) statt Top-Nav – spiegelt die alte Navigationsstruktur, nur daumenfreundlich
- Sticky Status-Leiste oben (Geld, Laune, laufende Cooldowns) – Konzept aus dem alten "Statusfenster" übernommen

**Signature-Element**
Der Geld-Behälter (Becher → Tüte → Beutel → Einkaufswagen → Container) als durchgehend sichtbares Fortschritts-Icon im Header, das sich sichtbar füllt und bei Upgrade die Form wechselt. Erzeugt sofort Spannung ("gleich läuft's über") und kommt direkt aus der Spielidee statt aus einem Template.

---

## 6. Architektur

### 6.1 Leitprinzipien

1. **Server ist die einzige Wahrheit.** Der Client rendert nur HTML. Kein Spielzustand im Browser, kein JS, das rechnet. Alles, was Geld bewegt, passiert in einer DB-Transaktion auf dem Server.
2. **Lazy Tick statt Cron.** Eine laufende Aktion wird als Zeile mit `started_at` / `ends_at` gespeichert und **beim nächsten Seitenaufruf aufgelöst**, nicht von einem Timer. Das braucht keine Job-Queue, keinen Scheduler und ist nach einem Serverneustart automatisch korrekt. Nur das, was wirklich *von selbst* passieren muss (passives Einkommen, Aufräumen), läuft in **einer** Sweeper-Goroutine.
3. **Reine Spiellogik, getrennt von SQL.** `internal/game` enthält Funktionen ohne DB und ohne `time.Now()` – Uhr und Zufall werden hineingereicht. Dadurch ist die gesamte Wirtschaft ohne Datenbank testbar, und Balancing-Änderungen sind Tabellenänderungen statt Refactorings.
4. **Eine Transaktion pro Spieleraktion.** Jede Aktion ist genau eine Funktion `Do<Aktion>(tx, userID, params) (Result, error)`. Kein halb-committeter Zustand, keine Race-Bedingung zwischen zwei Tabs.
5. **Geld nur über das Ledger.** Kein Handler schreibt `UPDATE users SET money = ...`. Jede Geldbewegung geht durch `ledger.Post(tx, userID, amount, reason, ref)`. Das gibt Nachvollziehbarkeit bei Bug-Reports, macht Exploits sichtbar und ist die Grundlage für das Charity-Overflow-Feature.
6. **Kein neuer Service, solange eine Datei reicht.** SQLite bleibt, bis eine gemessene Grenze fällt – nicht, wenn sie jemand vermutet.

### 6.2 Schichten

```
Browser (HTML + htmx, kein Build-Step)
        │  Formular-POST / hx-post → HTML-Fragment zurück
        ▼
cmd/server/main.go          Routing, Wiring, Middleware-Kette, Shutdown
        │
        ▼
internal/handlers/          dünn: Formular parsen → Aktion rufen → rendern
        │                   nie Spielregeln, nie SQL-Details
        ▼
internal/actions/           Transaktionsgrenze: BeginTx → laden →
        │                   game-Funktion → speichern → Ledger → Commit
        │                   SQL steht hier direkt drin
        ▼
internal/game/              reine Regeln, keine DB, keine Zeit
```

**Nur eine Trennung, und die verdient sie sich:** die Wirtschaft ist der Teil, der schiefgehen wird – Balancing, Exploits, Rundungsfehler bei Geld. Genau der muss ohne HTTP und ohne DB in Millisekunden testbar sein. Deshalb ist `game` abgetrennt.

**Kein `internal/store`.** Eine Persistenzschicht zwischen Aktion und SQLite wäre eine Indirektion ohne zweiten Aufrufer: `actions` ist der einzige Nutzer, und die Queries sind größtenteils ein `SELECT`, das direkt neben seiner Transaktion am besten lesbar ist. `actions/collect.go` schreibt sein SQL selbst. Wenn dieselbe Query in der dritten Datei auftaucht, wandert **sie** in einen Helper – nicht vorsorglich das ganze Paket.

Alles andere bleibt flach: kein Interface mit einer Implementierung, kein Repository-Pattern um `waitlist`, kein DI-Container.

### 6.3 Paketkarte (Ziel nach Phase 3)

| Paket | Inhalt | Regel |
| --- | --- | --- |
| `cmd/server` | `main()`, Routen, Middleware, Sweeper-Start, Graceful Shutdown | keine Logik |
| `internal/config` | Env-Vars | bleibt wie es ist |
| `internal/db` | Verbindung, **Migrationen**, `WithTx` Helper | — |
| `internal/auth` | Session, CSRF, Passwörter, Mailer | fertig |
| `internal/httpx` | Rate-Limit, `ClientIP`, Security-Header, Request-Logging, Recovery | generisch, spielunabhängig |
| `internal/game` | **Balancing-Tabellen + reine Formeln**: Ertrag, Skillkosten, Kampf, Verbrechen, Behältergrößen | **darf nichts importieren außer stdlib** |
| `internal/actions` | eine Datei pro Spieleraktion: Transaktion, SQL, `game`-Aufruf | der ganze Bauch des Spiels |
| `internal/web` | Templates + Rendering | + Template-Funktionen (Geld, Dauer) |
| `internal/handlers` | HTTP | dünn |

### 6.4 Datenmodell (Zielbild)

`users` bleibt reiner Account. Spielzustand kommt in `characters` – **ein Character pro Edition pro Account**. Damit sind die vier NRW-Städte (Abschnitt 3.1) eine Spalte, kein Datenbank-Umbau.

```sql
editions(id, slug, name, active)   -- koeln, dortmund, duisburg, duesseldorf
                                   -- active=0 hält eine Stadt geschlossen
districts(id, edition_id, name, yield_factor, unlock_at_cents,
          risk_factor, distance_rank)                  -- Bezirke je Edition

characters(id, user_id, edition_id, name, district_id, money_cents,
           container_level, cleanliness, mood, created_at)
  -- UNIQUE(user_id, edition_id)   → max. 1 Character je Stadt
  -- UNIQUE(edition_id, name)      → Name nur je Stadt eindeutig
character_skills(character_id, skill, level)           -- PK(character_id, skill)
skill_queue(id, character_id, skill, ends_at, cost_cents, slot)  -- max 2 Slots
actions(id, character_id, kind, district_id, started_at, ends_at,
        resolved_at, payload_json, result_json)        -- laufende + Historie
inventory(character_id, item_id, qty)
ledger(id, character_id, delta_cents, balance_after, reason, ref_id, created_at)
fights(id, attacker_id, defender_id, created_at, winner_id, loot_cents, log_json)
fight_cooldowns(attacker_id, defender_id, until)       -- PK-Paar, 36 h
gangs(id, name UNIQUE, password_hash, bank_cents, created_at)
gang_members(gang_id, character_id, role, joined_at)
messages(id, from_id, to_id, subject, body, read_at, created_at)
referrals(id, character_id, code UNIQUE, clicks, last_click_at)
charity_pool(id, cents, source, created_at)            -- Overflow-Spenden
```

**Geld immer als `INTEGER` in Cent.** Nie Float. Ein einziger `0.1 + 0.2`-Fehler in einer Wirtschaftssimulation kostet mehr Debug-Zeit als die Umrechnung je gespart hätte.

**Editions-Isolation ist eine Invariante, keine Höflichkeit.** `gangs`, `fights`, `messages` und der Highscore hängen alle an `edition_id`. Zwei Regeln:

1. Jede Query, die andere Spieler auflistet oder trifft, filtert auf `edition_id`.
2. Es gibt **einen** Test, der genau das beweist: Character in Köln darf einen in Dortmund weder angreifen noch anschreiben noch in dessen Bande. Das ist die Art Bug, die man erst bemerkt, wenn Spieler ihn ausnutzen.

**Indizes von Anfang an:** `actions(character_id, resolved_at)`, `ledger(character_id, created_at)`, `sessions(expires_at)`, `fight_cooldowns(until)`, `characters(edition_id, money_cents)` (Highscore), `districts(edition_id)`.

### 6.5 Aktions-Lebenszyklus (das Herzstück)

```
POST /aktion/sammeln  (Dauer=1h, Bezirk=3)
  → CSRF prüfen · Rate-Limit
  → BeginTx
      1. offene Aktionen dieses Characters auflösen   (Lazy Tick)
      2. Vorbedingungen: keine laufende Aktion? Bezirk freigeschaltet?
      3. INSERT actions(kind='collect', ends_at=now+1h, payload={...})
         — Ergebnis wird NICHT vorab gewürfelt, sondern beim Auflösen
  → Commit → HTML-Fragment „Läuft bis 14:32“

Später, irgendein Request desselben Spielers:
  resolveDue(tx, characterID, now):
      für jede Aktion mit ends_at <= now und resolved_at IS NULL:
        ertrag := game.CollectYield(bezirk, dauer, skills, preisDesTages, rng)
        gutschrift, overflow := game.ApplyContainer(kontostand, ertrag, level)
        ledger.Post(+gutschrift, "collect", actionID)
        if overflow > 0 { charity_pool += overflow; ledger.Post(0, "overflow", ...) }
        cleanliness -= abnutzung
        UPDATE actions SET resolved_at=now, result_json=...
```

Drei Eigenschaften, die daraus gratis folgen:

- **Serverneustart schadet nichts** – nichts liegt im RAM.
- **Kein Doppel-Cash** – `resolved_at IS NULL` in derselben Transaktion ist die Sperre.
- **Ergebnis erst beim Auflösen würfeln** – sonst könnte man den Ausgang aus der DB-Zeile ableiten, sobald es je eine API oder ein Leak gibt.

### 6.6 Nebenläufigkeit

- SQLite läuft mit `SetMaxOpenConns(1)` → ein Schreiber, keine `SQLITE_BUSY`-Rennen. Reicht für ein rundenbasiertes Textspiel weit über tausend aktive Spieler hinaus.
- Genau **eine** Sweeper-Goroutine (`time.Ticker`, 1 Minute) für: passives Einkommen (Instrumente, 6 h) und das Aufräumen abgelaufener Session-/Reset-Tokens. Mehr nicht – der Pfandpreis wird aus dem Datum berechnet (keine Tabelle, kein Job) und der Highscore direkt aus `characters` sortiert (kein Snapshot, solange es keine Millionen Zeilen sind).
- Der Sweeper nutzt dieselben `actions.*`-Funktionen wie die Handler. Kein zweiter Code-Pfad, der die Wirtschaft anders rechnet.
- Wenn ein Schreiber je zum Engpass wird: erst messen, dann Lesepfade auf eine zweite Read-Only-Verbindung legen, erst danach über Postgres reden.

### 6.7 Sicherheit – Basislinie

| Maßnahme | Status |
| --- | --- |
| bcrypt, httpOnly + Secure + SameSite=Lax Cookies | [x] |
| CSRF Double-Submit auf allen POSTs | [x] – **außer `/waitlist`** (D3) |
| Rate-Limit auf Login/Register/Reset | [x] – fehlt auf `/waitlist` und später auf allen Aktionen |
| Sessions bei Passwortänderung invalidieren | [ ] D4 |
| Server-Timeouts + Graceful Shutdown | [ ] D6 |
| Security-Header (`X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`, CSP) | [ ] D9 |
| htmx lokal ausliefern statt CDN → erlaubt strenge CSP | [ ] D8 |
| Alle Ausgaben über `html/template` (Auto-Escaping) | [x] – **niemals** `template.HTML` auf Nutzereingaben (Bandenchat, Profile, Signaturen!) |
| Doppelklick-Schutz über die Transaktion („keine laufende Aktion“) + `hx-disabled-elt` | [ ] Phase 3 |
| Serverseitige Autorität: nie ein Ergebnis aus dem Formular übernehmen | Regel |

### 6.8 Testkonzept

Drei Ebenen, alle mit `go test ./...`, keine externe Infrastruktur. Der Aufwand steigt nach unten, der Nutzen sinkt — deshalb ist die unterste Ebene bewusst dünn.

#### Ebene 1 – Unit: `internal/game` *(hier liegt der Schwerpunkt)*

Reine Funktionen, keine DB, keine Uhr, kein HTTP. **Tabellentests** (`[]struct{...}` + `t.Run`), Millisekunden Laufzeit.

`game` bekommt Zeit und Zufall **hineingereicht**, nicht selbst geholt:

```go
func CollectYield(d District, dur time.Duration, sk Skills, price int, rng *rand.Rand) int
```

Dadurch ist jedes Ergebnis mit `rand.New(rand.NewSource(1))` reproduzierbar — kein flaky Test, kein `time.Sleep` in der Suite.

Was hier abgedeckt wird:

- Ertragsformel: Grenzen (0 min, 12 h), abnehmender Grenzertrag, Editions-Multiplikator (Köln-Event / Dortmund-Spieltag / Duisburg-konstant / Düsseldorf-Hochpreis)
- Behälter: exakt am Limit, 1 Cent drüber, Overflow-Betrag stimmt cent-genau
- Skillkosten `basis × (level+1)`, 80-%-Erstattung inkl. Rundung
- Kampfformel: ATT/DEF-Extreme, Verlustdeckelung, Neulingsschutz
- Verbrechen: Erfolgswahrscheinlichkeit an den Rändern (0 Skill / max Skill)
- **Ein Schleifentest über Zufallseingaben:** über ein paar tausend Durchläufe muss immer `ertrag >= 0` und `gutschrift + overflow == ertrag` gelten. Zehn Zeilen, findet Rundungslecks, an die kein Tabellenfall denkt.

#### Ebene 2 – Integration: `internal/actions` gegen `:memory:`-SQLite

Echtes Schema, echte Transaktionen, kein Netzwerk. Ein Helper `newTestDB(t)` legt eine frische In-Memory-DB an, spielt alle Migrationen ab und seedet einen Character — pro Test ~1 ms.

Pro Aktion **ein** Happy Path plus die kritischen Grenzen:

| Test | Beweist |
| --- | --- |
| Aktion doppelt auflösen | zahlt genau **einmal** (`resolved_at`-Sperre hält) |
| Zwei gleichzeitige Requests, dieselbe Aktion | kein Doppel-Cash (Transaktion serialisiert) |
| Ledger-Summe == `characters.money_cents` | Buchhaltung stimmt, keine Geldquelle am Ledger vorbei |
| Behälterüberlauf | Differenz landet cent-genau in `charity_pool` |
| Kampf-Cooldown | zweiter Angriff < 36 h wird abgewiesen |
| Skill-Queue | 3. Skill wird abgewiesen, Abbruch erstattet 80 % |
| **Editions-Isolation** | Köln kann Dortmund nicht angreifen/anschreiben/rekrutieren |
| Migration von leer | alle Migrationen laufen sauber durch |

#### Ebene 3 – End-to-End: `httptest` gegen den echten Handler-Stack

Ja, das lohnt sich — **wenn es billig bleibt**. Mit `httptest.NewServer(mux)` und einem `http.Client` mit `cookiejar` bekommst du echte Sessions, echte Cookies, echtes CSRF, echtes HTML, ohne Browser und ohne Docker. Laufzeit: unter einer Sekunde für die ganze Strecke.

```go
srv := httptest.NewServer(buildMux(testDeps(t)))
jar, _ := cookiejar.New(nil)
c := &http.Client{Jar: jar}          // Cookies + CSRF verhalten sich wie im Browser
```

Drei Szenarien, mehr nicht:

1. **Auth-Strecke:** Registrieren → Bestätigungsmail abfangen (Test-Mailer sammelt in einem Slice) → Link aufrufen → Login → geschützte Seite → Logout → geschützte Seite ist wieder gesperrt
2. **Spielschleife:** Login → Character in Dortmund anlegen → Sammelaktion starten → Uhr vorstellen (injizierte Clock, kein `Sleep`) → Dashboard laden → Geld ist gutgeschrieben
3. **Sicherheit:** POST ohne CSRF-Token → 403 · geschützte Route ohne Session → Redirect · 11 Logins in einer Minute → 429

Geprüft wird **Verhalten**, nicht Markup — Statuscode, Redirect-Ziel, DB-Zustand, und höchstens ein `strings.Contains` auf einen Kerntext. Assertions gegen HTML-Struktur brechen bei jeder Designänderung und testen nichts, was kaputtgehen kann.

> **Kein Playwright, kein Selenium, kein Testcontainers.** Ein Browser-Runner wäre hier reine Zeremonie: die App hat kein clientseitiges JS, das Zustand hält — htmx tauscht nur vom Server geliefertes HTML. Es gibt schlicht nichts, was nur ein echter Browser aufdecken würde. Falls später doch komplexe Client-Interaktion dazukommt: Playwright nachrüsten, dann aber **eine** Smoke-Strecke, nicht eine Suite.

#### Was bewusst *nicht* getestet wird

Getter/Setter, Template-Rendering-Details, `config.Load()`, Mocks für Dinge mit einer Implementierung, Coverage als Zielwert. Eine grüne Zahl ist kein Beweis — die acht Grenzfälle in Ebene 2 sind einer.

#### Werkzeug

- Nur stdlib `testing`. Kein testify, kein ginkgo, kein gomock.
- `t.Parallel()` in Ebene 1, **nicht** in Ebene 2/3 (eine SQLite-Verbindung, ein Schreiber).
- Keine Golden Files. Erwartete Werte stehen direkt in der Testtabelle – sichtbar im Diff, ohne zweites Dateiformat und ohne `-update`-Flag.
- CI (Phase 1.5) fährt `go vet ./... && go test -race ./...` bei jedem Push. `-race` findet genau die Nebenläufigkeitsfehler, die im Sweeper lauern.

### 6.9 Betrieb

- Deploy: `docker compose up -d --build` auf einem VPS, Caddy terminiert TLS.
- **Backup ist Pflicht, sobald echte Nutzer da sind**: nächtlich `sqlite3 db .backup` in eine Datei + Kopie weg vom Server (Restic/rclone). Ein Spiel ohne Backup ist ein Spiel mit Ablaufdatum.
- Logs: `log/slog` als JSON auf stdout, Docker rotiert.
- `/healthz` für Uptime-Monitoring (UptimeRobot Free).

### 6.10 Bewusst weggelassen

Redis, Message-Queue, Microservices, GraphQL, ORM, React, Node-Build-Chain, Kubernetes, Event-Sourcing, WebSockets, Feature-Flag-Service, Observability-Stack. Jeweils nachrüstbar, wenn eine **gemessene** Grenze fällt – keiner davon vorher.

---

## 7. Phasenplan

| Phase | Ziel | Grobe Dauer* |
| --- | --- | --- |
| 0 | Landingpage/Warteliste live | 1 Woche |
| 1 | Grundgerüst: Auth, DB, Deploy-Pipeline | 2–3 Wochen |
| **1.5** | **Härtung: Timeouts, Header, Logging, Backup, CI** | **~3 Tage** |
| 2 | Design-System umsetzen | 1 Woche |
| 3 | Kernschleife (Wirtschaft) | 3–4 Wochen |
| 4 | Sozial & PvP | 3–4 Wochen |
| 5 | Mobile/PWA-Polish, Notification-Entscheidung | 1–2 Wochen |
| 6 | Monetarisierung | 1 Woche |
| 7 | Launch & Community | laufend |

*Richtwerte bei ein paar Stunden/Woche nebenbei – wird real wahrscheinlich mehr, siehe Aufwandsschätzung von letztem Mal.

Reihenfolge-Empfehlung, abweichend von der reinen Nummerierung: **Phase 1.5 (Härtung + Deploy-Durchstich) vor Phase 2**, und **Phase 2 vor Phase 3**. Grund: einmal live deployen, solange die App noch trivial ist, ist billig; ein Design-System nachträglich über 30 fertige Seiten zu ziehen ist teuer.

---

### Phase 0 – Landingpage/Warteliste — *Code fertig, Betrieb offen*

**Ziel:** Die Seite ist unter der echten Domain erreichbar und sammelt E-Mails.

| # | Aufgabe | Wer |
| --- | --- | --- |
| 0.1 | Name entscheiden | 🙋 **DU** — A1, blockiert 0.2–0.6 |
| 0.2 | Domain registrieren, DNS A-Record auf VPS-IP | 🙋 **DU** — A2/A3 |
| 0.3 | Teaser-Text final schreiben (Platzhalter steht, Formulierung siehe Abschnitt 3) | 🙋 **DU** |
| 0.4 | `SITE_NAME`, `SITE_DOMAIN` in `.env` setzen | 🙋 **DU** — A4 |
| 0.5 | **Impressum + Datenschutzerklärung** als statische Seiten | 🙋 **DU** liefert Inhalt (Adresse, verantwortliche Person), ich baue die Routen |
| 0.6 | **Double-Opt-In für die Warteliste** — DSGVO-relevant, weil du die Adressen später anschreiben willst | ich |
| 0.7 | CSRF + Rate-Limit auf `/waitlist` (D3) | ich |
| 0.8 | Waitlist-Export als CSV hinter Admin-Token | ich |
| 0.9 | Posten und Reaktionen beobachten | 🙋 **DU** |

**Definition of Done:** `https://<domain>/` liefert die Landingpage mit gültigem TLS, eine Testadresse landet nach Klick auf den Bestätigungslink als `confirmed` in der DB, Impressum und Datenschutz sind verlinkt.

> ⚠️ **Rechtlich:** Eine E-Mail-Adresse ohne Double-Opt-In einzusammeln und später anzuschreiben ist in Deutschland abmahnfähig. 0.5 und 0.6 sind keine Politur, sondern Voraussetzung dafür, die Seite öffentlich zu bewerben.

---

### Phase 1 – Grundgerüst — *~80 % fertig*

**Ziel:** Ein Mensch kann sich registrieren, seine Adresse bestätigen, einloggen, das Passwort zurücksetzen — und das läuft auf einem echten Server.

Erledigt: Schema, Registrierung, Login, Logout, Passwort-Reset, bcrypt, Session-Cookies, CSRF, Rate-Limit, Basis-Layout, Docker + Caddyfile.

**Was noch fehlt:**

| # | Aufgabe | Details |
| --- | --- | --- |
| 1.1 | **Migrations-System** (D11) | `internal/db/migrations/0001_init.sql`, `0002_....sql` per `embed.FS`, Tabelle `schema_migrations`, beim Start der Reihe nach anwenden. ~40 Zeilen, keine Library. **Muss vor der ersten echten Nutzer-DB stehen** — danach ist jedes Schema-Update Handarbeit auf Prod. |
| 1.2 | **Bestätigungsmail / Double-Opt-In** | `users.confirmed_at`, Token wie beim Reset, unbestätigte Accounts können sich einloggen, aber nicht spielen. |
| 1.3 | **Echter Mailversand** | `SMTPMailer` neben `ConsoleMailer`, Auswahl per Env-Var. 🙋 **DU:** Anbieter + SPF/DKIM-DNS-Records (A6). |
| 1.4 | Sessions bei Passwortänderung löschen (D4) | Eine Zeile in `ResetPassword`. Sicherheit, nicht Komfort. |
| 1.5 | `Register`-Fehlerbehandlung korrigieren (D5) | Nur auf UNIQUE-Verletzung `ErrEmailTaken`, sonst durchreichen. |
| 1.6 | `Dockerfile` auf `golang:1.27-alpine` (D1) | Eine Zeile. Ohne das schlägt jeder Container-Build fehl. |
| 1.7 | **Deploy-Durchstich** | 🙋 **DU:** VPS, DNS, `.env`. Ich: Deploy-Anleitung + Healthcheck. |
| 1.8 | Erster Git-Commit + Remote | 🙋 **DU** (A5). Repo hat aktuell **null** Commits — ein `rm -rf` und alles ist weg. |

**Definition of Done:** Registrierung auf dem Live-Server → Mail kommt real an → Bestätigen → Login → Reset funktioniert → Neustart des Containers verliert keine Daten.

---

### Phase 1.5 – Härtung *(neu, ~3 Tage)*

**Ziel:** Der Server übersteht öffentliches Internet. Fertigstellen, bevor die Landingpage beworben wird.

- [ ] `http.Server` mit `ReadHeaderTimeout` / `ReadTimeout` / `WriteTimeout` / `IdleTimeout` + Graceful Shutdown auf SIGTERM (D6)
- [ ] Middleware-Kette: Panic-Recovery → Request-Logging (`log/slog`, JSON) → Security-Header → Handler
- [ ] Security-Header: `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`, CSP
- [ ] htmx lokal unter `static/js/` ausliefern statt CDN (D8) — macht CSP erst möglich und entfernt eine Laufzeit-Abhängigkeit von unpkg
- [ ] `GET /healthz` (D9) + 🙋 **DU:** UptimeRobot einrichten (kostenlos)
- [ ] Sweeper-Goroutine: abgelaufene Sessions/Reset-Tokens löschen (D7)
- [ ] `.DS_Store` entfernen, `static/` sauber halten (D10)
- [ ] GitHub Actions: `go vet`, `go test`, `go build` bei Push (D12)
- [ ] **Backup-Skript**: nächtliches `.backup` + Kopie außerhalb des Servers. 🙋 **DU:** Zielort (S3/Storage Box) + einmal einen Restore testen. Ein ungetestetes Backup ist kein Backup.

---

### Phase 2 – Design-System *(1 Woche)*

**Ziel:** Alle bestehenden Seiten im finalen Look; jede neue Seite ab hier ist automatisch richtig gestylt.

- [ ] Tailwind Standalone-CLI als Binary im Repo/Makefile (kein Node), `make css` + `make watch`
- [ ] Tokens aus Abschnitt 5 in `tailwind.config.js`: Asphalt, Pfandglas-Grün, Verkehrsorange, Münzgold, Pappe, Backstein-Rot
- [ ] Fonts **selbst hosten** (`static/fonts/`) — kein Google-Fonts-CDN. Spart einen DSGVO-Fallstrick und einen Roundtrip.
- [ ] Komponenten-Partials: `button`, `card`, `form_field`, `statusbar`, `flash`, `progress`
- [ ] **Sticky Statusleiste**: Geld, Laune, laufender Cooldown
- [ ] **Signature-Element: der Geld-Behälter** als SVG-Partial mit 5 Stufen (Becher → Tüte → Beutel → Einkaufswagen → Container) und Füllstand aus `money/limit`
- [ ] Bottom-Nav (Übersicht/Aktionen/Stadt/Bande/Kampf), mobile-first
- [ ] Landing, Login, Register, Reset, Dashboard umstellen
- [ ] CSS-Build in den Docker-Build integrieren
- [ ] 🙋 **DU:** Look einmal abnehmen — Farben und Typo sind Geschmack, nicht Technik

**Definition of Done:** Auf einem echten Handy sieht jede bestehende Seite fertig aus. Kein `<script>` von fremden Domains.

---

### Phase 3 – Kernschleife *(3–4 Wochen, der eigentliche Kern)*

**Ziel:** Ein Spieler kann sammeln, verkaufen, investieren, aufsteigen — die Schleife trägt sich allein.

> **Erst spielbar, dann vollständig.** 3a + 3b + 3c ergeben zusammen ein Spiel, das man tatsächlich spielen kann: sammeln, Geld ansammeln, Behälter aufrüsten. Das ist der Punkt, an dem du zum ersten Mal merkst, ob die Schleife trägt — und der richtige Moment, das Balancing zu prüfen, **bevor** Skills, Shop und Verbrechen dieselben Zahlen nochmal verkomplizieren. Wenn sich der Kern hier langweilig anfühlt, retten ihn 3d–3f nicht.

**3a – Fundament (zuerst, alles andere hängt daran)**

- [ ] Migration: `characters`, `ledger`, `actions`, `districts`
- [ ] `internal/game/` mit Balancing-Tabellen als Go-Konstanten + reine Formeln, **ohne DB, ohne `time.Now()`**
- [ ] `ledger.Post()` als **einziger** Weg, Geld zu bewegen
- [ ] `db.WithTx()` Helper
- [ ] Character-Erstellung beim ersten Login (Name wählen, 10 € Startkapital)

**3b – Sammeln**

- [ ] `POST /aktion/sammeln` mit Dauerwahl (10 min – 12 h), ein Bezirk
- [ ] Ertragsformel: `bezirk_faktor × dauer × geschicklichkeit × tagespfandpreis × rng`, mit **abnehmendem Grenzertrag** bei langer Dauer (Original: kurz = effizienter pro Zeiteinheit)
- [ ] Lazy-Tick-Auflösung (Abschnitt 6.5) + Fortschrittsanzeige, die per htmx-Polling aktualisiert
- [ ] Sauberkeit sinkt durchs Sammeln
- [ ] Tagespfandpreis: ein Wert pro Tag, deterministisch aus dem Datum abgeleitet — für alle Spieler gleich, ohne extra Tabelle

**3c – Geld-Behälter (das Signature-Feature)**

- [ ] 5 Stufen mit je einem Cent-Limit, Upgrade im Shop
- [ ] Überlauf geht in `charity_pool` statt verloren (Kuration Abschnitt 2)
- [ ] Warnung in der Statusleiste ab 80 % Füllstand
- [ ] Öffentlicher Zähler „bisher X € gespendet" — der Grund, warum sich das Feature gut statt frustrierend anfühlt

**3d – Skills**

- [ ] 9 Skills aus Abschnitt 1.3, Kosten `basis × (level+1)`
- [ ] Warteschlange mit 2 Slots, Abbruch erstattet 80 %
- [ ] Auflösung über denselben Lazy Tick
- [ ] Skill-Effekte kommen aus `internal/game`, nie aus Handler-Code

**3e – Stadt & Shop**

- [ ] Bezirke mit Freischalt-Bedingungen
- [ ] Items: Behälter-Upgrades, Ausrüstung, Instrumente (passives Einkommen alle 6 h), Wohnung
- [ ] Inventar
- [ ] Instrumenten-Einkommen im Sweeper

**3f – Verbrechen**

- [ ] Aktion mit Risiko/Ertrag, Erfolgswahrscheinlichkeit aus Skills
- [ ] Fehlschlag: Geldverlust + Cooldown („Ausnüchterungszelle")
- [ ] Ergebnis **beim Auflösen** würfeln, nicht beim Starten

**3g – Absicherung**

- [ ] Tabellentests für jede Formel in `internal/game`
- [ ] Aktions-Tests gegen `:memory:`-SQLite: Doppelauflösung, Behälterüberlauf, Cooldown
- [ ] Rate-Limit pro Character auf allen Aktionsrouten
- [ ] Doppelklick-Schutz: die Vorbedingung „keine laufende Aktion“ in derselben Transaktion erledigt das bereits — der zweite Request findet die Aktion vor und wird abgewiesen. Dazu `hx-disabled-elt` am Button. **Kein Idempotenz-Token**; das wäre eine zweite Sperre für ein Problem, das die erste schon löst.
- [ ] 🙋 **DU:** Balancing spielen und bewerten. Zahlen kann ich setzen, ob sich die Kurve gut *anfühlt*, kannst nur du entscheiden.

**Definition of Done:** Du spielst 30 Minuten ohne Handbuch, verstehst die Schleife und willst weiterspielen. Alle Geldbewegungen stehen im Ledger.

> ⚠️ **Größtes Risiko der ganzen Phase:** Balancing, nicht Code. Zahlen in eine zentrale Tabelle, damit Nachjustieren ein Zahlen-Edit bleibt und kein Refactoring wird.

---

### Phase 4 – Sozial & PvP *(3–4 Wochen)*

**Ziel:** Andere Spieler sind der Grund, morgen wiederzukommen.

**4a – Kampf**

- [ ] `fights`, `fight_cooldowns`; Ausgang aus ATT/DEF + Zufall
- [ ] 36-h-Cooldown pro Gegnerpaar (Anti-Griefing, Werte tunen)
- [ ] Kampfdauer abhängig von der Bezirks-Distanz
- [ ] Kampflog mit Historie; ab Geschicklichkeit 20 eingehende Angriffe sehen
- [ ] Verlust gedeckelt (DEF senkt ihn) — **Neulingsschutz: die ersten 7 Tage oder unter X € nicht angreifbar.** Ohne das killen Veteranen jeden Neuzugang.

**4b – Banden**

- [ ] Erstellen/Beitreten (Name + Passwort), zwei Rollen: Chef und Mitglied. Ein feinkörniges Rechtesystem ist etwas für den Tag, an dem sich jemand über die zwei beschwert.
- [ ] Bandenkasse; Ausbauten (Waffenkammer/Haus/Training/Konto), je 5 Stufen mit %-Boni
- [ ] Bandenchat — **Auto-Escaping streng, kein `template.HTML`**
- [ ] Bandenhighscore
- [ ] Bündnisse und Kriege: **erst nach dem Launch**, wenn genug Banden existieren, dass Diplomatie überhaupt eine Bedeutung hat. Vorher baut man ein Feature für drei Banden.

**4c – Kommunikation**

- [ ] Nachrichten, Freunde, **Blockliste**
- [ ] Meldefunktion + Admin-Ansicht. 🙋 **DU:** Moderationsregeln festlegen. Ein Spiel mit Chat braucht ab Tag 1 eine Antwort auf „jemand ist widerlich" — sonst regelt es sich von selbst, und zwar schlecht.

**4d – Highscore & Referral**

- [ ] Highscore Spieler + Banden, direkt aus `characters` sortiert — kein Snapshot-Job, solange ein `ORDER BY` mit Index unter einer Millisekunde bleibt
- [ ] **Referral-Link** `/kleingeld/{code}` (Abschnitt 1.5): Klick → In-Game-Spende, IP-/Zeitfenster-Deduplizierung gegen Selbstklick-Farming
- [ ] Profil mit Signatur-Snippet zum Kopieren

**Definition of Done:** Zwei Testaccounts können kämpfen, eine Bande gründen, chatten. Ein Neuling kann nicht in Grund und Boden geprügelt werden.

---

### Phase 5 – Mobile/PWA *(1–2 Wochen)*

- [ ] `manifest.json`, Icons, installierbar — das ist der eigentliche Nutzen: Icon auf dem Homescreen, Vollbild ohne Browserleiste
- [ ] Performance-Pass: Lighthouse mobil > 90, HTML-Antwort < 30 kB
- [ ] Service Worker + Offline-Seite: **nur wenn nach dem Manifest noch Lust da ist**. Ein Spiel, das ohne Server nichts anzeigen kann, gewinnt durch eine Offline-Seite genau einen hübscheren Fehlertext. Wenn doch: ausschließlich statische Assets cachen, niemals Spielzustand — ein Service Worker, der HTML cached, zeigt Spielern veralteten Kontostand.
- [ ] Touch-Targets ≥ 44 px, Tastatur-Fokus sichtbar, Kontrast AA — die Tokens sind dunkel, das muss geprüft werden
- [ ] Web-Push: **weglassen** (Entscheidung aus dem Plan bleibt). Stattdessen zeigt die Statusleiste, wann die Aktion fertig ist.
- [ ] 🙋 **DU:** auf deinem echten Handy testen. Emulator zählt nicht.

---

### Phase 6 – Monetarisierung *(1 Woche)*

- [ ] „Unterstützer"-Status: werbefrei, Badge, Kosmetik — **null Gameplay-Vorteil** (wie im Original)
- [ ] Kosmetik-Shop
- [ ] Stripe oder Paddle. **Empfehlung: Paddle** — Merchant of Record, übernimmt EU-Umsatzsteuer komplett. Mit Stripe machst du OSS-Meldungen selbst.
- [ ] Webhook-Handler, **idempotent** (Zahlungs-Webhooks kommen doppelt), Signatur prüfen
- [ ] Spenden-Button + Charity-Kopplung des Overflow-Pools
- [ ] 🙋 **DU:** Paddle-Account + Steuerdaten (A6-artig), **Charity-Partner auswählen und anschreiben** (echte Obdachlosenhilfe — das Versprechen aus Abschnitt 2 muss real sein), AGB + Widerrufsbelehrung

> ⚠️ Ab hier nimmst du Geld ein: Gewerbeanmeldung, Umsatzsteuer-Frage, AGB, Widerruf. 🙋 **DU** — an dem Punkt lohnt sich eine Stunde Steuerberater.

---

### Phase 7 – Launch & Community *(laufend)*

- [ ] Closed Beta an die Warteliste, „Urgestein"-Badge für früh Eingeloggte
- [ ] Onboarding: „Warst du dabei?"-Nostalgie-Quiz, rein kosmetisch
- [ ] 🙋 **DU:** Discord-Server (ab Phase 0!), Reddit (Selbstpromo-Regeln vorher lesen, immer als Entwickler kenntlich), TikTok/Reels im „erinnert ihr euch"-Format
- [ ] Referral-Link prominent (ist ab Phase 4 gebaut)
- [ ] Öffentliches Changelog — billigster Weg, Vertrauen aufzubauen
- [ ] Metriken: DAU, D1/D7-Retention, Warteliste→Registrierung-Konversion. Ohne Retention-Zahl weißt du nicht, ob das Balancing trägt.

---

## 8. Risiken

| Risiko | Wirkung | Gegenmaßnahme |
| --- | --- | --- |
| Namens-/Markenkonflikt | Rebrand nach Launch | A7 vor dem ersten öffentlichen Post |
| Balancing trägt nicht | Spieler brechen nach Tag 2 ab | Zahlen zentral, früh mit echten Testern messen |
| Kein Backup, DB weg | Projekt tot | Phase 1.5, Restore **testen** |
| Neulinge werden von Veteranen zerlegt | Retention kollabiert | Neulingsschutz in 4a |
| Chat-Missbrauch | rechtliches + Reputationsproblem | Melden + Blockliste + Moderationsregeln in 4c |
| Scope-Kriechen | nie fertig | Reihenfolge einhalten, Phase 4 erst nach 3g |
| Solo-Projekt, Zeit | Stillstand | Phase 1.5 macht die Seite betriebsfähig, auch wenn Monate Pause kommen |
| Vier Editionen, zu wenig Spieler | leere Städte, PvP tot | zum Launch nur zwei Städte offen (3.1) |
| Plan wird gebaut statt Spiel | Monate ohne spielbaren Stand | nach 3c anhalten und spielen, bevor 3d–3f folgen |

---

## 9. Community: Alt + Neu

- **Alte Spieler**: der ehrliche "inspiriert von"-Hinweis auf der Landingpage plus vertraute Begriffe/Mechaniken (Pfandflaschen, Container-Progression, Bandenkriege) lösen Wiedererkennung aus
- **Neue Spieler**: schnelles, mobile-first Onboarding, keine Verifizierung alter Accounts nötig (technisch eh nicht möglich) – stattdessen optional ein "Warst du dabei?"-Nostalgie-Quiz beim Onboarding, rein für Spaß/Badge, keine echte Prüfung
- Früh eingeloggte Beta-Tester bekommen einen kosmetischen "Urgestein"-Badge – belohnt Treue ohne Pay-to-win-Charakter

---

## Quellen

- Offizielle Pennergame-Anleitung (Hamburg-/Berlin-/Atlantis-Edition), erschlossen über Suchindex
- Ausführlicher Fan-Guide, pennergame-pro.blogspot.com (2009)
- Handelsregisterauszug Farbflut Entertainment GmbH, HRB 104994
- giga.de, gutefrage.net – aktueller Seitenstatus (2026)
- Stellenanzeige Farbflut Entertainment, python-forum.de – Tech-Stack-Hinweis

---

**Nächster Schritt – empfohlene Reihenfolge:**

1. **Ich, sofort machbar, blockiert von nichts:** D1 (Dockerfile-Version), D3 (CSRF+Rate-Limit auf `/waitlist`), D4 (Sessions bei Passwortänderung), D5 (Register-Fehler), D10 (`.DS_Store`) — fünf kleine Fixes.
2. **Ich, direkt danach:** Punkt 1.1 (Migrationen) — muss vor der ersten echten Nutzer-DB stehen.
3. **Ich:** Phase 1.5 komplett (Timeouts, Header, Logging, Healthz, CI, Backup-Skript).
4. **🙋 DU parallel:** A1 Name → A2 Domain → A3 VPS → A4 `.env` → A5 Git-Remote. Danach ist Phase 0 live schaltbar.
5. Dann Phase 2 (Design), dann Phase 3 (Kernschleife).

Sag, womit ich anfangen soll.
