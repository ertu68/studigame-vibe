# Entwicklungsplan: Studigame

*Stand: 30.08.2026 – lebendes Dokument, wird während der Entwicklung laufend angepasst.*

Browser-Game im Geist von Pennergame (2007), aber eigenständig: **Studentenleben
statt Obdachlosigkeit.** Go, server-rendered, SQLite, kein JS-Build. Spielwelt
ist eine fiktive Uni-Stadt, das Studienfach übernimmt die Rolle des
Klassensystems.

Dieses Dokument ist die einzige Quelle der Wahrheit für Konzept, Architektur
und Umfang. `ROADMAP.md` hält die offenen To-Dos, `README.md` den Stand,
`CHANGELOG.md` die Änderungen. Widerspricht sich etwas, gewinnt dieser Plan.

---

## 1. Funktionsanalyse: Das Original

Die offizielle Anleitung (Hamburg-, Berlin-, Atlantis-Edition) und ein
ausführlicher Fan-Guide von 2009 sind über den Suchindex erschlossen. Dadurch
lässt sich das Spiel fast vollständig rekonstruieren. Diese Analyse ist die
Basis für die Kuration in Abschnitt 2.

### 1.1 Grundgerüst

- Start: 10 € Startkapital, Ziel: vom Pennerleben zum "Schlossbesitzer"
- Mehrere parallele "Editionen" pro Stadt (Hamburg, Berlin, Köln, München,
  Mallorca, dazu ein Fun-Server "Atlantis/Vatikan")
- Finanzierung nur über Spenden + "Ehrenmitgliedschaft" – Premium gab laut
  eigener Aussage des Originals keine echten Gameplay-Vorteile, nur
  Kosmetik/Komfort (eigenes Haustier-Design, größeres Bandenavatar, eigene
  E-Mail-Adresse, kein Werbebanner)

### 1.2 Die Kernschleife

1. **Ressourcen sammeln** – Pfandflaschen sammeln (Dauer wählbar 10 Min–12 h;
   kürzer = effizienter pro Zeiteinheit, aber absolute Menge steigt mit Dauer;
   Ertrag hängt von Stadtteil, Geschicklichkeit und tagesaktuellem Pfandpreis ab)
2. **Verkaufen/Betteln** – Flaschen verkaufen, dazu passives Bettel-/Spenden-Einkommen
3. **Investieren** – in Skills, Ausrüstung, Wohnung
4. **Risiko eingehen** – Verbrechen begehen (mehr Ertrag, aber Erwischt-Risiko)
   oder andere Spieler angreifen
5. **Aufsteigen** – bessere Stadtteile, Ausrüstung, Bande, Highscore-Rang

Ein Geld-"Behälter" begrenzt, wie viel man auf einmal tragen kann
(Trinkbecher → Tüte → Beutel → Einkaufswagen → Container). Wer sein Limit nicht
rechtzeitig upgradet, verliert überschüssiges Geld. Kleine, aber clevere
Mechanik: zwingt zum Reinvestieren und passt zum Thema.

### 1.3 Skillsystem ("Weiterbildung")

Kosten = Basiswert × nächstes Level. Abbruch erstattet 80 %. Zwei Skills
gleichzeitig in der Warteschleife möglich.

*Vermutliche Ursache fürs "schnell overpowered"-Problem: diese Kostenformel ist
**linear**. Wächst das Einkommen mit dem Level mit, wird jedes weitere Level
relativ günstiger statt teurer – ein Schneeball-Effekt. Fix: überlinear
skalieren (z. B. Basiswert × Level^1.8), siehe Abschnitt 3.*

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

- **Kampf**: gleicher Gegner nur alle 36 h angreifbar, Kampfdauer abhängig von
  der "Entfernung" der Wohnbezirke, Kampf-Log mit Gewinn/Verlust-Historie
- **Haustierkämpfe**: separates Wett-System mit Einsatz und Stances
  (Aggressiv/Defensiv/Neutral) – nebenbei der einzige Weg, Geld direkt an
  andere Spieler zu senden
- **Banden**: Name+Passwort, Bandeneigentum (Waffenkammer/Bandenhaus/Training/
  Konto, je 5 Stufen mit %-Boni für alle Mitglieder), Bandenchat,
  Bündnisse/Kriege, eigener Bandenhighscore
- **Freunde/Blockliste, Nachrichten, Forensignaturen** (jeder Klick auf die
  Signatur triggert eine Spende!)
- **Highscore** für Spieler und Banden getrennt

### 1.5 Der eingebaute Wachstumsmotor

Das Original hatte einen Referral-Mechanismus direkt in der Spielfantasie
eingebaut. Jeder Spieler bekommt einen persönlichen Link
(`.../change_please/{RefID}/` – "haste mal Kleingeld?"), den man in
Forensignaturen oder der eigenen Profil-Homepage postet. Klickt jemand drauf,
gibt es eine In-Game-Spende. Virales Wachstum, direkt an die Kernfantasie
(Betteln) gekoppelt, ohne wie ein aufgesetzter "Invite 3 friends"-Hack zu wirken.

### 1.6 Technischer Befund

- Backend laut einer alten Stellenanzeige von Farbflut Entertainment:
  **Python + MySQL**
- Die Domain läuft technisch noch, Neuregistrierung ist kaputt, keine
  Weiterentwicklung mehr
- Farbflut Entertainment GmbH (Hamburg) ist laut Handelsregister weiterhin
  wirtschaftsaktiv – der Markenname "verwaist" also nicht, deshalb weiterhin:
  eigenen Namen verwenden

---

## 2. Kuration: Übernehmen, ändern, weglassen

*Thematische Verpackung ist "Studigame" statt Penner-Thema (siehe Abschnitt 3).
Die Mechanik-Entscheidungen gelten so; eine Ausnahme (Bezirk → Studienfach)
ist in 3a dokumentiert.*

| Element | Entscheidung | Warum |
| --- | --- | --- |
| Grundschleife (sammeln → verkaufen → investieren → Risiko → aufsteigen) | **Übernehmen** | Zeitlos, funktioniert einfach |
| Geld-Behälter mit Kappung | **Übernehmen, leicht verändert** | Statt "Geld weg für immer" bei Überlauf: automatische Spende an eine echte Obdachlosenhilfe/Studierendenwerk. Gleiche Spielspannung, fühlt sich aber gut statt frustrierend an |
| Promille-System | **Übernehmen** | Passt im Studenten-Kontext (Vorglühen, WG-Partys) organischer als im Original – mit Augenmaß umgesetzt |
| Sauberkeit sinkt durchs Sammeln, erhöht Spenden | **Weglassen (MVP)** | Willkommene Bonus-Ressource, aber zweite parallel pflegbare Ressource = doppelte UI-/Balancing-Fläche für ein Feature ohne echten Sog. Parken für "Später" (§7b), nicht streichen |
| Referral-Link als "Spendenlink" | **Übernehmen & ausbauen** | Genialer eingebauter Wachstumsmotor, siehe 1.5 |
| Ehrenmitgliedschaft ohne Gameplay-Vorteil | **Übernehmen** | Deckt sich mit der Monetarisierungs-Empfehlung |
| Reale Städte/Stadtteile (Hamburg, St. Pauli, …) | **Ändern → fiktive Uni-Stadt** | Mehr kreative Freiheit, keine Nähe zum Original; Rolle des Orts-Multiplikators übernimmt das Studienfach (3a) |
| Shock-Humor bei manchen Item-/Haustiernamen | **Weglassen** | Harter, dated 2000er-Forenhumor, der schon damals grenzwertig war |
| "Dein Leben beenden" als Label für Account löschen | **Ändern** | Klingt nach Selbstverletzung – einfach "Account löschen" |
| 36h-Cooldown pro Angriffsziel | **Übernehmen, Werte tunen** | Guter Anti-Griefing-Mechanismus |
| Kein direkter Geldtransfer (nur Bandenkasse/Haustierkämpfe) | **Übernehmen** | Verhindert simples Farming/RMT |
| Textbasiert, schlichtes UI | **Übernehmen, modernisieren** | Gleicher Minimalismus, aber mobile-first & schnell |

---

## 3. Konzept fürs Remake: "Studigame"

- **Ton**: gleiche trocken-schwarzhumorige Satire wie das Original, jetzt auf
  Studentenleben statt Obdachlosigkeit gemünzt. Universeller relatable (fast
  jeder war mal pleite im Studium), niemand fühlt sich als reale verletzliche
  Gruppe dargestellt, trotzdem bissig.
- **Setting**: fiktive Uni-Stadt **"Studistadt"** (final entschieden) statt echter
  Städte/Unis. Das **Studienfach übernimmt die Rolle des
  Klassensystems** – unterschiedliche Skill-Boni, Nebenjobs und Spielweisen
  statt eines gekauften Orts-Multiplikators. Eine leichte Campus-Orte-Facette
  (Mensa/Bibliothek/Kneipenviertel) bleibt als reines Flavor obendrauf, ohne
  eigenes Bezirks-System (kein Ertrags-/Risiko-Multiplikator, keine
  Unlock-Schwelle).
- **Name**: entschieden – **Studigame**. Liegt so nah an "Pennergame"
  (Slang-Wort + "game"), dass der Wiedererkennungswert fast von allein kommt.

### Die Kernschleife von Studigame

1. **Leergut sammeln** – nach der WG-Party. Dauer wählbar, Ergebnis wird erst
   beim Zurückkommen ausgewürfelt.
2. **Geld behalten** – der Behälter kappt das Tragen: Was nicht hineinpasst,
   fließt in den Spendentopf der Uni-Stadt (Charity-Überlauf). 🙋 Der
   Übergang von diesem virtuellen Overflow zu einer echten Spende (wer zahlt,
   Kurs, Buchhaltung/Steuer) ist vor Phase 6 (Payments) ungeklärt — siehe
   ROADMAP „Vor dem Launch", nicht per Code lösbar.
3. **Reinvestieren** – Behälter aufrüsten (Geld-Sink), später Skills, Ausrüstung,
   Wohnung.
4. **Risiko eingehen** – Hausarbeit ghostwriten, Spickzettel verkaufen,
   Notenkonkurrenz.
5. **Aufsteigen** – besseres Studiensemester, Fachwissen, Highscore.

**Der MVP-Kern (Phase 3, siehe 3b) ist Punkt 1–3** – erst wenn der Geld-Sink
trägt, kommen Skills und Risiko dazu.

### Studienfächer (Klassensystem)

Start mit 3–4 Fächern für den MVP, mehr später – jedes zusätzliche Fach ist
mehr Balance-Arbeit.

| Studienfach | Bonus | Malus/Flavor |
| --- | --- | --- |
| Informatik | Konzentration günstiger (Debug-Modus), gute Werkstudent-Nebenjobs | — |
| BWL | Größerer Geld-Container von Start, günstigeres Netzwerken | Teurere "Look"-Ausrüstung für Status nötig |
| Jura | Bonus in der Notenkonkurrenz/PvP (cutthroat, wie im echten Leben) | Teurere, längere Weiterbildungen |
| Geisteswissenschaften | Günstige, vielseitige Skills, Sozialkontakte/Musik-Bonus | Niedrigere Einkommens-Decke (selbstironischer Running Gag, kein echtes Abwerten des Fachs) |

**Wichtig für den MVP:** bis das Skillsystem steht, ist das Studienfach **reine
Identität ohne Spieleffekt** (siehe 3b). Die Boni/Mali oben werden erst mit den
Skills aktiv.

**Deferred Skills — Statusklärung:** Die MVP-Skills sind drei (Lerneffizienz,
Auftreten, Konzentration, siehe §3b). Die übrigen sechs aus der
Mechanik-Übersetzung (Bildungsstufe, Sozialkontakte, Musik, Taschendiebstahl,
ATT, DEF) sind **aufgeschoben, nicht gestrichen** — Wohnungsleiter
(WG-Zimmer→Eigentumswohnung) und die Fach-Boni (Sozialkontakte/Musik,
BWL-Netzwerken, Jura-PvP) bauen genau darauf auf. Solange sie nicht gebaut
sind, sind die zugehörigen Tabellen-/Fach-Zeilen Zukunftsziel, kein
Ist-Zustand. Die explizite Entscheidung (zurückkommen vs. endgültig weg) steht
in ROADMAP „Bewertung & Nacharbeiten" aus.

**Fachwechsel**: spätere Mechanik – kostet Zeit/Geld, resettet etwas Fortschritt,
spiegelt den echten Fachwechsel-Schmerz und liefert eine natürliche "große
Entscheidung" im Spielverlauf.

**Gegen zu schnelles Overpowered-Werden** (reales Problem im Original, siehe
1.3): Skill-Kosten überlinear skalieren AND horizontale Ziele nach dem Soft-Cap
anbieten – z. B. "Zweitstudium" als Prestige-Mechanik (kleiner permanenter
Bonus, dafür bei Null neu anfangen, evtl. anderes Studienfach), thematisch
stimmig und ein bewährtes Muster aus Idle-Games.

### Mechanik-Übersetzung

| Original | Studigame |
| --- | --- |
| Pfandflaschen sammeln | Leergut sammeln nach der WG-Party |
| Verkaufen | Automatisch beim Einsammeln (kein separater Verkaufsschritt – ein Tap) |
| Geld-Container (Becher→…→Container) | Portemonnaie → Jutebeutel → Rucksack → IKEA-FRAKTA-Tasche → **Bausparvertrag** (Insider-Witz als Endgame) |
| Bildungsstufe (Skill) | ECTS-Punkte/Semesteranzahl – das reale Uni-System als Spielmechanik |
| Angriff (ATT) | Durchsetzungsvermögen (Notenkonkurrenz, Gruppenarbeit) |
| Verteidigung (DEF) | Stressresistenz (weniger Laune-/Geldverlust) |
| Geschicklichkeit | Lerneffizienz (mehr Ertrag pro Aktion) |
| Sprechen | Präsentationstechnik (Voraussetzung für viele andere Skills) |
| Konzentration (Skill) | Bleibt "Konzentration" – beschleunigt Lernphasen vor Klausuren |
| Musik (Skill) | Straßenmusik / WG-Flur-Konzerte für Trinkgeld (passives Einkommen) |
| Taschendiebstahl (Skill) | "Rabattcode-Jäger" – Studentenrabatte maximal ausreizen |
| Sozialkontakte (Skill) | Networking (schaltet WGs/Wohnungen, spätere Banden frei) |
| Verbrechen | Hausarbeit ghostwriten, Spickzettel verkaufen, Top-Tier: Prüfungsfragen aus dem Lehrstuhl klauen |
| Passives Betteln | "Eltern um Geld bitten" (verhalten), später BAföG/Stipendium |
| Kampf/PvP | Notenkonkurrenz bei Kurven-Klausuren, Gruppenarbeit sabotieren |
| Banden | WGs – WG-Rivalität oder Fakultäten-Zoff (BWL vs. Informatik vs. Geisteswissenschaften) |
| Bandeneigentum | WG-Ausstattung: WLAN-Speed, Kaffeemaschine, Bücherregal, WG-Kasse |
| Eigenheim-Leiter | WG-Zimmer → Wohnheim-Einzelzimmer → 1-Zimmer-Wohnung → Altbauwohnung → Eigentumswohnung |
| Ehrenmitgliedschaft | "Alumni-Status"/Fördermitgliedschaft, weiterhin ohne Gameplay-Vorteil |
| Kaffee statt Promille | Temporärer Konzentrations-Boost (Kaffee), gleiche Mechanik |

**Wartungs-/Landingpage**: *"Bald verfügbar: Studigame – im Geist der
Kult-Browsergames der 2000er (u. a. Pennergame), aber ein eigenständiges, neues
Projekt ohne Verbindung zu Farbflut Entertainment."* Referenziert die
Inspiration ehrlich (rechtlich unproblematische nominative Nutzung), macht aber
klar: eigenständiges Konzept, keine Fortsetzung. Dazu ein simples
E-Mail-Waitlist-Formular.

---

## 3a. Abgleich mit dem bestehenden Code-Stand

Vor diesem Dokument existierte bereits eine Codebasis für die Pennergame-Version
(vier NRW-Städte als Editionen, Bezirke als Ertrags-/Risiko-Multiplikator).
Entscheidungen zum Übergang, kritisch geprüft:

- **Bezirk vs. Studienfach**: beides parallel (Studienfach als Klasse UND
  Campus-Ort als Multiplikator) wäre für ein Solo-MVP doppelte Balancing-Fläche.
  Entscheidung: **Studienfach ersetzt Bezirk vollständig.** Die `districts`-
  Tabelle ist entfernt, `subjects` (Studienfach) eingeführt,
  `characters.district_id` → `characters.subject_id`. Campus-Orte bleiben, wenn
  sie kommen, reine Text-Labels.
- **Editionen**: MVP startet mit **einer** fiktiven Uni-Stadt statt vier
  parallelen Welten. Die `editions`-Tabelle bleibt technisch bestehen (ein
  aktiver Eintrag), damit sich später weitere Uni-Städte ohne Schema-Änderung
  nachrüsten lassen.
- **Modulname**: `pennergame` → `studigame` (go.mod, alle Imports, DB-Dateiname,
  Backup-Skript, docker-compose). Umgesetzt – rein mechanisch und repo-weit.
- **Migrationen**: keine echten Nutzerdaten (Projekt vor dem Launch). Die
  betroffene Migrationsdatei (`0003_core_gameplay.sql`) wurde direkt angepasst
  statt additiv, `0005_open_duesseldorf.sql` gelöscht. Additive Migrationen für
  ein Schema, das nie live war, wären Ballast ohne Nutzen.

Umsetzung abgeschlossen und grün (`go test ./...`).

---

## 3b. MVP-Umfang & nächster Code-Schritt

Kritische Prüfung des aktuellen Stands (eine Aktion: Sammeln; Behälter mit
Charity-Überlauf; Studienfach bisher ohne Spieleffekt): wo ist die Wirtschaft
löchrig, was ist gerechtfertigt, was wird bewusst nicht gebaut.

**Das eine echte Loch: kein Geld-Sink.** Spieler können Geld nur einsammeln und
sich den Behälter vollaufen lassen, dann fließt der Rest an Charity. Ohne eine
Möglichkeit, Geld sinnvoll wieder auszugeben, ist der Überlauf-Mechanik jede
Spannung genommen – man grindet ins Leere. Der erste Code-Schritt ist deshalb
ein minimaler, thematischer Geld-Sink, nicht der breite Shop.

**Nächster Schritt (konkret, entschieden): Behälter-Upgrade als Geld-Sink.**

- **Kauf sofort, ein Tap, eine Transaktion.** Kein Cooldown, keine
  zeitgebundene Aktion (passives Upgrade, keine Schicht). Ledger-Debit mit
  `reason='container_upgrade'`, kein `actions`-Eintrag. UI: ein Knopf auf dem
  Dashboard, zeigt aktuelle Stufe, neue Stufe, Preis.
- **Max-Stufe 4 (Bausparvertrag)** ist eine harte Obergrenze. Die
  Upgrade-Kostenfunktion liefert bei Stufe 4 "nicht kaufbar", die UI blendet
  den Knopf aus. Danach kommt ohnehin das Skillsystem.
- **Kostenkurve** `Kosten(L→L+1) = Basiswert × (L+1)^1.8`, auf volle Euro
  gerundet, mit **Basiswert = 2000 Cent**, überlinear (anti-Schneeball, siehe
  §1.3). Konkret (Probe gegen Sammel-Einkommen 10 €/h bei Level 0):

  | Upgrade | Stufenname neu | Formel | Kosten | Kapazität neue Stufe |
  | --- | --- | --- | --- | --- |
  | L0→L1 | Jutebeutel | 2000 × 1^1.8 | 20 € | 80 € |
  | L1→L2 | Rucksack | 2000 × 2^1.8 | 70 € | 200 € |
  | L2→L3 | IKEA-FRAKTA-Tasche | 2000 × 3^1.8 | 144 € | 1000 € |
  | L3→L4 | Bausparvertrag | 2000 × 4^1.8 | 243 € | 5000 € |

  `Basiswert` ist der eine Tun-Knopf. Formal rein in `internal/game` als
  `UpgradeCostCents(level, base)` — tabellengetestet, gleiches Muster wie
  `ContainerCapacity`. Die Kapazität jeder Stufe muss die Kosten der
  nächsten Stufe decken, sonst ist der Behälter ein Softlock — Jutebeutel
  wurde deshalb von 50 € auf 80 € angehoben (`TestContainerCapacityAffordsNextUpgrade`
  prüft das für jede Stufe).
- **Danach scharfes Spielen**: erst wenn der Core (Sammeln + Behälter +
  Upgrade) sich langweilig anfühlt, kommt der nächste Block. Balancing wird
  vorher abgenommen (Test auf einem Tuner-Account).

**Entscheidungen zu diesem Block (keine offenen Punkte mehr):**

- Kostenkurve `Basiswert × Level^1.8`, Basiswert 2000 Cent ✔
- Kauf sofort + ein Tap + Ledger-Debit, kein `actions`-Eintrag ✔
- Max-Stufe 4 (Bausparvertrag), danach nicht kaufbar ✔
- Behälter-Stufen-Namen final: Portemonnaie → Jutebeutel → Rucksack →
  IKEA-FRAKTA-Tasche → Bausparvertrag ✔
- Studienfach differenziert jetzt über den Skillbaum (Kosten-/Dauer-Boni),
  siehe Skillbaum unten — nicht mehr reine Identität ✔

**Bewusst NICHT jetzt** (nicht "für später", sondern weil es den MVP sprengen
würde):

- Breiter Shop (Ausrüstung, Instrumente, Wohnung) – braucht erst Einkommens-
  quellen außer Flaschen, sonst kauft nichts zusammen.
- Skillsystem – die eigentliche Differenzierung der Studienfächer, aber groß;
  kommt als eigener Block.
- Weitere Zustands-Stats (Schlaflosigkeit, Motivation, Sauberkeit) – geparkt,
  Laune allein reicht.
- Campus-Orte als Text-Label – umgesetzt, sobald es mehrere Aktionen mit Ortsbezug gab
  (Sammeln: Mensa-Hinterhof, Job: Bibliothek, Nebenverdienst: Kneipenviertel).

**Architektur-Stand, der Bestand hat:**
Die generische `actions`-Tabelle mit `payload_json`/`result_json` und
"Aktionsergebnis erst beim Auflösen würfeln" ist der richtige Rückgrat für alle
zeitbasierten Aktionen. `ResolveDueAction` löst heute per `kind`-Dispatch auf:
ein `switch` über `collect`/`job` mit gemeinsamen Buchungspfad (Geld, Laune,
Ledger, Charity) in `bookOutcome`. Eine neue Aktion fügt eine Case im `switch`
hinzu — kein Resolver-Framework. `internal/game` bleibt stdlib-only und
infrastrukturfrei.

---

## 4. Tech-Stack (final)

| Bereich | Wahl | Warum |
| --- | --- | --- |
| Backend | Go, `net/http` | passt zu einem textbasierten, formularbasierten Spiel |
| Templating | `html/template` | typsicher, kompiliert, kein JS-Build nötig |
| Interaktivität | htmx | Server-rendered Partial-Updates statt SPA-Overhead |
| Styling | Tailwind (Standalone-CLI, kein Node-Build) | Farben als Tokens, siehe Abschnitt 5 |
| DB | SQLite (WAL-Modus) | 1 Datei, kein DB-Service; Migration auf Postgres nur falls wirklich nötig |
| Auth | Sessions statt JWT | einfacher & sicherer für diese Architektur (Begründung unten) |
| Game-Ticks | Listener in den Handlern statt Queue/Ticker | Aktionsergebnis wird beim Auflösen gewürfelt, kein Hintergrund-Ticker nötig |
| Hosting | 1 VPS (z. B. Hetzner) + Caddy (Auto-HTTPS) | kein Kubernetes |
| Payments | Stripe oder Paddle (Merchant of Record, übernimmt EU-USt.) | (erst Phase 6) |

**Zur JWT-Frage, kurz:** Für eine klassische, server-gerenderte htmx-App ist
JWT die unnötig komplizierte Wahl. JWTs lohnen sich erst bei einer zustandslosen
API für mehrere Clients (z. B. einer späteren nativen Mobile-App). Für "eine
Web-App, ein Server" ist eine simple Server-Session (bcrypt-Hash +
httpOnly-Secure-Cookie + Session-Tabelle) einfacher, sicherer und schneller
gebaut. Braucht es später eine echte externe API, lässt sich JWT gezielt
nachrüsten. Zwei Dinge dabei nicht vergessen: CSRF auf jedem htmx-POST, und
Rate-Limiting auf Login/Registrierung.

---

## 5. Corporate Design – Tokens

Ausgangspunkt: Straße, Improvisation, schwarzer Humor – kein generischer
SaaS-Look.

**Farben**

- Asphalt (dunkler Hintergrund): `#1C1B1A`
- Pfandglas-Grün (Primär/Akzent): `#7A9B57`
- Verkehrsorange (Sekundär/CTA): `#D9711F`
- Münzgold (Währung, Premium, Erfolge): `#C9A24B`
- Pappe/Papier (helle Flächen, Karten): `#E8E1D3`
- Backstein-Rot (Warnung, Angriff): `#A83E32`
- Textmarker-Gelb (Highlights/Badges/Benachrichtigungen): `#E8D630`

**Typografie**

- Display/Headlines: kräftige, leicht kondensierte Groteske (z. B. Archivo
  Black oder Barlow Condensed Bold) – wirkt wie ein Pappschild, ohne kitschig
  zu sein
- Fließtext: gut lesbare humanistische Sans (z. B. Inter oder Source Sans)
- Zahlen/Stats (Geld, ECTS): Monospace (z. B. JetBrains Mono) – gibt Werten
  einen "Ticker"-Charakter, klar abgesetzt vom Fließtext

**Layout**

- Mobile-first, eine Spalte, Bottom-Nav (Übersicht/Aktionen/Bildung/Stadt/Bande)
  statt Top-Nav – daumenfreundlich
- Sticky Status-Leiste oben (Geld, Laune, laufende Cooldowns)

**Signature-Element**
Der Geld-Behälter (Portemonnaie → … → Bausparvertrag) als durchgehend sichtbarer
Fortschritts-Indikator im Header, der sich sichtbar füllt und beim Upgrade die
Form wechselt.

---

## 6. Phasenplan

| Phase | Ziel | Status |
| --- | --- | --- |
| 0 | Landingpage/Warteliste | **erledigt** |
| 1 | Grundgerüst: Auth, DB, Deploy | **erledigt** |
| 2 | Design-System | **erledigt** |
| 3 | Kernschleife (Wirtschaft) | **läuft** (Sammeln + Behälter + Upgrade stehen, Balancing offen) |
| 4 | Sozial & PvP | offen |
| 5 | Mobile/PWA-Polish | offen |
| 6 | Monetarisierung | offen |
| 7 | Launch & Community | offen |

### Phase 0 – Landingpage/Warteliste (erledigt)

- [x] 1-Seiter, E-Mail-Feld, htmx ohne Reload (`GET /`, `POST /waitlist`)
- [x] Impressum + Datenschutzerklärung
- [x] Web-Fonts selbst hosten statt CDN (kein Cookie-Banner nötig)
- [x] Name entschieden (Studigame, Uni-Stadt Studistadt)
- [ ] 🙋 Domain registrieren – erledigt der Entwickler später beim Deploy

**Rechtliches für den Start, kurz:** Impressum und Datenschutzerklärung sind
schon für die reine Warteliste Pflicht (Gewinnerzielungsabsicht reicht) – §5 DDG
bzw. Art. 13 DSGVO. Cookie-Banner ist verzichtbar, solange keine
Tracking-/Analytics-Skripte laufen (§25 TDDDG). AGB erst ab tatsächlichen
Käufen (Phase 6). Keine Rechtsberatung, aber bei Bußgeldrahmen bis 50.000 €
lohnt ein einmaliger Check vor Go-Live.

### Phase 1 – Grundgerüst (erledigt)

- [x] SQLite-Schema (`users`, `sessions`, …)
- [x] Registrierung, Login, Logout, Passwort-Reset, Double-Opt-In,
      Studienfach-Auswahl
- [x] Session-Auth, CSRF, Rate-Limiting
- [x] Basis-Layout
- [x] Deploy-Pipeline (Docker + Caddy)

### Phase 2 – Design-System (erledigt)

- [x] Tailwind (Standalone), Tokens aus Abschnitt 5 als Config
- [x] Komponenten als handgeschriebene Tailwind-Partials (bewusst **kein
      DaisyUI**): Button, Card, Formular, Status-Leiste, Bottom-Nav, Flash, Progress
- [x] UX-Prinzip: "ein Tap = neue Aktion starten" – kein mehrstufiges Formular,
      damit ein Check-in zwischen zwei Vorlesungen in ~2 Minuten klappt

### Phase 3 – Kernschleife (läuft)

Der Umfang folgt §3b: zuerst der Geld-Sink, dann scharf spielen, dann Blöcke
dazu. Die Reihenfolge unten ist verbindlich:

- [x] **Sammeln-Aktion** mit Dauer-Wahl, Cooldown, Ergebnis erst beim
      Zurückkommen (`/actions`)
- [x] **Live-Countdown**: Alpine.js zählt die Restzeit einer laufenden Schicht
      im Browser runter und lädt beim Ablauf neu; Meta-Refresh bleibt als
      No-JS-Fallback, Ergebnis wird serverseitig beim Zurückkommen aufgelöst
- [x] **Geld-Behälter mit Kappung + Charity-Overflow** (Portemonnaie → … →
      Bausparvertrag)
- [x] **Tabellentests** für die Spielregeln in `internal/game` (Sammelertrag,
      Behältergrößen, Charity-Split, Grenzfälle)
- [x] **Behälter-Upgrade als Geld-Sink** (siehe §3b): Kauf, Ledger-Debit,
      überlineare Kostenkurve in `internal/game`, tabellengetestet
- [x] **Werkstudenten-Job als zweite Aktion**: fester Stundenlohn (kein Glück),
      Auflösung per `kind`-Dispatch (`collect`/`job`) mit gemeinsamen Buchungspfad
- [ ] **Balancing abnehmen:** Grundertrag 40 Flaschen/h, 25 ct je Flasche,
      Laune-Kosten 5 Punkte/h; Behälter-/Upgrade-Kosten liegen in §3b fest und
      werden beim ersten echten Spieldurchlauf nur noch getunt 🙋
- [ ] **Hier anhalten und spielen.** Erst wenn der Kern sich langweilig anfühlt:
- [x] **Skillbaum** (reduzierte Variante der 9 Original-Skills): drei Skills
      (Lerneffizienz, Auftreten, Konzentration), ein Slot pro Charakter (bewusst
      statt zweien), Studienfach-Boni/Mali über Kosten und Dauer; Bonus wird
      beim Aktion-Start eingefroren
- [ ] **Klausur als dritte zeitbasierte Aktion** (ECTS-Produktion) – sinnvoll
      erst mit dem Skillsystem, daher bewusst später
- [x] **BAföG** (passives Grundeinkommen, Muster wie "Instrumente") + Stipendium
      (seltener, besser, ab Bildungsstufe-Schwelle — Stipendium noch offen)
- [ ] **Shop/Campus:** Grundausstattung, Wohnung, ggf. Haustier, Kaffee
      (temporärer Konzentrations-Boost, gleiche Mechanik wie Promille im Original)
- [ ] **Verbrechen-Aktion** mit Risiko/Ertrag (Hausarbeit ghostwriten,
      Spickzettel verkaufen)

### Phase 4 – Sozial & PvP

- [ ] Kampfsystem → Notenkonkurrenz (ATT/DEF, Cooldown pro Ziel, Kampflog)
- [ ] Banden/WGs: erstellen/beitreten, gemeinsame Kasse, gemeinsamer Chat
- [ ] Freunde/Nachrichten
- [ ] Highscore (Spieler + WGs)

### Phase 5 – Mobile/PWA

- [ ] `manifest.json`, Service Worker, installierbar
- [ ] Performance-/Mobile-Pass
- [ ] Web-Push-Notifications: Empfehlung weglassen – passt zum ruhigeren
      Nostalgie-Stil, vermeidet Sucht-Mechanik, jederzeit nachrüstbar

### Phase 6 – Monetarisierung

- [ ] Kosmetik-Shop + "Unterstützer"-Status (werbefrei, kein Gameplay-Vorteil)
- [ ] Stripe/Paddle-Integration
- [ ] Spenden-Button, echte Charity-Kopplung (Overflow-Feature aus Phase 3)

### Phase 7 – Launch & Community

- [ ] Reddit: relevante Subs, transparent als Entwickler posten
- [ ] Discord: eigenen Server seit Phase 0 (auch Feedback-Kanal, kein eigenes
      Forum bauen), in Retro-/Nostalgie-Discords authentisch mitmachen
- [ ] TikTok/Reels im "erinnert ihr euch"-Format
- [ ] Flyer/Aushänge an Unis (Poster mit QR-Code am Campus)
- [ ] Referral-Link-Feature (siehe 1.5) von Anfang an einbauen – bester
      Wachstumshebel

---

## 7. Community: Alt + Neu

- **Alte Spieler**: ehrlicher "inspiriert von"-Hinweis auf der Landingpage plus
  vertraute Begriffe/Mechaniken (Pfandflaschen, Container-Progression,
  WG-Rivalitäten) lösen Wiedererkennung aus
- **Neue Spieler**: schnelles, mobile-first Onboarding, keine Verifizierung
  alter Accounts nötig – stattdessen optional ein "Warst du dabei?"-Nostalgie-
  Quiz beim Onboarding, rein für Spaß/Badge
- Früh eingeloggte Beta-Tester bekommen einen kosmetischen "Urgestein"-Badge –
  belohnt Treue ohne Pay-to-win-Charakter

## 7a. Saison-Content-Idee

Der echte Semesterkalender liefert kostenlosen Event-Content: **Klausurenphase**
als zeitlich begrenzte Hochintensitätsphase (höhere Belohnungen, höheres
Risiko), **Semesterferien** als ruhigere Phase mit mehr Nebenjob-Fokus. Später
umsetzbar, kein MVP-Bestandteil.

## 7b. Geparkt (bewusst nicht für den MVP)

- **Sauberkeit** (aus §2) – Bonus-Ressource für später, kein MVP
- **Erasmus** – guter Season-/Event-Content für später
- **i18n/Englisch** – der Humor lebt von deutschen Insidern (BAföG, Mensa,
  ECTS-Bürokratie), die auf Englisch kaum funktionieren. Einzige sinnvolle
  "Versicherung": Texte von Anfang an nicht hart codieren, sondern über eine
  simple Lookup-Funktion referenzieren (kostet fast nichts), keine zweite
  Sprache bauen
- **Diskussionsforum** – kein eigenes bauen (komplettes Extra-System); Discord
  übernimmt diese Rolle, siehe Phase 7
- **Weitere Uni-Städte als zusätzliche Editionen** – technisch vorbereitet
  (3a), erst wenn eine volle erste Stadt langweilig wird

---

## Quellen

- Offizielle Pennergame-Anleitung (Hamburg-/Berlin-/Atlantis-Edition),
  erschlossen über Suchindex
- Ausführlicher Fan-Guide, pennergame-pro.blogspot.com (2009)
- Handelsregisterauszug Farbflut Entertainment GmbH, HRB 104994
- giga.de, gutefrage.net – aktueller Seitenstatus (2026)
- Stellenanzeige Farbflut Entertainment, python-forum.de – Tech-Stack-Hinweis
