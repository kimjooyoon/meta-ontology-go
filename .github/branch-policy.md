# Branch and pull-request policy

The current routes are:

```text
agent/* -> dev
dev -> main (direct fast-forward when ancestry permits)
agent/main-promotion-snapshot-<dev-sha> -> main (exact dev tree, main parent)
```

Work branches target `dev`. Promotion is a same-repository pull request with
`base=main` and either `head=dev` or `head=agent/main-promotion-snapshot-<dev-sha>`.
CI verifies the live PR tuple and refs before it writes a
`promotion_authorization` record. A snapshot is accepted only when its tree is
identical to the live `dev` tree and its sole parent is the live `main` commit.
That record is bound to the six canonical CI checks and exact proof artifacts.
CI does not update refs.

The repository defines six canonical required status contexts:

- `gofmt`
- `go vet`
- `go test`
- `go test -race`
- `Semantic conformance`
- `CI policy`

The GitHub `main` protection rule now requires exactly these six contexts. The
retired `CI guardian` check has been removed from that rule. Its up-to-date
requirement, linear-history rule, administrator enforcement, and force-push and
deletion restrictions remain enabled. The `dev` rule has no required status
contexts.

Review identities, review counts, last-push approval, a Guardian workflow, an
app-bound status, and a separate branch-protection snapshot are not proof inputs.
The proof decision comes from exact source, run, job, artifact, scope, and
provenance observations. Unknown or mismatched machine evidence fails closed.

## Protected kernel changes

Changes to CI policy, verifier code, governance data, and proof code run through
the same six required checks as other changes. The policy job checks the exact
candidate revision and emits a proof and receipt bound to that run. There is no
separate app or one-time human-authorization route.

## Pull-request scope

Any syntactically valid `agent/*` branch can target `dev` without a per-branch
registration. Every pull request runs the same six canonical checks against
its exact head, so branch-to-path allowlists do not narrow verification and do
not authorize changes. The old scope tables remain only as compatibility data
for legacy verifier fixtures; the active CI route does not consult them.

The workflow pins `GOOO_CONFORMANCE_STAGE=0`: the Go verifier remains
authoritative until the criteria in `.github/conformance-plan.md` have runnable
evidence. Formatting, vet, unit tests, race tests, semantic conformance, and CI
policy remain the canonical checks.
