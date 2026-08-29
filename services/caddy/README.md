# Caddy

Caddy is the HTTP ingress and reverse proxy for BloodForBuds. It listens on port 80; TLS termination is handled by infrastructure outside this repository.

The `Caddyfile` is copied into the image at build time, so the container does not use a configuration volume. Rebuild the image after changing the routing configuration.

Requests under `/api` are forwarded to the Go server after the `/api` prefix is stripped. All other traffic is forwarded to the frontend. Both upstreams are reached over the private Compose application network.

Caddy is the only service attached to the edge network and the only service that publishes a host port. The public health check is available at `/api/healthz`.
