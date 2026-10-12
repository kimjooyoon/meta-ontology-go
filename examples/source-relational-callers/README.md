# Gooo checks relationships between executions

Sometimes the requirement is a relationship: raising an input should not lower
a fee, or decoding an encoded value should return the original value. Those
checks can be ordinary Gooo activities. They can call activities whose bodies
are still being assembled.

The existing joint constructor can use their Boolean results as caller
obligations while keeping each called activity's local criteria. This example
uses that public route to assemble Gooo programs. The comparison and composition
logic is written in Gooo.

## Increasing a fee input

`monotonic.gooo.fixture` allows either operand order in `11-input`. A local
criterion anchors `Fee(11)` at zero. The relationship activity calls the same
function twice:

```gooo
activity Nondecreasing(Integer) -> Boolean id "relation://nondecreasing"
 computes "let first=Fee(input); let second=Fee(input+1); return first <= second"
```

Caller construction asks this activity to return true at input 4. The first
candidate passes its local anchor but fails the relationship. The second passes
both. Additional evaluation uses input 23.

`insufficient.gooo.fixture` only swaps the outer operands in `(11-input)+0`.
Both permitted bodies decrease as the input grows. Local anchors pass while
caller construction remains incomplete.

## Composing an inverse

`roundtrip.gooo.fixture` gives `Encode` and `Decode` two local-variable choices:
add seven or subtract seven. Each local criterion accepts either offset at zero.
A third Gooo activity states the relationship:

```gooo
activity RoundTrip(Integer) -> Boolean id "relation-codec://roundtrip"
 computes "return Decode(Encode(input)) == input"
```

The initial pair adds seven twice. Construction changes `Encode` to subtract
seven while `Decode` still adds seven. It passes the caller's relationship at
12 and the separate evaluation at -23. Two of the four possible combinations
were tried; the remaining two are unobserved in this run.

`roundtrip-insufficient.gooo.fixture` gives both variables the same `input+7`
expression in both activities. All four combinations fail the round trip.

## Local conditions can conflict with the caller

`conflicting.gooo.fixture` requires `Fee(4)=7` and `Fee(5)=6` through its local
Gooo criterion. The caller simultaneously asks `Fee(4) <= Fee(5)` to be true.
The recorded attempts keep both sides visible:

| Candidate | Local predicate/input pairs | Caller relationship |
| --- | --- | --- |
| `11-input` | 2/2 | 0/1 |
| `input-11` | 0/2 | 1/1 |

Construction remains `PARTIAL_FINITE`. It retains the first candidate, whose
local criteria pass. The additional caller evaluation is 0/1. The numbers let a
reader locate the disagreement in the supplied requirements. The constructor
records finite outcomes; it does not produce a general inconsistency proof.

## Once-only observations

| Source | Attempted combinations / declared space | Selected local criteria | Separate native relationship |
| --- | --- | --- | --- |
| Monotonic | 2/2 | 1/1 | 1/1 |
| Monotonic, insufficient choices | 2/2 | 1/1 | 0/1 |
| Round trip | 2/4 | 2/2 | 1/1 |
| Round trip, insufficient choices | 4/4 | 2/2 | 0/1 |
| Conflicting requirements | 2/2 | 2/2 | 0/1 |

Each of these five source programs was constructed and evaluated once. Model
calls were zero. Success refers to the finite supplied relation inputs. Integer
overflow boundaries and other input ranges were not evaluated here. The model,
performance and whole-language completion were not measured.

The development binary contains the scalar-criteria implementation prepared for
PR #1506. It was built from the #1505 source plus a bound temporary overlay before
the later record-only helper extraction. [Observation metadata](observation.json)
keeps the actual binary, base, overlay, source, case and output digests.

Saved reports preserve their warning that recorded claims are not reverified:
[monotonic](monotonic-report.md), [insufficient](insufficient-report.md),
[round trip](roundtrip-report.md),
[round trip insufficient](roundtrip-insufficient-report.md), and
[conflicting](conflicting-report.md). Rendering them made no new program or model
calls. Files named `*.selected.gooo.fixture` contain the selected executable
Gooo source from that run with source-line trailing whitespace removed. The
observation retains the raw selected source and both digests. Their `assembling`
blocks have been replaced by fixed
bodies. Keep the original `*.gooo.fixture` and the construction receipt when
continuing or inspecting the search: the original declares the conditions and
choices, and the receipt binds those contracts to the attempted masks and
observations. The compact observation also publishes each candidate’s contract
digest, predicate binding and finite local check results. The selected file
preserves activity identities, signatures and
fixed relation/predicate activities; it does not carry the assembly declaration.

## Run a program

Use a development build containing scalar criteria. From the repository root:

```sh
gooo body-construct \
  --source examples/source-relational-callers/monotonic.gooo.fixture \
  --entry Nondecreasing \
  --construction-cases examples/source-relational-callers/construction-cases.json \
  --cases examples/source-relational-callers/evaluation-cases.json \
  --attempts 2
```

For a round trip, use `roundtrip.gooo.fixture`, entry `RoundTrip`, the two
`roundtrip-*-cases.json` files, and `--attempts 4`. Substitute an insufficient or
conflicting source to inspect its outcome. Each command starts a fresh native
construction/evaluation; reading the saved Markdown alone does not.

## Research connection

Chen, Cheung and Yiu's
[Metamorphic Testing: A New Approach for Generating Next Test Cases](https://arxiv.org/abs/2002.12543)
explores deriving follow-up tests when exact expected outputs are difficult to
obtain. The Gooo examples use explicit source-written relationships and finite
caller inputs; they do not automatically discover those relationships.

Fu, Sun, Zhang and Liu's
[Test Adequacy for Metamorphic Testing](https://arxiv.org/abs/2412.20692)
considers the adequacy of both relations and inputs. It informs a remaining
question here: whether a relation and its sample inputs exclude the mistakes we
care about. Our counts retain that limited scope.
