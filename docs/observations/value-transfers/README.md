# Value origins across calls and package bindings

The observations distinguish native outputs, static source paths, and a Gooo
consumer of saved path facts. Sources and their original failures remain linked
to the compiler that produced them.

## Producers and outcomes

The `producerNNNN.patch.gz` files preserve each source patch. Every patch applies to compiler revision
`836acea0bafc4622ace4b144a224a455858d7c03`. The corresponding JSON lists the SHA-256
of each complete replacement file. Later source corrections do not change earlier
observations. Saved programs were not executed again locally.

| Producer | New observation | Outcome |
| --- | --- | --- |
| [6083](producer6083.json) | Initial call and bind implementation | Compilation failed before native execution; [log](observation6083.log). |
| [6203](producer6203.json) | Two calls and a bind, plus an unused argument | Four native outputs matched, then origin assertion failed because a helper was absent from the catalog. The other program failed before execution because Integer tuples took the record route. [Log](observation6203.log). |
| [6373](producer6373.json) | Fresh call and bind program, fresh unused argument | Native 4/4 and 2/2 respectively. Candidate and external inputs stay distinct; unused second argument contributes no returned-value origin. [Records](transfer6373.json), [unused argument](unused6389.json). |
| [6599](producer6599.json) | Tuple decoding, scope bounds, registry version 48 | Fresh static checks passed; no native result claimed. [Log](observation6599.log). |
| [6781](producer6781.json) | Gooo explains four actual saved path/location facts; new Integer subtraction | Consumer 4/4 with no original program execution; subtraction selection 1/1, holdout 1/1, separate native inputs 2/2. [Consumer](consumer6781.json), [subtraction](difference6763.json). |
| [6841](producer6841.json) | Continuation record and metadata limit | Original fill locations retained; over-limit metadata reports an issue. Static record construction only. [Log](observation6841.log). |

The CLI used producer 6781 to read the saved 6373 result. JSON returned three
static origin rows, a candidate filter returned one, and text/help succeeded.
[CLI observation](ended6789.json) records unchanged input bytes and binary removal.
The first JSON invocation took 580.6 ms; candidate/text reads took 12.6/10.9 ms.
These are individual invocations, with different outputs and startup conditions.
They are not a latency distribution or a whole-machine CPU measurement. Program
and model executions during these reads were zero.

The final module is producer 6841. Consumer/CLI native observations belong to their
listed earlier producers; the submitted revision gets its own CI.

## Original CI failure and correction

PR #1535 added the corpus entry but omitted the matching native registry,
counters, and denominator migration. [Registration repair](registration-repair.json)
adds the case to all required representations: 82 corpus cases, 80 language
capability cases, denominator version 48. Previous denominator JSON stays intact.

The earlier PR description attributed integer-loop rewrites to the formatter.
The [original CI log](original-ci-plan-failure.log) shows that `go fix ./...`
required them. This successor applies the modernizer changes; its source-only
`go fix -diff` inspection was empty. The original failed head, body and runs were
not rewritten or rerun.

## Scope

Paths represent static may-dependencies in the selected source. Branches are
joined; recursion, loop approximation and bounded output retain partial reasons.
An unused argument may still be evaluated at runtime, even when its value does
not influence the return. Rows do not assert successful execution of every path.
The four consumer facts have four path/location pairs from three origin rows.
All observations here used zero model/provider calls and zero synthetic facts in
the consumer. Whole-language completeness and independent adoption are unmeasured.
