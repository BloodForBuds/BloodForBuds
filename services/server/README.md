# Server

The BloodForBuds API is a Go HTTP server that listens on port 8080 inside the private Compose application network.

The server exposes `GET /healthz`. Caddy removes the `/api` prefix before forwarding requests, so the public health-check endpoint is `GET /api/healthz`.

`just dev` runs the server with Air and bind-mounts this directory into the development container. Changes to Go files trigger an automatic rebuild and restart. Production uses the compiled runtime image without Air or source mounts.

PostGIS runs on the private application network and stores its data in a named Docker volume. Development and production use separate volumes. Compose waits for the database before starting the server. The server registers the Goose Go migrations from `internal/migrations` and applies pending migrations before it starts accepting HTTP traffic. The official PostGIS image is currently amd64-only, so Compose explicitly enables Docker's amd64 emulation on arm64 development machines.

Production database settings can be overridden with `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD`. Use deployment-managed secrets rather than the local defaults outside development.

To run the server directly, point the standard PostgreSQL environment variables at an accessible database:

```sh
PGHOST=localhost PGPORT=5432 PGDATABASE=bloodforbuds \
  PGUSER=bloodforbuds PGPASSWORD=bloodforbuds \
  go -C services/server run ./cmd/httpserver
```

Run its checks with:

```sh
go -C services/server test ./...
```
