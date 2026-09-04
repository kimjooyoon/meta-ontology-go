# Public generate self-observation

This small `.gooo` project is an opt-in example for the public compiler
self-observation ledger.

```text
first generate + empty ledger  -> UNKNOWN: one comparable observation
second generate + same ledger  -> CLOSED: stable non-executing candidate
third generate + same ledger   -> CLOSED: deterministic candidate replay
```

The normal `gooo generate` path is unchanged unless the caller supplies
`--observation-ledger DIRECTORY`. The directory is caller-owned and receives a
digest-only append-only ledger, an invocation receipt, and machine/human
reports. The generated source and manifest remain in the ordinary `--out`
directory.

The candidate is evidence only: it contains no executable patch, has
`execution_allowed: false`, and requires the repository's existing explicit
proposal, authorization, and certificate controls before any future use. No
repository source is modified by this observation loop.

The continuity handoff is explicit and caller-owned:

```text
candidate -> gooo authorize-discovery --decision accept|reject
          -> gooo certify-discovery (accepted only)
          -> gooo generate --continuity-certificate CERTIFICATE
```

The decision receipt and certificate carry the raw candidate digest and every
source, semantic, toolchain, contract, evaluator, output, and manifest digest.
Certification replays generation and records a typed conversion because a
public-discovery candidate and a semantic-retention certificate represent
different operations. Rejection is a terminal human decision; missing,
tampered, stale-source, stale-policy, and mismatched-toolchain evidence remain
distinct fail-closed outcomes. The continuity verifier requires four equal
candidate-digest edges and zero manual transformations.

The incompatible-source case is derived into caller-owned temporary output from
`project.gooo`; no second tracked language source is needed.

Corpus accounting is explicit: the fixed language corpus moves from 51 cases,
56 sources, and 997 `.gooo` lines to 53 cases, 59 sources, and 1,022 lines after
the v15 partial-reuse example is appended.

## Utility acceptance contract

The `Compile` activity in `project.gooo` carries the v1 utility contract as
lowered meta-code. It fixes a nine-step public journey from ordinary baseline
generation through observation, explicit authorization, certification,
certified consumption, generated-project build/test, and a human dossier.
The project input is two regular files (the `.gooo` source and the tagged Go
acceptance test), with one `.gooo` source, two entities, one activity, and two
derived relations. The generator must publish two output artifacts: one Go
file and one manifest. The lane records three baseline and three certified
consumption `wall_ms`/`peak_rss_kib` observations under the same source,
semantic, compiler, toolchain, contract, and evaluator digests.

The six utility cases are exactly two `CLOSED`, two `UNKNOWN`, and two
`REFUTED`. The runtime comparison remains `UNKNOWN` unless the predeclared
runtime modes are equivalent; deterministic semantic operation reduction is
reported independently. The published evidence denominator is exactly 32.

## Public generated-test reuse contract

The same canonical project also declares a v1 test-reuse policy in its
lowered \`Compile\` meta-code. An independent caller may run the fixed
\`go test -tags generated_project -count=1 .\` contract once, publish an
immutable caller-owned receipt, and later request reuse explicitly. The
second public command validates the receipt against the canonical source and
semantic digest, generated program and manifest, compiler and released-tool
identity, Go 1.27 toolchain, test command, test contract, and successful
original result. It never auto-skips a test.

The v13 journey is baseline generation -> baseline build -> baseline test ->
receipt -> explicit authorization -> replay validation -> reused test ->
evidence. The six cases are exactly two \`CLOSED\`, two \`UNKNOWN\`, and two
\`REFUTED\`: successful baseline execution, authorized immutable receipt reuse,
missing authorization, stale or unbounded evidence, tampered receipt, and
policy mismatch. No score is emitted; the artifact denominator is exactly 24.

## Public orchestration protocol

The same lowered `Compile` meta-code now declares the v1 orchestration state
machine. Its ordinary journey is semantic discovery -> proposal -> explicit
authorization -> durable certificate -> ordinary generate -> real project
validation -> immutable test-receipt reuse -> evidence. The public orchestrator
executes the discovery and proposal steps, then stops at `AUTHORIZE` and emits a
structured `UNKNOWN` handoff until the caller supplies an existing explicit
authorization artifact. It never approves, mutates source, writes the
repository, or silently skips a state.

The resume path consumes the caller-owned authorization, certifies it, performs
ordinary generation, validates the generated project with the real Go build and
test contract, and reuses only the exact immutable successful test receipt. The
state transitions and all six acceptance cases are read from the lowered
`.gooo` activity; Go provides only generic transition dispatch and artifact
validation. The six orchestration cases are exactly two `CLOSED`, two
`UNKNOWN`, and two `REFUTED`, with an artifact denominator of 24.

The utility contract measures the fixed manual route as 15 public CLI
invocations and the orchestrated route as 4: one prepare, two explicit human
decisions (accept and reject), and one resume. Semantic, lowering, generation,
test, handoff-artifact, and explicit-decision counts are preserved. Wall-clock
and RSS comparison remains `UNKNOWN` because the routes are not equivalent
runtime modes. The v14 dossier preserved the 53-case / 58-source snapshot; v15
appends one canonical multi-partition source and keeps the case corpus fixed.
