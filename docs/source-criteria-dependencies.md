# Follow a requirement through its source calls

A criterion often delegates part of its rule to a fixed activity. Changing that
helper changes what must be checked, even when the candidate body's choices stay
the same. Gooo can expose the source-owned path beside those choices:

```gooo
activity Allowed(Integer,Integer) -> Boolean id "dependency://allowed" computes "return Within(input1)"
activity Within(Integer) -> Boolean id "dependency://within" computes "return input <= Ceiling(input)"
activity Ceiling(Integer) -> Integer id "dependency://ceiling" computes "return 20"
```

`Fee` declares `Allowed` as a fixed criterion. Its source dependency view follows
`Fee → Allowed → Within → Ceiling` using their stable activity IDs. A display
rename can retain those IDs; an exact source change keeps its own identity.

From a development checkout containing this change:

```sh
go run ./cmd/gooo body-context --value-flow --activity Fee \
  examples/criteria-dependencies/main.gooo.fixture
```

## Read the result

`criteria_flow.dependencies` uses `gooo/criteria-dependencies/v1`. An outgoing
edge means the activity on the left depends on the activity on the right.

| Field | Meaning |
| --- | --- |
| `activity_id` | The candidate activity owning the criteria |
| `edges[].kind` | `requires_predicate` or `calls` |
| `from_activity_id`, `to_activity_id` | Stable source identities at the ends |
| `criterion_id` | The source criterion that owns this closure |
| `ordinal` | Predicate declaration order, or call order inside that caller |

Each `predicates[].call_closure` exposes the same prepared closure used for fixed
program generation: helper IDs, program hashes, call edges and maximum call
budget. The new view projects that existing information. A primitive such as
`len` contributes no activity edge unless a declared source activity owns that
name. Separate criteria may retain the same helper edge under different criterion
IDs; repeated calls retain their own ordinals.

Reverse these edges to find candidate owners whose criteria depend on a changed
helper. Keep the source identity and criterion ID when presenting that path.
Cross-file references, candidate-body dependencies, external callers and a
workspace invalidation command require additional coverage.

## Keep source knowledge and observations separate

The dependency view follows all potential source calls, including unexecuted
branches. Its `RESOLVED` state records a complete prepared closure in the existing
fixed-body profile. It does not report whether a criterion passed. If an identity
is missing, the view returns `UNRESOLVED` and a reason without partial edges.
The symbolic value graph has its own state; a valid fixed program can use syntax
outside that graph's representation.

Existing pure-call limits apply: up to 32 activities, 16 call levels and 4,096
potential activity calls per invocation. Recursive calls are outside this profile.
The export runs no candidate or predicate and makes no model prediction. It adds
no model feature contract. Previous execution receipts remain attached to their
exact source and must retain their original observations.

See the [source example](../examples/criteria-dependencies/README.md) and
[Korean source-flow guide](criteria-source-flow.ko.md).

## Research context

[Nemo's rule language](https://knowsys.github.io/nemo-doc/intro/tour/) shows how
explicit relations and recursive rules make paths through knowledge readable.
That suggests a practical question for Gooo: which source requirement depends on
which helper? The implemented view answers it using the compiler's existing
finite call closure. This is an application of that reading to Gooo's current
source model; no cross-language execution or performance comparison is recorded.

See [the related-language comparison and development gaps](ontology-language-comparison.ko.md)
for the source review that informed this maintenance view.
