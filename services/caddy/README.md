# Caddy

Caddy is the HTTP ingress and reverse proxy for BloodForBuds. It listens on port 80; TLS termination is handled by infrastructure outside this repository.

The `Caddyfile` is copied into the image at build time, so the container does not use a configuration volume. Rebuild the image after changing the routing configuration.

The initial configuration exposes `GET /healthz` and returns a 404 for unmatched requests. Add application routes to `Caddyfile` as services are introduced. For example:

```caddyfile
reverse_proxy /api/* api:3000
```
