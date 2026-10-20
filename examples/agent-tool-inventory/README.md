# Compare an agent's reported tool inventories with Gooo

An agent may receive different tool lists before and after a host or
configuration change. The `Inspect` activity takes three explicit lists:
required IDs, IDs reported before, and IDs reported after. Its Gooo body checks
exact ID membership in the order of the required list and returns the first
missing ID for each side. It also carries the caller's three origin identifiers
into the result. The classifier has no model call and does not read the host's
configuration itself.

Run the eight named cases from the repository root with Go 1.27.2:

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/agent-tool-inventory/cases.json \
  examples/agent-tool-inventory/gooo.workspace.json \
  > /tmp/gooo-agent-tool-inventory.json
```

The cases cover loss and recovery of required IDs, preserved availability,
missing IDs on both sides, an empty required list or empty-string required ID,
an absent origin, and different servers exposing the same local tool name. The
output statuses are `REGRESSED`, `RECOVERED`, `PRESENT_BOTH`, `MISSING_BOTH`,
`INVALID_REQUIRED`, and `INVALID_ORIGIN`. An empty-string required ID is
invalid; whitespace is not trimmed. An empty reported inventory is valid and
means no IDs were reported.
Repeated reported IDs do not change membership; the count fields retain the
list lengths, including repeats.

The origin strings identify the caller's source for each list. They are echoed,
not authenticated. `PRESENT_BOTH` means only that each required exact ID occurs
in both supplied lists. It does not mean a tool invocation works or that the
host supplied a complete inventory. The caller must capture both lists from
the intended host and preserve their source separately. Tool IDs should include
the host and server namespace, for example `mcp://server-a/alpha`; the Gooo
comparison neither qualifies nor normalizes IDs.

This is a small source-owned policy: a reader can change the membership and
classification rules in `main.gooo` without adding a Go comparator. Inputs are
bounded by the current native collection profile (up to 256 values per list).
