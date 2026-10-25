# Contextual mapping claims in Gooo

A mapping can help an agent find a related concept while retaining both original
identifiers. This example records the claimed relation, its origin, the two
source revisions and the query context. Missing or stale context remains visible.

```sh
gooo package execute --json \
  --cases examples/contextual-mappings/cases.json \
  examples/contextual-mappings/gooo.workspace.json > /tmp/mapping-claims.json

gooo package replay --json --receipt /tmp/mapping-claims.json \
  --cases examples/contextual-mappings/cases.json \
  examples/contextual-mappings/gooo.workspace.json
```

The program uses ordinary Gooo records, conditions, a local variable and a binding.
`Classify` describes a supplied assertion; `View` returns the lookup hint. The
source keeps `mapping_validity=UNASSESSED`, `identity_merge=false`,
`behavior_verified=false` and `permission_granted=false` for every output.
Even the exact-match assertion and equal input identifiers grant no new authority.

Twenty synthetic input rows check forty named activity outputs per invocation.
Six rows expose a recognized asserted relation; fourteen retain missing identity,
origin, query context, revision mismatch or unsupported relation. Every request
is compiled and executed twice. Saved replay reconstructs the selected source
and observes current outputs again, with no model calls. Expected records are
declared in `cases.json`; this checks the example policy over those rows.

The inputs are caller-supplied claims and opaque revision/reference strings.
The program checks their presence and exact contextual match. It does not fetch
those sources, validate URI syntax or establish the real-world validity of the
claimed mapping. Host adapters supply observed revision and origin bindings when
using the pattern with real sources. Execution/authorization decisions need their
own current contracts. This example runs when explicitly requested and adds no
mapping work to the compiler's ordinary execution path.

## Relation scope

The recognized predicates are SKOS `exactMatch`, `closeMatch`, `relatedMatch`,
and MMO `isMappedTo` and `isConstantRefinementOf`. They remain relation IRIs in
source data. This example describes directly supplied assertions and implements
no RDF/OWL entailment or graph traversal. SKOS declares `exactMatch` transitive
and distinguishes its semantics from OWL identity/class equivalence;
`closeMatch` and `relatedMatch` are not transitive. A future scoped RDF export
needs explicit assertion/context graphs instead of silently publishing a global
equivalence triple.
`UNSUPPORTED_RELATION` describes this policy's boundary, rather than a source
contradiction or a defect in the OWL standard.

## Research connection

Inspired by Blitz et al., *Meta Mesh Ontology for Cross-Domain Mental Health
Knowledge Integration: Development and Evaluation Study*, JMIR Mental Health
2026;13:e96311, [doi:10.2196/96311](https://mental.jmir.org/2026/1/e96311).
The article describes provenance-bearing mappings, separately versioned modules
and predefined competency questions. The example adapts those design ideas to
synthetic programming-tool concepts. It contains no clinical ontology/data copy.

See the [Gooo applicability review](../../docs/research/meta-mesh-applicability-20261025.ko.md)
and [W3C SKOS Reference, section 10](https://www.w3.org/TR/skos-reference/#mapping).
