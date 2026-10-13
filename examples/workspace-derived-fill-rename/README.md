# Keep helper identity through derived slot construction

This workspace derives two condition expressions and two value expressions,
retaining four complete assignments. A template calls the imported `Base`
activity, and a dormant consumer calls the same activity by its file-local alias.

With a compiler containing versioned derived-fill rename support:

```sh
gooo package rename-activity --package derive129/helpers --id urn:derive129:base --to 기본값 --json workspace.json
```

The command returns source edits and an updated manifest, keeping original files
intact. Current supported fill grammars create values and operators without new
activity calls or lexical bindings. Template calls keep their original scope;
renaming requires no candidate enumeration, scoring, model or program execution.
Unknown grammar versions require a candidate reference-scope map.

## Distinct development observations

[observation.json](observation.json) retains each actual producer and result.

- A new per-hole profile updated 2/2 references across three files, then used the
  dedicated fill API on the renamed source. Selected cases matched 1/1, and the
  reserved post-selection input matched 1/1, with zero model/native calls.
- The condition grammar retained 2/8 expressions and the value grammar 2/5.
  Their combined expression coverage was 4/13 (about 30.8%); all 4/4 assignments
  of those retained expressions were constructed. The recorded grammar is
  incomplete. Assignment coverage does not include omitted expressions or
  unstated intent.
- A separate shared-grammar workspace used the compiled CLI once. From
  [`actual-cli`](actual-cli), use the helper ID `urn:derive148:add`, package
  `derive148/helpers`, new name `가산`, and `--json workspace.json`. It proposed
  updates to 1/1 reference in two source files, with zero enumeration/scoring/
  model/program execution. Original files remained intact. The temporary
  25,657,570-byte executable was hashed and removed. Its single process wall
  interval was 563,928,667 ns including startup/parsing; comparative CPU/RAM,
  latency and inference speed were not measured.
- A distinct search source obtained the dedicated consumer diagnostic before
  its unfilled body reached ordinary projection. No source code was emitted.
- The reference classifier accepted 16 known grammar versions and rejected
  three absent/unknown profiles. This checks the classifier contract; it is not
  an execution test of every grammar production.

Two earlier distinct tests failed: one chose the scalar/record consumer API for
fill construction; another wrote the search clause fields in the wrong order.
Those producers and failures remain saved. Corrected historical local profiles
were not replayed. The successful profiles have different source identities.

Fixture paths add `.fixture` to the original source filenames. The contents and
producer hashes retain their original identity. There is no training, competitor
execution, all-domain functional proof or complete PROV equivalence claim.

See the [Korean guide](../../docs/language/workspace-derived-fill-refactoring.ko.md).
