# Arbeitsanweisungen für Coding-Agents

Browser-Game im Stil von Pennergame. Go, server-rendered, SQLite, kein JS-Build.
Konzept und Phasenplan: [pennergame-remake-plan.md](./pennergame-remake-plan.md).
Was als Nächstes ansteht: [ROADMAP.md](./ROADMAP.md).

## Grundhaltung

Dies ist ein MVP eines Solo-Entwicklers. Der teuerste Fehler ist nicht fehlende
Abstraktion, sondern zu viel davon. Im Zweifel: die kürzere Lösung.

- Kein Interface mit einer Implementierung. Konkreter Typ, bis es einen zweiten gibt.
- Keine neue Dependency für etwas, das die stdlib kann.
- Keine Konfigurationsoption für einen Wert, der sich nie ändert.
- Kein Code „für später“. Später kann sich selbst bedienen.
- Löschen ist eine Verbesserung.

## Nach jeder Aufgabe

Diese vier Schritte gehören zur Aufgabe, nicht danach:

1. `go build ./... && go vet ./... && go test ./...` — muss grün sein
2. **`README.md` aktualisieren**, sobald sich Struktur, Kommandos, Env-Vars oder
   der Funktionsumfang geändert haben. Die README ist die Dokumentation und
   veraltet sonst binnen einer Woche.
3. **`CHANGELOG.md`**: Eintrag unter `[Unreleased]` in der passenden Rubrik
   (Hinzugefügt / Geändert / Behoben / Entfernt). Ein Satz, aus Nutzersicht,
   kein Commit-Log.
4. **`ROADMAP.md`**: erledigte Punkte abhaken, neu entstandene ergänzen.

Beim Abschluss einer ganzen Phase zusätzlich: den Abschnitt „Stand“ in der README
auf die neue Phase heben und im Plan den Status-Check (Abschnitt 0) nachziehen.

## Architektur-Regeln

Ausführlich in Abschnitt 6 des Plans. Die Kurzfassung:

| Regel | Warum |
| --- | --- |
| `internal/game` importiert **nur stdlib** — keine DB, kein `time.Now()`, kein HTTP | die Wirtschaft muss ohne Infrastruktur testbar sein |
| Zeit und Zufall werden hineingereicht, nie selbst geholt | reproduzierbare Tests, keine Flakes |
| Geld ausschließlich über `ledger.Post(...)` | Nachvollziehbarkeit, macht Exploits sichtbar |
| Geld immer `int` in **Cent**, nie Float | `0.1 + 0.2` |
| Eine Spieleraktion = eine DB-Transaktion | kein halb-committeter Zustand |
| Handler bleiben dünn: parsen, Aktion rufen, rendern | keine Spielregeln in HTTP-Code |
| Aktionsergebnis erst **beim Auflösen** würfeln | sonst steht der Ausgang vorab in der DB |
| Jede Query auf andere Spieler filtert `edition_id` | Städte sind getrennte Welten |

## Sicherheit — nicht verhandelbar

- Nutzereingaben nie mit `template.HTML` ausgeben. Auto-Escaping bleibt an.
- Jeder POST braucht CSRF-Prüfung, jede Aktionsroute ein Rate-Limit.
- Kein Ergebnis, kein Preis, kein Betrag aus dem Formular übernehmen — Server rechnet.
- Fehlermeldungen bei Login und Passwort-Reset dürfen nicht verraten, ob eine
  E-Mail-Adresse existiert.

## Tests

Details in Abschnitt 6.8 des Plans.

- Schwerpunkt sind Tabellentests für `internal/game`. Dort verstecken sich
  Balancing-Bugs und Exploits.
- Aktionen: ein Test gegen `:memory:`-SQLite pro Aktion, Happy Path plus die
  kritische Grenze (Doppelauflösung, Überlauf, Cooldown).
- Eine `httptest`-Strecke für Auth und eine für die Spielschleife. Geprüft wird
  Verhalten, nicht Markup.
- Nur stdlib `testing`. Kein testify, kein Mock-Framework, kein Browser-Runner.
- Neue Logik mit einer Verzweigung bekommt einen Test. Ein Einzeiler nicht.

## Sprache

- Alles, was ein Spieler sieht: **Deutsch**, Duz-Form, trocken.
- Code, Bezeichner, Commit-Messages, Kommentare: **Englisch**.
- Dokumentation im Repo: Deutsch.
- Ton: sachlich. Keine Ausrufezeichen, keine Emoji in der Oberfläche.

## Commits

Eine Zeile Betreff im Imperativ, ohne Präfix-Zeremonie. Body nur, wenn das Warum
nicht offensichtlich ist. Nicht committen, ohne dass die Tests grün sind.

## Sensibler Inhalt

Das Spiel handelt von Obdachlosigkeit. Der Ton ist Satire auf Aufstiegsträume,
nicht auf Betroffene. Bezirke sind reale NRW-Stadtteile: sie unterscheiden sich
über Zahlen (Ertrag, Risiko, Freischaltschwelle), nie über abwertende
Beschreibungen. Faustregel für jeden Text: würde ich ihn der lokalen
Stadtverwaltung zeigen?

## Wenn etwas unklar ist

Nachfragen statt raten — besonders bei Balancing-Zahlen, Namensgebung, Kosten
verursachenden Entscheidungen und allem mit Rechtsbezug (Impressum,
Datenschutz, Zahlungen). Der Plan markiert solche Punkte mit 🙋.
