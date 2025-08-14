# Leaf Love Advisor — Go Edition

This is a lightweight refactor of your React/Tailwind app into a self-contained Go web server using only the standard library.

## Features
- Filter plants by light, care level, type, location, and size
- HTML templates rendered server-side
- Simple JSON API: `GET /api/recommend?lightCondition=...&careLevel=...&plantType=...&location=...&size=...`

## Run
```bash
go run ./cmd/server
# open http://localhost:8080
```

## Build
```bash
go build -o leaf-love ./cmd/server
./leaf-love
```

## Structure
```
cmd/server/main.go        # HTTP server and handlers
internal/models/types.go  # domain models
internal/data/plants.go   # in-memory dataset
web/templates/*           # (inline for now; see main.go)
web/static/*              # images + css
```

