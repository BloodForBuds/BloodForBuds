# OpenBao

OpenBao is the KMS boundary for the future cryptographic-erasure implementation. It is reachable only from the private Compose application network; no OpenBao port is published to the host or edge network.

Both Compose environments set `mem_swappiness: 0` and make the memory and memory-plus-swap limits equal. These layered controls prevent OpenBao's in-memory secrets from being written to swap. OpenBao 2.6 removed `mlock` support, so this deployment does not request the obsolete `IPC_LOCK` capability.

Development runs OpenBao's ephemeral dev server with the root token supplied through `BAO_DEV_ROOT_TOKEN_ID`. The Go server receives the matching value through `BAO_TOKEN`. These credentials and development mode are strictly for local development.

Staging and production run the persistent integrated-Raft configuration in `openbao.hcl`. Its static seal reads the 32-byte `OPENBAO_UNSEAL_KEY` from the OpenBao container's environment; the deployment must inject this value from Infisical. The key is never passed to the Go server. Losing it permanently makes the OpenBao data and its backups unrecoverable.

A fresh production volume still requires a one-time `bao operator init`. Initialization returns recovery-key shares rather than Shamir unseal shares; distribute those securely and use the initial root token only to provision authentication, policies, and secrets engines. After initialization, OpenBao automatically unseals on restart while the matching static seal key is available. TLS for OpenBao remains a deployment responsibility. The Go server can verify OpenBao's state but does not perform key lifecycle operations yet.

Do not attach an existing initialized Shamir-sealed volume to this configuration without completing OpenBao's seal-migration procedure first. Fresh or still-uninitialized volumes can be initialized directly with the static seal.
