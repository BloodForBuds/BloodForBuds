# Server

The BloodForBuds API is a Go HTTP server that listens on port 8080 inside the private Compose application network.

The server exposes `GET /healthz`. Caddy removes the `/api` prefix before forwarding requests, so the public health-check endpoint is `GET /api/healthz`.

`just dev` runs the server with Air and bind-mounts this directory into the development container. Changes to Go files trigger an automatic rebuild and restart. Production uses the compiled runtime image without Air or source mounts.

The server has three infrastructure boundaries:

- `internal/app_store` owns the pgx-backed application database pool and its embedded Goose SQL migrations. Its PostgreSQL image includes PostGIS.
- `internal/key_store` owns a separate pgx-backed PostgreSQL pool and an independent set of embedded Goose SQL migrations.
- `internal/kms` owns the OpenBao API client. It currently verifies connectivity and reports initialization/seal state; key lifecycle and cryptographic erasure are not implemented yet.

All three dependencies run only on the private application network. Each database has its own named volume, and production OpenBao also uses a persistent volume. Compose waits for the dependencies before starting the server, which applies both migration sets before accepting HTTP traffic. The official PostGIS image is currently amd64-only, so Compose explicitly enables Docker's amd64 emulation on arm64 development machines.

Production settings can be overridden with `APP_DB_*`, `KEY_DB_*`, and `BAO_TOKEN`. Use deployment-managed secrets rather than the local defaults outside development. A new production OpenBao instance must be initialized and unsealed out of band.

To run the server directly, point both store configurations and the OpenBao client at accessible services:

```sh
APP_DB_HOST=localhost APP_DB_PORT=5432 APP_DB_NAME=bloodforbuds \
  APP_DB_USER=bloodforbuds APP_DB_PASSWORD=bloodforbuds \
  KEY_DB_HOST=localhost KEY_DB_PORT=5433 KEY_DB_NAME=bloodforbuds_keys \
  KEY_DB_USER=bloodforbuds_keys KEY_DB_PASSWORD=bloodforbuds_keys \
  BAO_ADDR=http://localhost:8200 BAO_TOKEN=development-token \
  go -C services/server run ./cmd/httpserver
```

Run its checks with:

```sh
go -C services/server test ./...
```

Run the infrastructure integration suite with:

```sh
just test-integration
```

This starts the development server and its PostGIS, key-store PostgreSQL, and OpenBao dependencies, then verifies both connection pools, both Goose migration histories, the PostGIS extension, the key-store schema, and OpenBao's initialized/unsealed development state. The services remain running afterward for local development.
