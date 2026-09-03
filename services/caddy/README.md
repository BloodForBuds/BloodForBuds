# Caddy

Caddy is the HTTP ingress and reverse proxy for BloodForBuds. It listens on container port 80 and publishes host port `CADDY_HTTP_PORT`, which defaults to 80; TLS termination is handled by infrastructure outside this repository.

The `Caddyfile` is copied into the image at build time, so the production container does not use a configuration volume. Rebuild the production image after changing the routing configuration.

In development, the Caddy service directory is bind-mounted read-only and
Caddy watches `Caddyfile.dev` for changes. Production continues to use only the
`Caddyfile` configuration baked into the image.

Development uses `Caddyfile.dev`, which additionally proxies the Firebase Auth
Emulator API. The production image contains only `Caddyfile` and never routes to
the emulator.

Requests under `/api` are forwarded to the Go server after the `/api` prefix is stripped. All other traffic is forwarded to the frontend. Both upstreams are reached over the private Compose application network.

Caddy is the only service attached to the edge network and the only service that publishes a host port. The public health check is available at `/api/healthz`.
