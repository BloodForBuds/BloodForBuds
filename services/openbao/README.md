# OpenBao

OpenBao is the KMS boundary for the future cryptographic-erasure
implementation. It is reachable only from the private Compose application
network; no OpenBao port is published to the host or edge network.

Both Compose environments set `mem_swappiness: 0` and make the memory and
memory-plus-swap limits equal. These layered controls prevent OpenBao's
in-memory secrets from being written to swap. OpenBao 2.6 removed `mlock`
support, so this deployment does not request the obsolete `IPC_LOCK`
capability.

Development runs OpenBao's ephemeral dev server with the root token supplied
through `BAO_DEV_ROOT_TOKEN_ID`. The Go server receives the matching value
through `BAO_TOKEN`. These credentials and development mode are strictly for
local development.

## Staging and production initialization

Staging and production run the persistent integrated-Raft configuration in
`openbao.hcl`. Its static seal reads the 32-byte `OPENBAO_UNSEAL_KEY` from the
OpenBao container's environment; the deployment injects this value from
Infisical. The key is never passed to the Go server or backup service. Losing
it permanently makes the OpenBao data and its backups unrecoverable.

OpenBao 2.6 [self-initializes](https://openbao.org/docs/configuration/self-init/)
an empty Raft volume on first startup. The ordered initialization requests:

1. Create the narrow `bloodforbuds-backup` policy.
2. Create the administrative `bloodforbuds-operator` policy.
3. Enable AppRole authentication.
4. Create both roles and register the Role IDs and Secret IDs injected from
   Infisical.

The temporary initialization root token is never returned and OpenBao revokes
it after these requests finish. There is no `BAO_BOOTSTRAP_TOKEN`, separate
bootstrap container, or `bao operator init` step. The initialization blocks
run only for empty storage; restarts use the persisted Raft state.

Each staging and production Infisical environment requires its own values:

```text
OPENBAO_UNSEAL_KEY
BAO_BACKUP_ROLE_ID
BAO_BACKUP_SECRET_ID
BAO_OPERATOR_ROLE_ID
BAO_OPERATOR_SECRET_ID
```

Generate independent values for every environment. The OpenBao static seal
expects the unseal key to decode to exactly 32 bytes:

```sh
openssl rand -hex 32 # OPENBAO_UNSEAL_KEY
openssl rand -hex 32 # BAO_BACKUP_ROLE_ID
openssl rand -hex 32 # BAO_BACKUP_SECRET_ID
openssl rand -hex 32 # BAO_OPERATOR_ROLE_ID
openssl rand -hex 32 # BAO_OPERATOR_SECRET_ID
```

After those values exist, `just staging` or `just prod` is sufficient for a
fresh cluster. Compose considers OpenBao healthy only after it is initialized
and unsealed. The server and backup service wait for that state.

An existing initialized volume does not replay self-initialization when this
configuration is introduced or changed. Provision changes to such a cluster
with an operator token, or restore/migrate it deliberately; never delete a
production Raft volume merely to rerun initialization.

Do not attach an initialized Shamir-sealed volume to this configuration without
completing OpenBao's seal-migration procedure first. Fresh or still-
uninitialized volumes can self-initialize directly with the static seal.

## Service and operator access

The backup service uses `BAO_BACKUP_ROLE_ID` and `BAO_BACKUP_SECRET_ID` to
obtain a fresh 30-minute token before every Raft snapshot. That token receives
no default policy and can only read `sys/storage/raft/snapshot`. No OpenBao
token is stored in Infisical.

The `bloodforbuds-operator` AppRole is the break-glass administrative path
after the initialization root token is revoked. Its login produces a
15-minute token, capped at 30 minutes, with full OpenBao administration
permissions. Infisical access to `BAO_OPERATOR_SECRET_ID` must therefore be
restricted to operators and deployment automation; it must never be injected
into the Go server or backup service.

Because OpenBao is not published to the host, open a shell in its container
through the environment-specific Compose project and authenticate without
printing the resulting token:

```sh
infisical run --env=staging -- docker compose \
  --project-name bloodforbuds-staging \
  --file compose.deploy.yaml exec openbao sh

export BAO_TOKEN="$(bao write -field=token auth/approle/login \
  role_id="$BAO_OPERATOR_ROLE_ID" \
  secret_id="$BAO_OPERATOR_SECRET_ID")"
```

Use `prod` and `bloodforbuds-prod` for production. Exit the shell when the
administrative work is complete; do not persist the short-lived token.

## Recovery keys

Self-initialization deliberately creates no recovery-key shares. This does not
affect normal startup or automatic unsealing, but OpenBao recovery mode and
root-token generation by a quorum are unavailable until recovery shares are
created.

The operator AppRole is authorized to use OpenBao's authenticated
`sys/rotate/recovery/*` endpoints. Run a separate recovery-key ceremony after
the cluster is established with `bao operator rotate-keys
-target=recovery`. Encrypt each resulting share to a different custodian's PGP
key, enable verification, and verify the shares before considering the
ceremony complete. Never print plaintext shares into deployment logs or store
all shares alongside `OPENBAO_UNSEAL_KEY` in Infisical. See the official
[`operator rotate-keys`](https://openbao.org/docs/commands/operator/rotate-keys/)
procedure.

Recovery shares authorize exceptional recovery and root-token workflows; they
cannot decrypt OpenBao storage without the static seal key. Back up and protect
`OPENBAO_UNSEAL_KEY` independently from the Raft snapshots.
