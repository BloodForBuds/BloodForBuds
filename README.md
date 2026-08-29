# BloodForBuds

## Links

- Firebase: [prod](https://console.firebase.google.com/project/bloodforbuds/overview), [staging](https://console.firebase.google.com/project/bloodforbuds-staging/overview)
- [Infisical (secrets manager)](https://app.infisical.com/organizations/bf0482f1-aa9e-406f-b4b7-fed082f80829/projects/secret-management/f5166ea4-7823-40f7-9de9-9f646d645d96/overview)

## Environment configuration

Infisical is the source of truth for runtime configuration and secrets. The
repository's `.infisical.json` selects the BloodForBuds project, and the
Justfile recipes select the corresponding Infisical environment:

- `just dev` loads `dev`.
- `just staging` loads `staging`.
- `just prod` and `just main` load `prod`.

Authenticate the Infisical CLI before running these recipes; deployment
automation should authenticate with its environment-specific machine identity.
`.env.example` is only a reference listing the expected variable names and
safe example values. Do not copy credentials into local `.env` files.

Staging and production share `compose.deploy.yaml` because their service
topology is identical. Their Just recipes use distinct Compose project names,
which isolate containers, networks, and volumes if both environments run on
the same Docker host. Both publish Caddy on port 80 by default; set a different
`CADDY_HTTP_PORT` in staging if they are colocated.

The Firebase Web App values are public identifiers from Firebase project
settings. `FIREBASE_ADMIN_CREDENTIALS_BASE64` is a secret. Outside an
environment that provides Google Application Default Credentials, create it
from the Admin service-account JSON without checking that JSON into the repo:

```sh
openssl base64 -A -in firebase-admin.json
```

In each real Firebase project, enable Email/Password and Email link
(passwordless) authentication, then add the deployed application hostname to
Authentication's authorized domains.

Local development always uses the isolated Firebase Auth Emulator and ignores
real Firebase credentials.

## Backups

Staging and production run one Supercronic backup service with separate scripts
and B2/R2 repositories for the application database, key database, and OpenBao.
The key store has a hard 24-hour retention policy; application and OpenBao
snapshots default to 30 days and retain their newest snapshot during a
prolonged outage. Development does not run backups. See
[`services/backup/README.md`](services/backup/README.md) for bucket permissions,
required Infisical values, initialization, retention, and restore guidance.
OpenBao backups use a short-lived token obtained through an AppRole; see
[`services/openbao/README.md`](services/openbao/README.md) for its automatic
self-initialization, operator-access, and recovery-key procedures.
