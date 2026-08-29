# Frontend

The BloodForBuds frontend is a TypeScript and React application built with TanStack Start, Tailwind CSS, and Mantine UI.

## Local development

From the repository root, run `just dev` and open <http://localhost:3456>. Caddy is the only service exposed to the host and forwards requests to the frontend over the private Compose network.

To run the frontend without Docker:

```sh
npm --prefix services/frontend install
npm --prefix services/frontend run dev
```

## Checks

```sh
npm --prefix services/frontend run build
npm --prefix services/frontend run typecheck
```
