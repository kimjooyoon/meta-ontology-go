# Rename references inside source-declared fill candidates

The workspace declares three complete assignments to two holes. `Boost` occurs
in the body template, two candidates and a consumer outside the entry closure.
An unrelated block declares a local `fees` value; it does not own the later calls.
The peer package declares a different activity with the same display name.

With a compiler containing declared-candidate workspace rename support:

```sh
gooo package rename-activity --package scope83/helpers --id urn:scope83:boost --to 가산 --json workspace.json
```

The command returns edited source and an updated workspace manifest. Original
files remain intact. Each candidate is read in the template's lexical scope.
Resolved calls map to the original template or fill literal byte ranges; repeated
template references count once. This process performs no candidate scoring,
model prediction or program execution.

[observation.json](observation.json) retains the exact development producer:
public compiler `ae957835` plus six new/changed source overlays.

- A new three-candidate profile updated 6/6 references across three of four files.
  The two holes, all candidate labels and same-name peer source were retained.
- A different one-hole profile found a local value shadowing the import alias at
  the actual fill position, returning an error and zero partial proposal files.
- An earlier separate profile failed at workspace build because its manifest
  omitted the authored import. Its failure is retained; it is not a successful
  rename observation and was not repeated locally after correction.

A third, distinct coupons workspace called the compiled CLI once in JSON mode.
From [`actual-cli`](actual-cli), use:

```sh
gooo package rename-activity --package coupons106/helpers --id urn:coupons106:add --to 추가 --json workspace.json
```

The development invocation proposed updates to 3/3 references in two source
files and retained the original files. It used one CLI process, with zero model
predictions, candidate scoring and source program executions. Its single process
wall interval was 673,132,292 ns, including startup and parsing. Comparative
latency, CPU/RAM usage and inference speed were not measured. The temporary
25,657,282-byte executable was hashed and removed after this observation.

The fixture filenames add `.fixture` to the original test source names. Source
contents and original producer identities are preserved. These are source
reference observations, with zero model predictions, candidate scoring and
program execution. Generated `derive` candidates and `search` require separate
scope maps. Renamed candidate behavior, native execution and editor application
need their own observations.

See the [Korean guide](../../docs/language/workspace-source-fill-refactoring.ko.md).
