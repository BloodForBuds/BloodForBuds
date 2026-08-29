# Backups

One Supercronic service backs up the application PostGIS database, key
PostgreSQL database, and OpenBao integrated-Raft state. Every target has a
dedicated script and independent Restic repositories in Backblaze B2 and
Cloudflare R2. The service runs only in staging and production.

## Schedule and retention

All targets are backed up at container startup and every six hours. Scheduled
jobs are staggered by ten minutes and share a process-level lock, so they never
overlap inside this service.

| Target | Backup | Expiry check | Retention | Outage behavior |
| --- | --- | --- | --- | --- |
| Key store | Every 6 hours | Hourly | 24 hours | Hard expiry can remove every old snapshot |
| App store | Every 6 hours | Daily | 30 days | The newest snapshot is retained |
| OpenBao | Every 6 hours | Daily | 30 days | The newest snapshot is retained |

Key expiry always runs `restic prune --max-unused 0`, including when no current
snapshot needs forgetting. This retries physical deletion after an interrupted
prune and gives a practical deletion bound of less than 25 hours plus pruning
and provider-side deletion time. App and OpenBao pruning runs daily.

## Data flow

- `backup-key-store` creates one custom-format `pg_dump` in memory-backed
  `/tmp`, uploads that exact artifact independently to B2 and R2, then removes
  it.
- `backup-app-store` streams a separate consistent `pg_dump` directly into
  each Restic repository. This avoids sizing `tmpfs` for the larger application
  database; the two provider snapshots may differ by the writes occurring
  between their dumps.
- `backup-openbao` authenticates with its AppRole immediately before each run,
  uses the resulting 30-minute token for `bao operator raft snapshot save`,
  keeps the snapshot only in memory-backed `/tmp`, uploads it to both
  providers, then removes it.

No database or OpenBao data volume is mounted into the backup service.

## Required configuration

Each of `APP`, `KEY`, and `OPENBAO` requires the following eight variables in
both staging and production, with the prefix substituted accordingly:

```text
<TARGET>_BACKUP_B2_REPOSITORY
<TARGET>_BACKUP_B2_ACCOUNT_ID
<TARGET>_BACKUP_B2_ACCOUNT_KEY
<TARGET>_BACKUP_B2_REPOSITORY_PASSWORD
<TARGET>_BACKUP_R2_REPOSITORY
<TARGET>_BACKUP_R2_ACCESS_KEY_ID
<TARGET>_BACKUP_R2_SECRET_ACCESS_KEY
<TARGET>_BACKUP_R2_REPOSITORY_PASSWORD
```

The service reuses the existing `APP_DB_*` and `KEY_DB_*` connection values.
It additionally requires `BAO_BACKUP_ROLE_ID` and `BAO_BACKUP_SECRET_ID`. The
backup exchanges these AppRole credentials for a short-lived token whose
policy only allows reading Raft snapshots:

```hcl
path "sys/storage/raft/snapshot" {
  capabilities = ["read"]
}
```

Generate distinct random values for both AppRole credentials in each
environment and store them in Infisical. The Secret ID is a long-lived machine
credential and must remain secret. OpenBao registers these values during
automatic self-initialization; see `services/openbao/README.md`. Neither
credential is the removed Go-server `BAO_TOKEN`.

Generate a distinct Restic password for all twelve repositories—three targets,
two providers, and two environments:

```sh
openssl rand -base64 48
```

B2 and R2 credentials must be scoped to their respective target bucket and
allow listing, reading, writing, and deleting objects. Do not configure Object
Lock, bucket locks, or provider lifecycle rules that retain key-store objects
beyond the cryptographic-erasure deadline.

The native B2 backend is deliberate. [Restic currently recommends B2's
S3-compatible API](https://restic.readthedocs.io/en/stable/030_preparing_a_new_repo.html#backblaze-b2)
for better error handling, but also documents that S3 deletes only hide B2
objects until a lifecycle rule removes old versions. Native B2 deletion better
matches the key store's short erasure window; the independent R2 repository and
stale-backup health check mitigate its reliability tradeoff.

## One-time initialization

The service never initializes missing repositories during normal startup. This
prevents an empty replacement repository from appearing healthy after data
loss. After creating all six buckets for an environment and adding the required
Infisical values, initialize the repositories explicitly:

```sh
infisical run --env=staging -- docker compose \
  --project-name bloodforbuds-staging \
  --file compose.deploy.yaml run --rm backup init

infisical run --env=prod -- docker compose \
  --project-name bloodforbuds-prod \
  --file compose.deploy.yaml run --rm backup init
```

Initialization is idempotent when some or all repositories already exist.

## Health and recovery

The service is healthy only when all three targets reached both providers
within the last seven hours. Supercronic logs each job result to Docker logs.
Docker Compose does not restart unhealthy containers automatically, so
deployment monitoring must alert on this health status.

Restore the chosen Restic snapshot outside production first:

- Feed `app-store.dump` or `/tmp/key-store.dump` to `pg_restore` against an
  empty database.
- Restore `/tmp/openbao.snap` and use `bao operator raft snapshot restore`
  following the OpenBao recovery procedure. The matching static seal key must
  still be available from Infisical.
