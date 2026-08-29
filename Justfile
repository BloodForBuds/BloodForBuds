# List available recipes.
default:
    @just --list

# Start the development environment and build images.
dev:
    docker compose --file compose.dev.yaml up --build

# Start the production environment and optionally detach (`just prod true`).
prod detach="false":
    docker compose --file compose.prod.yml up --build {{ if detach == "true" { "--detach" } else { "" } }}

# Alias for the production environment.
main detach="false": (prod detach)
