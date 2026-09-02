# Pastebin REST API in Go

Eine kleine Pastebin-REST-API in Go, ausschließlich mit `net/http` aus der Standardbibliothek. Pastes werden in einem thread-sicheren In-Memory-Store gehalten, können optional ablaufen und lassen sich anlegen, abrufen, auflisten und löschen. Die API ist derzeit als lauffähiges Gerüst umgesetzt: der Store und die Hilfsfunktionen sind vollständig, die HTTP-Handler antworten als Stubs mit `501 Not Implemented`, bis die jeweiligen Endpunkt-Tickets sie ausfüllen.

## Tech Stack

- **Language**: Go 1.22+
- **Framework**: `net/http` (Standardbibliothek)
- **Storage**: In-Memory mit `sync.Mutex`
- **Testing**: `go test` / `httptest`

## Install

```sh
go mod download
```

## Run (dev)

```sh
go run .
```

Der Server lauscht auf `SERVER_ADDR` (Standard `localhost:8080`).

## Build (production)

```sh
go build ./...
```

## Konfiguration (Env)

| Variable     | Standard           | Bedeutung                     |
|--------------|--------------------|-------------------------------|
| `SERVER_ADDR`| `localhost:8080`   | Adresse, auf der der Server lauscht |

## Endpunkte

| Methode | Pfad              | Beschreibung                                      |
|---------|-------------------|---------------------------------------------------|
| POST    | `/pastes`         | Legt einen Paste an (Request `{content, language?, expires_in_seconds?}`) |
| GET     | `/pastes/{id}`    | Ruft einen Paste ab                                |
| GET     | `/pastes`         | Listet alle aktiven Pastes (ohne content)         |
| DELETE  | `/pastes/{id}`    | Löscht einen Paste                                |
| GET     | `/health`         | Health-Check, antwortet `200 {"status":"ok"}`     |

IDs sind 32 kleingeschriebene Hex-Zeichen, erzeugt aus 16 Bytes `crypto/rand`. Jeder Fehlerbody hat die Form `{"error":"<meldung>"}`. Zeiten werden als RFC3339 angegeben.

## Features

- Thread-sicherer In-Memory-Store für Pastes
- Zufällige, nicht aufzählbare 32-stellige Hex-IDs via `crypto/rand`
- Optionale Ablaufzeit (`expires_in_seconds`), abgelaufene Pastes werden wie nicht vorhanden behandelt
- REST-Endpunkte mit JSON-Fehlerkörpern und Health-Check
