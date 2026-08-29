# List available recipes.
default:
    @just --list

# Start the development environment and build images.
dev:
    docker compose --file compose.dev.yaml up --build --renew-anon-volumes

# Run integration tests against the development databases and OpenBao.
test-integration:
    docker compose --file compose.dev.yaml up --detach --build --wait server
    docker compose --file compose.dev.yaml exec --no-TTY server go test -count=1 -tags=integration ./integration

# Start the production environment and optionally detach (`just prod true`).
prod detach="false":
    docker compose --file compose.prod.yml up --build {{ if detach == "true" { "--detach" } else { "" } }}

# Alias for the production environment.
main detach="false": (prod detach)
