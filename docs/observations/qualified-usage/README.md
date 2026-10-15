# Gooo qualified usage: first source and inputs

The runnable [source example](../../../examples/recorded-use-origins/README.md#읽기-사건을-rdf로-내보내기)
adds a PROV-O qualified usage view to the registered recorded-use metaprogram.
Gooo owns the identities, conditions, literal escaping and N-Triples statements.
The production Go runtime and core ontology vocabulary are unchanged.

- Export cases: **12/12** in six bounded two-case inputs.
- Separate formatting cases: **32/32** (12 decimal, 10 SHA shape, 10 quoted text).
- New saved-record consumer queries: two original JSON byte sequences; the
  second adds a newline. Original case/frame/assignment/use coordinates stay
  fixed; receipt hash and all three RDF identities change.
- Independent RDF syntax/direction/type checks: seven saved graphs, then one
  graph from the new byte variant, each with 26 statements.
- New V5 source coherence: AST, bytes, semantics, GetPut and PutGet pass.
- Original producer executions and model calls: **0**.

[observations.json](observations.json) keeps each command's scope, first elapsed
time and source/input identities. The initial zero-based-loop restriction failure
is retained along with the corrected source. Native CI tests are new consumers
of the published files; completed historical local cohorts were not rerun.

[first-observations.tar.gz](first-observations.tar.gz) retains 73 raw members,
including all captured source/input/stdout/stderr payloads and the pinned independent
Go parser source/module. [archive-manifest.json](archive-manifest.json) records every
member's length and digest. Go executable location is shortened only in command
metadata and its original digest is retained. The subsequent readonly fixture
binding check has its own compressed record; it checks this unchanged archive.

These are finite source observations. Roles, times, output derivation, global
execution identity, complete causal dependencies and full PROV constraint checking
have additional recording requirements. The core kernel profile and historical
registration denominators retain their existing scope. Exact-source public CI
acceptance is recorded separately from the first private source/input executions.
