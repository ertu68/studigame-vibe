# Entwicklungsplan: Studigame (Pennergame-inspirierter Nachbau)

*Stand: 29.08.2026 – lebendes Dokument, wird während der Entwicklung laufend angepasst.*

## 1. Funktionsanalyse: Das Original

Die offizielle Anleitung (Hamburg-, Berlin-, Atlantis-Edition) und ein ausführlicher Fan-Guide von 2009 sind noch über den Suchindex erschlossen. Dadurch lässt sich das Spiel fast vollständig rekonstruieren.

### 1.1 Grundgerüst
- Start: 10 € Startkapital, Ziel: vom Pennerleben zum "Schlossbesitzer"
- Mehrere parallele "Editionen" pro Stadt (Hamburg, Berlin, Köln, München, Mallorca, dazu ein Fun-Server "Atlantis/Vatikan")
- Finanzierung nur über Spenden + "Ehrenmitgliedschaft" – Premium gab laut eigener Aussage des Originals keine echten Gameplay-Vorteile, nur Kosmetik/Komfort (eigenes Haustier-Design, größeres Bandenavatar, eigene E-Mail-Adresse, kein Werbebanner)

### 1.2 Die Kernschleife
1. **Ressourcen sammeln** – Pfandflaschen sammeln (Dauer wählbar 10 Min–12 h; kürzer = effizienter pro Zeiteinheit, aber absolute Menge steigt mit Dauer; Ertrag hängt von Stadtteil, Geschicklichkeit und tagesaktuellem Pfandpreis ab)
2. **Verkaufen/Betteln** – Flaschen verkaufen, dazu passives Bettel-/Spenden-Einkommen
3. **Investieren** – in Skills, Ausrüstung, Wohnung
4. **Risiko eingehen** – Verbrechen begehen (mehr Ertrag, aber Erwischt-Risiko) oder andere Spieler angreifen
5. **Aufsteigen** – bessere Stadtteile, Ausrüstung, Bande, Highscore-Rang

Ein Geld-"Behälter" begrenzt, wie viel du auf einmal tragen kannst (Trinkbecher → Tüte → Beutel → Einkaufswagen → Container). Wer sein Limit nicht rechtzeitig upgradet, verliert überschüssiges Geld einfach – kleine, aber richtig clevere Mechanik, zwingt zum Reinvestieren und passt perfekt zum Thema.

### 1.3 Skillsystem ("Weiterbildung")
Kosten = Basiswert × nächstes Level. Abbruch erstattet 80 %. Zwei Skills gleichzeitig in der Warteschleife möglich.

*Vermutliche Ursache fürs "schnell overpowered"-Problem: diese Kostenformel ist **linear**. Wächst das Einkommen mit dem Level mit, wird jedes weitere Level relativ günstiger statt teurer – ein Schneeball-Effekt. Fix fürs Remake: überlinear skalieren (z. B. Basiswert × Level^1.8), siehe Abschnitt 3.*

| Skill | Effekt |
|---|---|
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
Das Original hatte einen **Referral-Mechanismus direkt in der Spielfantasie eingebaut**. Jeder Spieler bekommt einen persönlichen Link (`.../change_please/{RefID}/` – "haste mal Kleingeld?"), den man in Forensignaturen oder der eigenen Profil-Homepage postet. Klickt jemand drauf, gibt es eine In-Game-Spende. Virales Wachstum, direkt an die Kernfantasie (Betteln) gekoppelt, ohne sich wie ein aufgesetzter "Invite 3 friends"-Hack anzufühlen.

### 1.6 Technischer Befund
- Backend laut einer alten Stellenanzeige von Farbflut Entertainment: **Python + MySQL**
- Die Domain läuft technisch noch, Neuregistrierung ist aber kaputt, keine Weiterentwicklung mehr
- Farbflut Entertainment GmbH (Hamburg) ist laut Handelsregister weiterhin wirtschaftsaktiv – der Markenname "verwaist" also nicht, deshalb weiterhin: eigenen Namen verwenden

---

## 2. Kuration: Übernehmen, ändern, weglassen

*Hinweis: Diese Tabelle stammt aus der ursprünglichen Mechanik-Analyse. Die thematische Verpackung ist "Studigame" statt Penner-Thema (siehe Abschnitt 3) – die Entscheidungen zur Mechanik selbst gelten unverändert, mit einer Ausnahme (Bezirk/Stadtteil, siehe 3a).*

| Element | Entscheidung | Warum |
|---|---|---|
| Grundschleife (sammeln → verkaufen → investieren → Risiko → aufsteigen) | **Übernehmen** | Zeitlos, funktioniert einfach |
| Geld-Behälter mit Kappung | **Übernehmen, leicht verändert** | Statt "Geld weg für immer" bei Überlauf: automatische Spende an eine echte Obdachlosenhilfe/Studierendenwerk. Gleiche Spielspannung, fühlt sich aber gut statt frustrierend an |
| Promille-System | **Übernehmen** | Passt im Studenten-Kontext (Vorglühen, WG-Partys) organischer als im Original – braucht weniger Abschwächung, aber weiterhin mit Augenmaß umgesetzt |
| Sauberkeit sinkt durchs Sammeln, erhöht Spenden | **Übernehmen** | Nette Nebenressource, unproblematisch |
| Referral-Link als "Spendenlink" | **Übernehmen & ausbauen** | Genialer eingebauter Wachstumsmotor, siehe 1.5 |
| Ehrenmitgliedschaft ohne Gameplay-Vorteil | **Übernehmen** | Deckt sich mit der Monetarisierungs-Empfehlung – hat sich schon damals bewährt |
| Reale Städte/Stadtteile (Hamburg, Berlin, St. Pauli, …) | **Ändern → fiktive Stadt, Bezirk als Multiplikator fällt weg** | Weniger Nähe zum Original, mehr kreative Freiheit; Rolle des Multiplikators übernimmt das Studienfach, siehe 3a |
| Shock-Humor bei manchen Item-/Haustiernamen | **Weglassen** | Harter, dated 2000er-Forenhumor, der schon damals grenzwertig war |
| "Dein Leben beenden" als Label für Account löschen | **Ändern** | Klingt nach Selbstverletzung, auch wenn nur der Charakter gemeint ist – einfach "Account löschen" nennen |
| 36h-Cooldown pro Angriffsziel | **Übernehmen, Werte tunen** | Guter Anti-Griefing-Mechanismus |
| Kein direkter Geldtransfer (nur Bandenkasse/Haustierkämpfe) | **Übernehmen** | Verhindert simples Farming/RMT |
| Textbasiert, schlichtes UI | **Übernehmen, modernisieren** | Gleicher Minimalismus, aber mobile-first & schnell |

---

## 3. Konzept fürs Remake: "Studigame"

- **Ton**: gleiche trocken-schwarzhumorige Satire wie das Original, jetzt auf Studentenleben statt Obdachlosigkeit gemünzt – universeller relatable (fast jeder war mal pleite im Studium), niemand fühlt sich als reale verletzliche Gruppe dargestellt, trotzdem bissig
- **Setting**: fiktive Uni-Stadt statt echter Städte/Unis. Statt der alten Stadt/Bezirk-Mechanik (reiner Ressourcen-Multiplikator) übernimmt **Studienfach die Rolle des Klassensystems** – unterschiedliche Skill-Boni, Nebenjobs und Spielweisen statt eines gekauften Orts-Multiplikators. Eine leichte Campus-Orte-Facette (Mensa/Bibliothek/Kneipenviertel) bleibt als reines Flavor obendrauf, ohne eigenes Bezirks-System (kein Ertrags-/Risiko-Multiplikator, keine eigene Unlock-Schwelle)
- **Name**: entschieden – **Studigame**. Liegt so nah an "Pennergame" (Slang-Wort + "game"), dass der Wiedererkennungswert fast von allein kommt

### Studienfächer (Klassensystem)

Start mit 3–4 Fächern für den MVP, mehr später – jedes zusätzliche Fach ist mehr Balance-Arbeit:

| Studienfach | Bonus | Malus/Flavor |
|---|---|---|
| Informatik | Konzentration günstiger (Debug-Modus), gute Werkstudent-Nebenjobs | — |
| BWL | Größerer Geld-Container von Start, günstigeres Netzwerken | Teurere "Look"-Ausrüstung für Status nötig |
| Jura | Bonus in der Notenkonkurrenz/PvP (cutthroat, wie im echten Leben) | Teurere, längere Weiterbildungen |
| Geisteswissenschaften | Günstige, vielseitige Skills, Sozialkontakte/Musik-Bonus | Niedrigere Einkommens-Decke (selbstironischer Running Gag, kein echtes Abwerten des Fachs) |

**Fachwechsel**: spätere Mechanik – kostet Zeit/Geld, resettet etwas Fortschritt, spiegelt den echten Fachwechsel-Schmerz und liefert eine natürliche "große Entscheidung" im Spielverlauf.

**Gegen zu schnelles Overpowered-Werden** (reales Problem im Original, siehe 1.3): Skill-Kosten überlinear skalieren UND horizontale Ziele nach dem Soft-Cap anbieten – z. B. "Zweitstudium" als Prestige-Mechanik (kleiner permanenter Bonus, dafür bei Null neu anfangen, evtl. anderes Studienfach), thematisch stimmig und ein bewährtes Muster aus Idle-Games.

### Mechanik-Übersetzung

| Original | Studigame |
|---|---|
| Pfandflaschen sammeln | Bleibt fast wörtlich – Leergut sammeln nach der WG-Party |
| Geld-Container (Becher→...→Container) | Portemonnaie → Jutebeutel → Rucksack → IKEA-FRAKTA-Tasche → **Bausparvertrag** (Insider-Witz als Endgame: du bist erwachsen geworden) |
| Bildungsstufe (Skill) | Wörtlich ECTS-Punkte/Semesteranzahl – das reale Uni-System als Spielmechanik |
| Konzentration (Skill) | Bleibt "Konzentration" – beschleunigt Lernphasen vor Klausuren |
| Musik (Skill) | Straßenmusik/WG-Flur-Konzerte für Trinkgeld |
| Taschendiebstahl (Skill) | "Rabattcode-Jäger" – Studentenrabatte maximal ausreizen |
| Verbrechen | Hausarbeit ghostwriten, Spickzettel verkaufen, Top-Tier: Prüfungsfragen aus dem Lehrstuhl klauen |
| Kampf/PvP | Notenkonkurrenz bei Kurven-Klausuren, Gruppenarbeit sabotieren |
| Banden | WGs – Bandenkrieg wird WG-Rivalität oder Fakultäten-Zoff (BWLer vs. Informatiker vs. Geisteswissenschaftler) |
| Bandeneigentum | WG-Ausstattung: WLAN-Speed, Kaffeemaschine, Bücherregal, WG-Kasse |
| Eigenheim-Leiter | WG-Zimmer → Wohnheim-Einzelzimmer → 1-Zimmer-Wohnung → Altbauwohnung → Eigentumswohnung |
| Ehrenmitgliedschaft | "Alumni-Status"/Fördermitgliedschaft, weiterhin ohne Gameplay-Vorteil |

- **Wartungs-/Landingpage**: *"Bald verfügbar: Studigame – im Geist der Kult-Browsergames der 2000er (u. a. Pennergame), aber ein eigenständiges, neues Projekt ohne Verbindung zu Farbflut Entertainment."* Referenziert die Inspiration ehrlich (rechtlich unproblematische nominative Nutzung), macht aber klar: eigenständiges Konzept, keine Fortsetzung. Dazu ein simples E-Mail-Waitlist-Formular.

---

## 3a. Abgleich mit dem bestehenden Code-Stand

Vor diesem Dokument existierte bereits eine Codebasis für die Pennergame-Version (vier NRW-Städte als Editionen, Bezirke als Ertrags-/Risiko-Multiplikator). Entscheidungen zum Übergang, kritisch geprüft und mit dir abgestimmt:

- **Bezirk vs. Studienfach**: beide Systeme parallel zu betreiben (Studienfach als Klasse UND Campus-Ort als zusätzlicher Multiplikator) wäre für ein Solo-MVP doppelte Balancing-Fläche für ein Feature, das dieses Dokument selbst schon als reines Flavor einstuft. Entscheidung: **Studienfach ersetzt Bezirk vollständig.** Die bestehende `districts`-Tabelle (`yield_factor`, `risk_factor`, `unlock_at_cents`, `distance_rank`) fällt weg. Campus-Orte (Mensa/Bibliothek/Kneipenviertel) werden zu reinen Text-Labels ohne eigenen Datensatz mit Werten.
- **Editionen**: MVP startet mit **einer** fiktiven Uni-Stadt statt vier parallelen Welten – deckt sich mit der bisherigen Roadmap-Begründung ("eine volle Stadt fühlt sich besser an als eine geteilte Spielerbasis"). Die `editions`-Tabelle bleibt technisch bestehen (ein aktiver Eintrag), damit sich später weitere Uni-Städte ohne Schema-Änderung nachrüsten lassen.
- **Modulname**: `pennergame` → `studigame`, betrifft `go.mod`, alle Imports, Docker-Image- und Binary-Namen. Rein mechanischer, aber repo-weiter Rename – eigener Schritt mit eigenem Umsetzungsplan, nicht Teil dieser Dokumentationsrunde.
- **Migrationen**: Es existieren noch keine echten Nutzerdaten (Projekt ist vor dem Launch). Die betroffenen Migrationsdateien (`0003_core_gameplay.sql`, `0005_open_duesseldorf.sql`) werden beim Umbau direkt angepasst statt additiv überschrieben – additive Migrationen für ein Schema, das nie live war, wären Ballast ohne Nutzen.

Konkrete Umsetzungsschritte stehen als Todo-Block in `ROADMAP.md`, werden aber erst über einen eigenen Umsetzungsplan angegangen (siehe `CLAUDE.md`, Abschnitt Vorgehen).

---

## 4. Tech-Stack (final)

| Bereich | Wahl | Warum |
|---|---|---|
| Backend | Go, `net/http` (1.22+) oder `chi` | passt zu einem textbasierten, formularbasierten Spiel |
| Templating | `templ` oder `html/template` | typsicher, kompiliert, kein JS-Build nötig |
| Interaktivität | htmx | Server-rendered Partial-Updates statt SPA-Overhead |
| Styling | Tailwind (Standalone-CLI, kein Node-Build) | Primärfarben & Co. als Tokens, siehe Abschnitt 5 |
| DB | SQLite (WAL-Modus) zum Start | 1 Datei, kein DB-Service zu pflegen; Migration auf Postgres nur falls wirklich nötig |
| Auth | **Sessions statt JWT** (Begründung unten) | einfacher & sicherer für diese Architektur |
| Game-Ticks | goroutines + `time.Ticker` | kein Queue-System nötig |
| Hosting | 1 VPS (z. B. Hetzner) + Caddy (Auto-HTTPS) | kein Kubernetes, auch wenn's aus dem Job vertraut ist |
| Payments | Stripe oder Paddle (Merchant of Record, übernimmt EU-USt.) | |

**Zur JWT-Frage, kurz:** Für eine klassische, server-gerenderte htmx-App ist JWT die unnötig komplizierte Wahl. JWTs lohnen sich, wenn eine zustandslose API für mehrere Clients gebraucht wird (z. B. später eine native Mobile-App). Für "eine Web-App, ein Server" ist eine simple Server-Session (bcrypt-Hash + httpOnly-Secure-Cookie + Session-Tabelle in der DB) einfacher, sicherer (kein Token-Diebstahl über localStorage, kein Revoke-Problem vor Ablauf) und schneller gebaut. Braucht es später eine echte externe API, lässt sich JWT gezielt für diesen Teil nachrüsten. Zwei Dinge dabei nicht vergessen: CSRF-Tokens auf jedem htmx-POST-Formular, und Rate-Limiting auf Login/Registrierung.

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
- Textmarker-Gelb (Highlights/Badges/Benachrichtigungen): `#E8D630` – sofort erkennbares "Studieren"-Signal

**Typografie**
- Display/Headlines: kräftige, leicht kondensierte Groteske (z. B. Archivo Black oder Barlow Condensed Bold) – wirkt wie ein Pappschild, ohne kitschig zu sein
- Fließtext: gut lesbare humanistische Sans (z. B. Inter oder Source Sans)
- Zahlen/Stats (Geld, ATT/DEF): Monospace (z. B. JetBrains Mono) – gibt Werten einen "Ticker"-Charakter, klar abgesetzt vom Fließtext

**Layout**
- Mobile-first, eine Spalte, Bottom-Nav mit den Hauptbereichen (Übersicht/Aktionen/Stadt/Bande/Kampf) statt Top-Nav – daumenfreundlich
- Sticky Status-Leiste oben (Geld, Laune, laufende Cooldowns)

**Signature-Element**
Der Geld-Behälter (Portemonnaie → Jutebeutel → Rucksack → IKEA-FRAKTA-Tasche → Bausparvertrag) als durchgehend sichtbares Fortschritts-Icon im Header, das sich sichtbar füllt und bei Upgrade die Form wechselt.

---

## 6. Phasenplan

| Phase | Ziel | Grobe Dauer* |
|---|---|---|
| 0 | Landingpage/Warteliste live | 1 Woche |
| 1 | Grundgerüst: Auth, DB, Deploy-Pipeline | 2–3 Wochen |
| 2 | Design-System umsetzen | 1 Woche |
| 3 | Kernschleife (Wirtschaft) | 3–4 Wochen |
| 4 | Sozial & PvP | 3–4 Wochen |
| 5 | Mobile/PWA-Polish, Notification-Entscheidung | 1–2 Wochen |
| 6 | Monetarisierung | 1 Woche |
| 7 | Launch & Community | laufend |

*Richtwerte bei ein paar Stunden/Woche nebenbei.

### Phase 0 – Landingpage/Warteliste
- [ ] Domain + Name final
- [ ] 1-Seiter: Teaser-Text (Formulierung siehe Abschnitt 3), E-Mail-Feld, Absenden per htmx ohne Reload
- [ ] Impressum (über Adressdienst/virtuelle Geschäftsadresse statt Privatadresse) + Datenschutzerklärung (Generator wie e-recht24.de reicht für den einfachen Fall) – beide Pflicht, siehe Erklärung unten
- [ ] Web-Fonts selbst hosten statt von Google-CDN laden (bekannte Abmahnfalle, gleichzeitig hält's dich im "kein Cookie-Banner nötig"-Bereich)
- [ ] Auf relevanten Kanälen posten, Reaktionen/Anmeldungen als Realitätscheck beobachten
- [ ] Öffentliche Roadmap (simples GitHub-Projects-Board reicht) verlinkt von der Landingpage – kostenlos und gutes "Build in Public"-Material

**Rechtliches für den Start, kurz:** Impressum und Datenschutzerklärung sind schon für die reine Warteliste Pflicht (Gewinnerzielungsabsicht reicht, muss noch kein Umsatz da sein) – §5 DDG (seit Mai 2024, löste die alte TMG-Regel ab) bzw. Art. 13 DSGVO. Cookie-Banner ist dagegen verzichtbar, solange keine Tracking-/Analytics-Skripte laufen (§25 TDDDG, vormals TTDSG). AGB sind keine gesetzliche Pflicht und erst relevant, sobald tatsächlich Verträge/Käufe stattfinden – frühestens Phase 6. Keine Rechtsberatung, aber bei Bußgeldrahmen bis 50.000€ lohnt ein einmaliger Check vor Go-Live.

### Phase 1 – Grundgerüst
- [ ] Repo, Go-Module, Ordnerstruktur
- [ ] SQLite-Schema: `users`, `sessions` – DB-Datei nicht ins Git-Repo (`.gitignore`), lokaler Dev-Pfad getrennt vom Produktions-Pfad auf dem VPS
- [ ] Registrierung (Name, E-Mail, Passwort, Bestätigungsmail, Studienfach-Auswahl), Login, Logout, Passwort-Reset
- [ ] Session-Auth: bcrypt-Hash, httpOnly-Secure-Cookie, CSRF-Schutz, Rate-Limiting auf Login
- [ ] Basis-Layout-Template (noch ohne finales Design), Grundnavigation
- [ ] Deploy-Pipeline (VPS + Caddy + systemd/Docker) einmal komplett durchspielen, bevor Feature-Arbeit beginnt

### Phase 2 – Design-System
- [ ] Tailwind einbinden (Standalone-CLI), Tokens aus Abschnitt 5 als Config
- [ ] Basis-Komponenten: Button, Card, Formular, Status-Leiste
- [ ] UX-Prinzip: "ein Tap = neue Aktion starten" – kein mehrstufiges Formular, damit ein Check-in zwischen zwei Vorlesungen wirklich in ~2 Minuten klappt
- [ ] Landingpage + Login/Registrierung im finalen Design

### Phase 3 – Kernschleife
- [ ] Flaschensammeln-Äquivalent (Aktion mit Dauer-Wahl, Cooldown, Ertrag nach Studienfach/Skill)
- [ ] Timer-Anzeige ohne Reload: Server sendet Ziel-Zeitstempel, erst mit reinem Vanilla-JS lösen (Alpine.js als Nachrüst-Option, falls später mehr lokaler State gebraucht wird); htmx-Polling (`hx-trigger="every 5s"`) oder SSE-Extension für Server-seitige Änderungen (neue Nachricht, Angriff)
- [ ] Geld-Behälter-System mit Kappung + Charity-Overflow
- [ ] Skillbaum (reduzierte Variante der 9 Original-Skills) mit Studienfach-Boni/Mali (3–4 Fächer für den Start)
- [ ] Shop/Campus: Grundausstattung, Wohnung, ggf. Haustier, Kaffee (temporärer Konzentrations-Boost, wie Alkohol im Original)
- [ ] BAföG (passives Grundeinkommen, gleiches Muster wie alte "Instrumente") + Stipendium (seltener, besser, ab Bildungsstufe-Schwelle)
- [ ] Verbrechen-Aktion mit Risiko/Ertrag

### Phase 4 – Sozial & PvP
- [ ] Kampfsystem (ATT/DEF, Cooldown pro Ziel, Kampflog)
- [ ] Banden/WGs: erstellen/beitreten, gemeinsame Kasse, gemeinsamer Chat
- [ ] Freunde/Nachrichten
- [ ] Highscore (Spieler + Banden)

### Phase 5 – Mobile/PWA
- [ ] `manifest.json`, Service Worker, installierbar
- [ ] Performance-/Mobile-Pass
- [ ] Entscheidung Web-Push-Notifications: Empfehlung erstmal weglassen – passt zum ruhigeren Nostalgie-Stil, vermeidet Sucht-Mechanik, jederzeit nachrüstbar

### Phase 6 – Monetarisierung
- [ ] Kosmetik-Shop + "Unterstützer"-Status (werbefrei, kein Gameplay-Vorteil – wie im Original)
- [ ] Stripe/Paddle-Integration
- [ ] Spenden-Button, echte Charity-Kopplung (Overflow-Feature aus Phase 3)

### Phase 7 – Launch & Community
- [ ] Reddit: relevante Subs, immer transparent als Entwickler posten, Selbstpromo-Regeln vorher lesen
- [ ] Discord: eigenen Server ab Phase 0 aufziehen (dient auch als Feedback-/Diskussionskanal – kein eigenes Forum bauen), zusätzlich in bestehenden Retro-/Nostalgie-Discords authentisch mitmachen statt nur Link zu posten
- [ ] TikTok/Reels im "erinnert ihr euch"-Format
- [ ] Flyer/Aushänge an Unis – ähnliches Playbook wie früher bei Jodel (Poster mit QR-Code direkt am Campus), passt jetzt thematisch perfekt
- [ ] Referral-Link-Feature (siehe 1.5) von Anfang an einbauen – bester Wachstumshebel, weil er direkt aus der Spielidee kommt

---

## 7. Community: Alt + Neu

- **Alte Spieler**: der ehrliche "inspiriert von"-Hinweis auf der Landingpage plus vertraute Begriffe/Mechaniken (Pfandflaschen, Container-Progression, Bandenkriege) lösen Wiedererkennung aus
- **Neue Spieler**: schnelles, mobile-first Onboarding, keine Verifizierung alter Accounts nötig (technisch eh nicht möglich) – stattdessen optional ein "Warst du dabei?"-Nostalgie-Quiz beim Onboarding, rein für Spaß/Badge, keine echte Prüfung
- Früh eingeloggte Beta-Tester bekommen einen kosmetischen "Urgestein"-Badge – belohnt Treue ohne Pay-to-win-Charakter

---

## 7a. Saison-Content-Idee

Der echte Semesterkalender liefert kostenlosen Event-Content, ohne dass eigene Inhalte erfunden werden müssen: **Klausurenphase** als zeitlich begrenzte Hochintensitätsphase (höhere Belohnungen, höheres Risiko), **Semesterferien** als ruhigere Phase mit mehr Nebenjob-Fokus. Später umsetzbar, kein MVP-Bestandteil.

## 7b. Geparkt (bewusst nicht für den MVP)
- **Erasmus**: guter Season-/Event-Content für später (z. B. zeitlich begrenztes "Erasmus-Semester" mit besonderen Items), nichts für den Start
- **i18n/Englisch**: der Humor lebt von deutschen Insidern (BAföG, Mensa, ECTS-Bürokratie), die auf Englisch kaum funktionieren – und die deutsche Zielgruppe ist noch nicht mal validiert. Einzige sinnvolle "Versicherung": Texte von Anfang an nicht hart codieren, sondern über eine simple Lookup-Funktion referenzieren (kostet fast nichts), eine echte zweite Sprache aber nicht bauen
- **Diskussionsforum**: kein eigenes bauen (komplettes Extra-System mit Threads/Mod-Tools) – Discord übernimmt diese Rolle bereits, siehe Phase 7
- **Weitere Uni-Städte als zusätzliche Editionen**: technisch vorbereitet (siehe 3a), aber erst wenn eine volle erste Stadt langweilig wird

---

## Quellen
- Offizielle Pennergame-Anleitung (Hamburg-/Berlin-/Atlantis-Edition), erschlossen über Suchindex
- Ausführlicher Fan-Guide, pennergame-pro.blogspot.com (2009)
- Handelsregisterauszug Farbflut Entertainment GmbH, HRB 104994
- giga.de, gutefrage.net – aktueller Seitenstatus (2026)
- Stellenanzeige Farbflut Entertainment, python-forum.de – Tech-Stack-Hinweis
