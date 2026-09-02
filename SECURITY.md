VERDICT: CHANGES_REQUESTED

## Security-Report

### Zusammenfassung
Es wurden keine kritischen oder hohen Schwachstellen wie hartkodierte Secrets, Injection/RCE, Auth-Bypass oder unsichere Abhängigkeiten gefunden. Der Code setzt die geforderten Sicherheitsanforderungen (Body-Limit, Content-Type-Prüfung, ID-Formatprüfung, `crypto/rand`, generische Fehler, keine Inhalts-Logs) korrekt um.

Es besteht jedoch Härtungsbedarf bei der Ressourcenbegrenzung und der Server-Konfiguration. Daher werden Änderungen angefordert, ohne dass ein Blocker vorliegt.

---

### 1. Secrets
**Keine Befunde.**  
Es sind keine hartkodierten Schlüssel, Passwörter, Token oder URLs im sichtbaren Code enthalten. Logs (`main.go`) geben nur Adresse und Serverfehler aus, keine Paste-Inhalte.

### 2. Injection & Eingaben
**Keine kritischen Befunde.**  
- `POST /pastes` erzwingt `Content-Type: application/json` inkl. sinnvoller Parameterbehandlung (`charset=utf-8` wird akzeptiert).
- Der Body wird über `http.MaxBytesReader` auf 1 MiB begrenzt; Überschreitungen werden als `413` beantwortet.
- JSON-Ausgaben erfolgen ausschließlich über `json.Encoder`, dadurch werden Inhalte korrekt escaped.
- Die ID-Pfadparameter werden vor jedem Store-Zugriff mit `^[0-9a-f]{32}$` validiert, sodass kein ungeprüfter Zugriff auf die Map erfolgt.
- SQL-, Command- oder Path-Injection sind nicht möglich.

**Härtungsbedarf:**  
Der Wertebereich für `expires_in_seconds` ist nur nach unten begrenzt. Ein extrem großer positiver Wert kann bei der Umrechnung in `time.Duration` überlaufen und zu einem sofort abgelaufenen Paste führen.

### 3. Authentifizierung / Autorisierung
**Kein Finding im Rahmen der Anforderungen.**  
Die API besitzt keine Authentifizierung oder Autorisierung; jeder kann Pastes lesen, auflisten und löschen. Für eine öffentliche Pastebin-API ist das gemäß Spezifikation plausibel. Falls das Produkt nicht öffentlich betrieben werden soll, müsste ein Besitzer-/Token-Modell ergänzt werden.

### 4. Dependencies
**Keine Befunde.**  
Das Projekt nutzt ausschließlich die Go-Standardbibliothek; `go.mod` enthält keine externen Abhängigkeiten. Der Scanner-Output enthält für diesen Projekttyp keine Ergebnisse. Diese Lücke wird zur Kenntnis genommen, ist aber mangels externer Pakete risikoarm.

### 5. Konfiguration & Transport
**Findings (siehe unten).**  
Der Server startet ohne explizite HTTP-Timeout-Konfiguration. Zudem ist der Speicherverbrauch der In-Memory-Store unbegrenzt, sofern keine weiteren Schutzmaßnahmen ergriffen werden.

---

## Findings

### 1. [Medium] Unbegrenztes Speicherwachstum durch fehlende Maximalanzahl / Hintergrund-Cleanup
- **Betroffene Stellen:** `internal/store/store.go`, `internal/api/create.go`
- **Beschreibung:**  
  Der In-Memory-Store (`map[string]Paste`) kann beliebig viele Pastes aufnehmen. Abgelaufene Einträge werden nur bei Zugriffen (`Get`, `List`, `Delete`) gelöscht, nicht proaktiv. Ein Angreifer kann durch viele `POST /pastes`-Anfragen ohne Ablaufdatum den Speicher des Prozesses erschöpfen.
- **Konkreter Fix:**  
  Eine konfigurierbare maximale Anzahl aktiver Pastes einführen (z. B. `10_000`). Beim Überschreiten entweder den ältesten Paste verdrängen oder `503 Service Unavailable` liefern. Zusätzlich einen Hintergrund-Job einplanen, der abgelaufene Einträge regelmäßig entfernt, und optional ein Rate-Limit für `POST /pastes` vorsehen. Die Limitierung muss so gewählt sein, dass der legitime Produktbetrieb nicht beeinträchtigt wird.

### 2. [Medium] Fehlende Server-Timeouts (`http.ListenAndServe` ohne `http.Server`-Konfiguration)
- **Betroffene Stelle:** `main.go`, Funktion `main()`
- **Beschreibung:**  
  `http.ListenAndServe(addr, newMux())` erzeugt intern einen `http.Server` ohne gesetzte Timeouts. Dadurch ist der Server anfällig für Slowloris-/Slow-Read-Angriffe und kann Verbindungen unbegrenzt offen halten.
- **Konkreter Fix:**  
  Expliziten `http.Server` mit angepassten Timeouts verwenden, z. B.:
  ```go
  srv := &http.Server{
      Addr:              addr,
      Handler:           newMux(),
      ReadHeaderTimeout: 5 * time.Second,
      ReadTimeout:       10 * time.Second,
      WriteTimeout:      20 * time.Second,
      IdleTimeout:       120 * time.Second,
  }
  if err := srv.ListenAndServe(); err != nil {
      log.Fatalf("server error: %v", err)
  }
  ```
  Die Werte sind mit dem 1-MiB-Upload und den kurzen Antworten verträglich.

### 3. [Low] Fehlende obere Schranke für `expires_in_seconds`
- **Betroffene Stelle:** `internal/api/create.go`
- **Beschreibung:**  
  Negative Werte werden abgelehnt, aber sehr große positive Werte werden nicht validiert. Ein Wert nahe `math.MaxInt64` führt beim Multiplizieren mit `time.Second` zu einem Überlauf; der Paste ist dann sofort abgelaufen oder verhält sich sonst unerwartet.
- **Konkreter Fix:**  
  Eine sinnvolle Obergrenze definieren (z. B. 1 Jahr in Sekunden) und vor dem Anlegen prüfen:
  ```go
  const maxExpirySeconds = 365 * 24 * 60 * 60 // 1 Jahr
  if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds > maxExpirySeconds {
      writeError(w, http.StatusBadRequest, "expires_in_seconds too large")
      return
  }
  ```

### 4. [Low] JSON-Antworten ohne `X-Content-Type-Options: nosniff`
- **Betroffene Stelle:** `internal/api/helpers.go`, Funktion `writeJSON`
- **Beschreibung:**  
  Die API liefert JSON aus; ohne `X-Content-Type-Options: nosniff` kann ein Browser den Inhalt in bestimmten Einbettungsszenarien möglicherweise als HTML/anderes interpretieren. Das Risiko ist gering, die Härtung jedoch einfach.
- **Konkreter Fix:**  
  In `writeJSON` den Header setzen:
  ```go
  w.Header().Set("X-Content-Type-Options", "nosniff")
  ```

---

## Hinweis zum Scanner
Für diesen Projekttyp wurde laut Output kein Security-Scanner ausgeführt (`no applicable security scanners for this project type`). Die Lücke wird dokumentiert; da keine externen Abhängigkeiten verwendet werden, ist das Risiko aus fehlenden Dependency-Scans gering. Die Beurteilung stützt sich auf die manuelle Code-Analyse.