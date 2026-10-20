# Assemble one byte-search condition

The Gooo source owns the loop, length check, return values and two allowed
predicates. Only `match` is left open. The first candidate uses `<`; the second
uses `==`. Selection checks six source-declared cases, then reports four held-out
cases separately. No model file or API key is required.

Use the public Gooo 0.6.32 compiler containing the one-hole grammar. The
[0.6.32 guide](../../docs/releases/0.6.32-dev.md) identifies the actual source,
platform checks and published files. The earlier 0.6.31 compiler predates this grammar.

From the compiler checkout with Go 1.27.3:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.3 go build -o ./gooo ./cmd/gooo
./gooo package execute --json \
  --cases examples/single-condition-byte-search/cases.json \
  examples/single-condition-byte-search/gooo.workspace.json > byte-execution.json
./gooo package replay --json --receipt byte-execution.json \
  --cases examples/single-condition-byte-search/cases.json \
  examples/single-condition-byte-search/gooo.workspace.json
```

The ten external rows repeat the source's six selection and four held-out inputs.
They verify package execution and saved replay; they add no independent unseen
inputs. The release reader compares every result with Go `strings.IndexByte` and
binds the committed manifest, cases, source graph, generated program and exact
original receipt. Text remains valid UTF-8 and at most 1,024 bytes. Offsets count
bytes; a needle outside 0..255 returns -1.

The retained local preparation first proposed `lower`, matching 1/6 selection
cases. The selected `equal` matched 6/6 selection and 4/4 held-out cases; native
execution and saved replay each matched 10/10. Model calls were zero. These finite
counts describe this declared operation, and no intent-wide completeness or
neural-model accuracy was measured.
