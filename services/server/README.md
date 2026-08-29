# Server

The BloodForBuds API is a Go HTTP server that listens on port 8080 inside the private Compose application network.

The server exposes `GET /healthz`. Caddy removes the `/api` prefix before forwarding requests, so the public health-check endpoint is `GET /api/healthz`.

`just dev` runs the server with Air and bind-mounts this directory into the development container. Changes to Go files trigger an automatic rebuild and restart. Production uses the compiled runtime image without Air or source mounts.

Run the server directly with:

```sh
go -C services/server run .
```

Run its checks with:

```sh
go -C services/server test ./...
```
