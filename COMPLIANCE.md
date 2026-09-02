VERDICT: CHANGES_REQUESTED

## Bericht zur Rechts- und Marktkonformität

Projekttyp: `go-backend` — reine REST-API ohne Endnutzer-UI.  
Pflichtbereiche: DSGVO (Verarbeitung, Logs), EU Cyber Resilience Act (CRA).  
Nicht einschlägig: EU AI Act (keine KI-Funktion), Cookie-/Impressumspflichten und Web-Accessibility (kein öffentliches Web-UI).

---

### 1. DSGVO

#### Befund 1.1 — Transportverschlüsselung fehlt
- **Schweregrad:** hoch
- **Datei:** `main.go`
- **Sachverhalt:** Der Server startet ausschließlich über `http.ListenAndServe`; Paste-Inhalte (potenziell personenbezogene Daten) würden im Klartext übertragen. Die Standardadresse ist zwar `localhost:8080`, aber über `SERVER_ADDR` kann der Dienst ohne TLS öffentlich exponiert werden. Art. 32 DSGVO verlangt geeignete technische Maßnahmen, typischerweise Transportverschlüsselung.
- **Maßnahme:** TLS im Code unterstützen, z. B. `http.ListenAndServeTLS` mit konfigurierbaren Zertifikatspfaden. Alternativ verbindlich in `README.md` dokumentieren, dass der Dienst ausschließlich hinter einem TLS-terminierenden Reverse Proxy betrieben werden darf, und im Code eine Schutzprüfung ergänzen, die den Betrieb ohne TLS außerhalb von `localhost` verhindert oder deutlich warnt.

#### Befund 1.2 — Unbegrenzte Speicherung bei fehlendem Ablauf
- **Schweregrad:** mittel
- **Datei:** `internal/store/store.go`, `internal/api/create.go`
- **Sachverhalt:** `expires_in_seconds` ist optional. Ohne Wert bleibt der Paste bis zum Prozessende bzw. Neustart gespeichert. Das ist potenziell eine unbegrenzte Speicherung personenbezogener Inhalte ohne dokumentierte Erforderlichkeit; das Prinzip der Speicherbegrenzung (Art. 5 Abs. 1 lit. e DSGVO) ist nicht technisch abgesichert.
- **Maßnahme:** Einen Standardablauf oder eine maximale Lebensdauer einführen (z. B. Konfigurationswerte `DEFAULT_EXPIRES_IN_SECONDS` und `MAX_EXPIRES_IN_SECONDS`). Falls bewusst unbegrenzt, in `README.md` und einer Datenschutzdokumentation begründen und eine Löschroutine für langlaufende Prozesse vorsehen.

#### Befund 1.3 — Datenschutzhinweise / Rechtsgrundlage nicht sichtbar
- **Schweregrad:** mittel
- **Datei:** `README.md` (nicht im Detail gezeigt, aber vorhanden)
- **Sachverhalt:** Die API verarbeitet freiwillig übermittelte Paste-Inhalte. Aus dem sichtbaren Code ist keine Rechtsgrundlage und keine Information nach Art. 13 DSGVO erkennbar. Für ein reines Backend ist keine Web-Datenschutzerklärung nötig, aber die Informationspflichten gegenüber den Nutzern/API-Kunden bleiben bestehen.
- **Maßnahme:** In `README.md` einen Abschnitt „Datenschutz & Rechtsgrundlage“ ergänzen: Verarbeitung zwecks Bereitstellung des Dienstes auf Anforderung (Art. 6 Abs. 1 lit. b bzw. berechtigtes Interesse nach lit. f), keine Protokollierung von Inhalten, Löschmöglichkeit über `DELETE /pastes/{id}`, Hinweis auf die optionale Ablaufsteuerung und auf verbleibende Rechte (Auskunft, Berichtigung, Löschung). Die API-Endpunkte `GET /pastes/{id}` und `DELETE /pastes/{id}` als technische Umsetzung der Betroffenenrechte dokumentieren.

#### Befund 1.4 — Protokollierung und Fehlerbehandlung
- **Schweregrad:** niedrig (positiv)
- **Datei:** `main.go`, `internal/api/*`
- **Sachverhalt:** Es werden keine Paste-Inhalte protokolliert. 500-Fehler enthalten nur generische Meldungen, keine internen Fehlerdetails oder Inhalte. IDs sind kryptografisch zufällig (`crypto/rand`, 16 Byte) und nicht aufzählbar. Diese Maßnahmen entsprechen den Vorgaben aus AC-13 bis AC-15.
- **Maßnahme:** Beibehalten. Optional ein strukturiertes Access-Log ohne Body-Inhalte ergänzen und darauf achten, dass auch zukünftige Logs keine `Content`-Werte enthalten.

---

### 2. EU Cyber Resilience Act (CRA)

#### Befund 2.1 — Fehlende HTTP-Timeouts
- **Schweregrad:** hoch
- **Datei:** `main.go`
- **Sachverhalt:** `http.ListenAndServe` verwendet den Default-Server ohne `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout` oder `IdleTimeout`. Das erleichtert langsame Angriffe (z. B. Slowloris) und unbegrenzt offene Verbindungen. Security by design/default (CRA) ist nicht vollständig umgesetzt.
- **Maßnahme:** Einen expliziten `http.Server` konfigurieren:
  ```go
  srv := &http.Server{
      Addr:              addr,
      Handler:           newMux(),
      ReadHeaderTimeout: 5 * time.Second,
      ReadTimeout:       15 * time.Second,
      WriteTimeout:      15 * time.Second,
      IdleTimeout:       60 * time.Second,
  }
  ```
  und `srv.ListenAndServe()` verwenden.

#### Befund 2.2 — SBOM- und Update-Dokumentation nicht sichtbar
- **Schweregrad:** mittel
- **Datei:** `README.md`, optional neues `SECURITY.md` oder `SBOM.md`
- **Sachverhalt:** Die CRA verlangt für Produkte mit digitalen Elementen dokumentierte Sicherheitseigenschaften, eine Software-Stückliste (SBOM) und eine Update-/Patch-Fähigkeit. `go.mod` enthält sichtbar keine Drittanbieter-Abhängigkeiten, aber eine explizite SBOM und ein dokumentierter Update-Prozess fehlen.
- **Maßnahme:** In `README.md` oder `SECURITY.md` dokumentieren: unterstützte Sicherheitsarchitektur, Update-Weg (Neubereitstellung des Binaries/Containers), Abhängigkeitsstatus (derzeit keine externen Module) und Prozess zur Einspielung von Sicherheitsupdates. Optional SBOM mit Standardwerkzeugen erzeugen und im Repo ablegen.

#### Befund 2.3 — Security-by-Design-Maßnahmen vorhanden
- **Schweregrad:** niedrig (positiv)
- **Datei:** `internal/api/create.go`, `internal/api/helpers.go`, `internal/store/store.go`
- **Sachverhalt:** Body-Größenbegrenzung (`MaxBytesReader`, 1 MiB), Content-Type-Prüfung (`application/json`), Validierung des ID-Formats (32 Hex-Zeichen), sichere Zufalls-IDs und generische Fehlermeldungen sind solide Security-by-Design-Maßnahmen.
- **Maßnahme:** Beibehalten und in der Sicherheitsdokumentation als bewusste Sicherheitsentscheidungen aufführen.

---

### 3. EU AI Act

Nicht anwendbar: Die Codebasis enthält keine KI-Funktion im Sinne des AI Act.

---

### 4. Pflichttexte und Web-UI

Nicht anwendbar: Reines Backend ohne öffentliche Endnutzer-UI. Impressum, Cookie-Banner, AGB und Accessibility-Pflichten nach WCAG/BITV/EAA entfallen.

---

### 5. Zusammenfassung

Keine fundamentalen Rechtsverstöße, die eine sofortige Blockade erfordern. Die Verarbeitung erfolgt zweckgebunden auf Anfrage und ohne Protokollierung sensibler Inhalte. Es bestehen jedoch behebbare Lücken bei Transportverschlüsselung, Speicherbegrenzung und CRA-Dokumentationspflichten. Die geforderten Maßnahmen sind konkret umsetzbar und brechen keine bestehende Produktfunktion.