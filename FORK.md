# Fork notes

`main` in this fork is SpecterOps/BloodHound `main` plus the patches below. The `upstream` remote points at
SpecterOps/BloodHound with pushes disabled.

| Patch | Upstream reference | Drop when |
|---|---|---|
| No AZAddSecret edges onto managed identities | https://github.com/SpecterOps/BloodHound/pull/3254 (cherry-picked), https://github.com/SpecterOps/BloodHound/issues/943 | PR 3254 is merged and managed identities are also recognised by `serviceprincipaltype` |
| Recognise managed identities by `serviceprincipaltype`, not only by an inbound AZManagedIdentity edge | none | As above |
| Only tenant-scoped AZHasRole edges grant tenant-wide role membership | https://github.com/SpecterOps/BloodHound/issues/942 | Upstream honours the assignment scope |

The scoped role assignment patch needs the role assignment scope patch in the AzureHound fork, which stops AzureHound
from reporting administrative unit scopes as `/`.

Not carried: https://github.com/SpecterOps/BloodHound/pull/3316 changes the default metrics port binding, but the shipped
Docker configs set `metrics_port` explicitly. Set `"metrics_port": "127.0.0.1:2112"` in the deployment config instead.

## Syncing with upstream

```sh
git fetch upstream
git switch main
git merge upstream/main
go tool stbernard generate
go test ./packages/go/ein/... ./packages/go/analysis/...
```

When upstream merges an equivalent of a patch, resolve the conflict in favour of upstream and remove the row above. To
drop a patch early, revert its commits (`git log --oneline upstream/main..main`).

## Integration tests

The graph integration tests need only the test Postgres from `docker-compose.testing.yml`:

```sh
BH_POSTGRES_CONFIG=/path/to/postgresql.conf docker compose -f docker-compose.testing.yml up -d testdb
INTEGRATION_CONFIG_PATH=/path/to/integration.config.json go test -tags integration ./packages/go/analysis/azure/
```

Start from `local-harnesses/*.template`. `BH_POSTGRES_CONFIG` must point at a file the Docker daemon can read; when
`docker` runs through `sudo`, pass it with `sudo --preserve-env=BH_POSTGRES_CONFIG`.
