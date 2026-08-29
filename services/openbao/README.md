# OpenBao

OpenBao is the KMS boundary for the future cryptographic-erasure implementation. It is reachable only from the private Compose application network; no OpenBao port is published to the host or edge network.

Both Compose environments set `mem_swappiness: 0` and make the memory and memory-plus-swap limits equal. These layered controls prevent OpenBao's in-memory secrets from being written to swap. OpenBao 2.6 removed `mlock` support, so this deployment does not request the obsolete `IPC_LOCK` capability.

Development runs OpenBao's ephemeral dev server with the root token `bloodforbuds-dev-only`. This token and mode are strictly for local development.

Production runs the persistent integrated-Raft configuration in `openbao.hcl`. A new production volume starts uninitialized and sealed. Initialization, secure distribution of unseal material, unsealing, authentication policy, and TLS for OpenBao are deployment responsibilities and are intentionally not automated by this repository. The Go server can verify that an uninitialized or sealed OpenBao process is reachable, but it does not perform key lifecycle operations yet.
