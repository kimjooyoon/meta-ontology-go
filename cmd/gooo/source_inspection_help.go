package main

const sourceImpactHelp = `Explain potential effects of a source change

` + sourceImpactUsage + `

Compare declarations by stable ID. Follow source calls, observation oracles,
input/output types, field types and explicit runtime bindings in each revision.
The default text report lists reasons and source-owned selection, condition,
held-evaluation and observation counts. Use --format json for structured output.
Each affected declaration has one nearest-change explanation per revision.
Unresolved calls, generated candidate call sets and policy consumers remain explicit.
Exit 0 means a report was produced. Check its static status and unresolved list.
Bodies, candidates and models are not executed. Recheck means potential impact;
source replay or evaluation must establish the actual result.
See docs/source-change-impact.md.
`

const semanticContextHelp = `Read a semantic neighborhood

` + semanticContextUsage + `

Start with a stable ID from gooo inspect. Context follows deterministic PROV-O
relations in both directions and the participants of explicit runtime bindings.
Depth, node and JSON byte limits keep the response bounded. Truncated responses
name the reason and a next_root; query that ID with the same --expect-source digest
to inspect another neighborhood. The count covers omitted adjacent nodes only.

Each activity keeps its value program, explicit multi-input order and full assembly contract.
Output preserves whole nodes, relations and bindings. Exit 0 means the requested
view was returned; inspect truncated before assuming the neighborhood is complete.
Source digests bind all source bytes, including bodies. The context command reads
declarations without executing activities or calling a model. Use body-context for
the separate body-assembly input view. See docs/semantic-context.md.
`
