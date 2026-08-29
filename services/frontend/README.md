# Frontend

The BloodForBuds frontend is a TypeScript and React application built with TanStack Start, Tailwind CSS, and Mantine UI.

## Local development

From the repository root, run `just dev` and open <http://localhost:3456>. Caddy is the only service exposed to the host and forwards requests to the frontend over the private Compose network.

The development container bind-mounts this directory for Vite hot module replacement. Its container-built dependencies are kept in an anonymous `/app/node_modules` volume, which `just dev` renews whenever the environment starts.

## Authentication

The login page uses Firebase passwordless email links. Firebase credentials
exist in browser memory only long enough to exchange the ID token at
`POST /api/auth/session`; the frontend then signs out of the Firebase browser
SDK. The Go server owns the resulting HTTP-only session cookie.

The development Auth Emulator does not send email. After requesting a link,
the login page displays an emulator-only link that completes the local flow.
No emulator route or UI is present in production.

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
