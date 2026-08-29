# List available recipes.
default:
    @just --list

# Test the databases, OpenBao, and Firebase Auth Emulator integrations.
test-integration:
    infisical run --env=dev -- docker compose --file compose.dev.yaml up --detach --build --wait server
    infisical run --env=dev -- docker compose --file compose.dev.yaml exec --no-TTY server go test -count=1 -tags=integration ./integration

# Start the development environment and build images.
dev:
    infisical run --env=dev -- docker compose --file compose.dev.yaml up --build --renew-anon-volumes

# Start the staging environment and optionally detach (`just staging true`).
staging detach="false":
    infisical run --env=staging -- docker compose --project-name bloodforbuds-staging --file compose.deploy.yaml up --build {{ if detach == "true" { "--detach" } else { "" } }}

# Start the production environment and optionally detach (`just prod true`).
prod detach="false":
    infisical run --env=prod -- docker compose --project-name bloodforbuds-prod --file compose.deploy.yaml up --build {{ if detach == "true" { "--detach" } else { "" } }}

# Stop the development environment
dev-stop:
    infisical run --env=dev -- docker compose --file compose.dev.yaml down

# Alias for the production environment.
main detach="false": (prod detach)
