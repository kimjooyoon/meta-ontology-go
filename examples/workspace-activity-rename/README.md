# Rename an activity across a declared workspace

The sources use the same `Review` spelling in two packages. Two files also use
the same `rules` alias for different imports. The target's stable ID distinguishes
the intended declaration and references.

With a compiler containing workspace rename support, run from this directory:

```sh
gooo package rename-activity --package review/rules --id urn:review:rule --to 검토 --json workspace.json
```

The command prints a source proposal. It updates the target declaration, a local
call, the app's escaped call, a fixed criterion, a binding and a consumer outside
the entry closure. The peer package, its consumer and ordinary `"Review()"` text
retain their bytes. The original files remain intact.

`result.files` contains edited source, ranges and before/after hashes.
`updated_workspace` retains source paths and carries any changed entry name.
The nested runtime manifest includes all supplied source content.

## Development observations

[observation.json](observation.json) records three separate new source profiles:

- The CLI route updated 5/5 resolved target references in three of five files.
- An imported fixed oracle retained its ID through a rename, answered one source
  question, and selected one surviving candidate. A reserved post-selection input
  matched 1/1. Two ranking records describe the question and the final survivor;
  two oracle evaluations cover the original case and the acquired answer.
- An implicit entry retained its original semantic ID by adding it to the source.
  The entry and its one dormant consumer both adopted the new name.

A fourth profile used the compiled CLI once in plain output mode. From
[`plain-cli`](plain-cli), the command is:

```sh
gooo package rename-activity --package reward/core --id reward://activity/credit --to 적립 workspace.json
```

The actual development invocation reported reference 1/1, two changed files,
the retained target ID and an updated entry. Original files remained intact.
This is one CLI process observation, with zero source program executions and
zero model predictions. Its single process wall interval includes startup and
parsing; CPU/RAM usage and comparative latency were not measured.

Producer revisions and exact source overlays belong to each observation. The
published fixture paths add `.fixture` to the test source names; source contents
and their original hashes are retained. Local historical tests and model cohorts
were not repeated. Failed development observations are listed separately.

Rename executes no program and makes no model prediction. The oracle profile
uses finite AST evaluation, with zero native processes and zero model predictions.
There is no comparative performance or fitted-model accuracy measurement here.

See the [Korean guide](../../docs/language/workspace-activity-refactoring.ko.md)
for scope, unsupported call surfaces and the source proposal contract.
