# Caddy

Caddy is the HTTP ingress and reverse proxy for BloodForBuds. It listens on port 80; TLS termination is handled by infrastructure outside this repository.

The `Caddyfile` is copied into the image at build time, so the container does not use a configuration volume. Rebuild the image after changing the routing configuration.

All traffic is forwarded to the frontend service over the private Compose application network. Caddy is the only service attached to the edge network and the only service that publishes a host port.
