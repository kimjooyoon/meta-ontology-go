# Capability discovery

`gooo discover` answers a narrow question: which declared capability in the
deterministic JEV catalog best matches a natural-language query? It binds the
query and observation trail to the supplied `.gooo` source. An optional saved
generation artifact can be replayed against that source, using the compiler's
pure projection and finite-case checks. `--execute-cases` additionally performs
a native build and two bounded executions. These paths make no model calls.

```sh
gooo discover --query "How do I generate a canonical .gooo declaration?" \
  examples/language-operation-catalog/main.gooo
```

Without a separate domain contract, `declaration_coverage` is `UNKNOWN`: the
target file cannot define its own denominator. Supply a Gooo file containing
the entity and activity declarations that define the domain scope:

```sh
gooo discover --json --query "How can I issue a billing receipt?" \
  --domain-contract examples/capability-discovery/domain-contract.gooo.fixture \
  examples/capability-discovery/current.gooo.fixture
```

The receipt compares stable declaration IDs, kinds, namespaces, entity fields,
and typed activity ports. Input types are compared in declaration order, so
`Merge(Integer, Text)` does not cover `Merge(Text, Integer)`. It ignores display
names and activity bodies, so renaming an entity does not erase coverage and a
declared body does not count as tested behavior. A missing or differently shaped
declaration is `PROGRESS`; the evidence lists uncovered IDs and ordered-port
mismatches. The exact contract bytes and its normalized semantic digest are
bound into the report and receipt.

Add `--json` to receive a machine-readable trail and the shared completeness
receipt. Its `catalog_match` dimension describes a catalog result only; it is
not a measure of whether Gooo implements the requested behavior.

## Reading the result

- `AVAILABLE` means the query matched a catalog entry. It does not mean that
  generation or runtime behavior was verified.
- `DEFERRED` means the catalog recognizes the capability but an external
  boundary remains unresolved.
- `UNKNOWN` means the query did not resolve to a known catalog entry. The
  compiler preserves this state instead of guessing.
- `PROGRESS` is the receipt decision while required evidence remains open.
  The receipt intentionally has no aggregate completeness score.

## Connect an existing generation

Save a source-owned IR search or typed path result and attach it to discovery:

```sh
env GOOO_LAYA_URL= GOOO_LAYA_API_KEY= go run ./cmd/gooo body-codegen --json \
  --activity ClampNegativeToZero examples/body-codegen/ir-search-source.gooo.fixture \
  > /tmp/gooo-generation.json
go run ./cmd/gooo discover --json --query "Generate Gooo code" \
  --domain-contract examples/capability-discovery/generation-domain.gooo.fixture \
  --generation /tmp/gooo-generation.json \
  examples/body-codegen/ir-search-source.gooo.fixture
```

The separate contract expects `ClampNegativeToZero` and `Pending`. The receipt
reports **generation coverage 1/2**: the first activity has a source-replayed
projection; the second still needs one. Stable IDs and ordered typed signatures
must match. The report binds the exact saved JSON digest, generated Go digest,
and activity ID. Changing the source, selected finite observations, or emitted
projection causes an error instead of producing a coverage result.

With no expected activity contract, the saved projection can still replay, but
generation coverage remains `UNKNOWN` with denominator zero. A successful
projection is a construction observation: even a candidate with zero matching
finite cases can be faithfully replayed. Its behavioral score remains in the
generation report. Discovery checks supported source-owned `assembling`
results; other generation routes need their own replay support.

Without `--generation`, generation coverage remains `UNKNOWN`. Independent
use-case coverage and reverse observation require runtime evidence: independent
inputs with expected outputs, mapped back to the originating source. A
catalog suggestion or a repeated fixture is not a substitute for those
observations. The domain contract measures declared scope only; it does not
define philosophical completeness or demonstrate real-world demand. See
[declared completeness receipts](../declared-completeness-receipt.md) and
[body generation](body-codegen.md) for the next stages.

## Observe the generated program

After saving the generation above, supply explicit native execution cases:

```sh
go run ./cmd/gooo discover --json --query "Generate Gooo code" \
  --domain-contract examples/capability-discovery/generation-domain.gooo.fixture \
  --generation /tmp/gooo-generation.json \
  --execute-cases examples/capability-discovery/runtime-cases.json \
  examples/body-codegen/ir-search-source.gooo.fixture
```

This command replays the source-owned selection, builds the closed
`Integer -> Integer` projection with Go 1.27.1, and runs the input sequence in
two separate processes. Add `--go-bin <path>` to select a local tool explicitly.
The runtime stage uses the existing 60-second operation budget and two-second
per-execution limits; its temporary workspace is removed afterward.

The example declares six inputs absent from this generation's training and
selection-holdout cases. The result reports generation **1/2**, independent
input expectations **6/6**, and reverse links **6/6**. These denominators measure
different obligations. `Pending` still has no generated artifact, so overall
completion remains `PROGRESS`. The nested `runtime` envelope retains the exact
generation parent receipt, source/artifact/toolchain identities, actual outputs,
two execution observations, and measured child resource use.

Independent input coverage counts each distinct input once. Every supplied
expectation for that input must match; conflicting duplicate expectations count
as one unmatched input. Cases already used in selection or its oracle feedback
do not enter this denominator. Selection-only inputs yield `UNKNOWN 0/0`, while
observed zero matches yield `PROGRESS 0/N`. Independence from model training
remains unobserved. Coverage applies to this activity and these inputs.

A build or execution error returns a nonzero exit code and retains the failed
stage in the report. Completed executions with mismatching expectations retain
their partial scores. Text output shows the runtime stage and independent-input
count; JSON includes the full observations. Fresh time and resource measurements
can vary between runs.

## Repository utility portfolio

The billing discovery example, which supplies no generation artifact, is also
the seventh case in the versioned language-utility portfolio. CI runs the same query twice, compares the full
reports byte-for-byte, and verifies the query trail against the copied source
and separate domain contract before adding it to the domain-completeness
receipt. The observed path closes source, syntax, semantic, outcome, replay,
and report-artifact stages. Resource use remains open because this path does
not currently record process time or memory. The underlying discovery receipt
continues to show its declaration coverage as `3/4`; code generation, reverse
observation, independent behavior cases, host permissions, and network
boundaries remain `UNKNOWN`. The catalog answer itself is never counted as
proof that Gooo implements the requested behavior.

The v2 utility contract defines 49 cells. Its CI floor remains 39 closed cells
and four complete use cases; CI computes the observed count from the current
evidence on each run. The discovery case cannot close until resource use is
measured. The common domain vector carries the exact report, replay, source,
and contract digests. Its generation, reverse-observation, and independent
use-case dimensions remain `UNKNOWN` for discovery. This is evidence for the
bounded command path, not closure of the broader receipt work tracked by
issue [#1023](https://github.com/kimjooyoon/meta-ontology-go/issues/1023).
