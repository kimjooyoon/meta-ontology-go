# Rename an activity and keep its source contract connected

`main.gooo.fixture` describes a fee calculation, a fixed Boolean requirement and
a selecting activity. The selector owns two expression choices and separate
selection/evaluation inputs. This example changes `Calculate` to `요금`.

```sh
gooo refactor rename-activity --from Calculate --to 요금 main.gooo.fixture > renamed.gooo
gooo refactor rename-activity --from Calculate --to 요금 --json main.gooo.fixture
```

The existing ID `pricing://activity/calculate` is pinned on the renamed
declaration. The requirement's call, the selector's body call and its alternative
expression use `요금`. The authored comment and intent still say `Calculate`;
their prose has no binding authority. Criterion IDs and finite inputs stay intact.

`rename.json` is a saved source transformation; `renamed.gooo.fixture` is its
output. `observation.json` binds those files to the exact local development build.
The observation runs the new rename command once with an empty executable search
path. It executes no Gooo program, calls no model and retrains nothing.

Renaming `Check` to a new name would update the `criteria` reference as well.
The source regression checks cover that second transformation. This example's
saved receipt records only the `Calculate` transformation.

`escaped.gooo.fixture` adds a local string containing literal `\b` and `\f`
sequences. It exposed a quoting bug during the implementation review: replacing
JSON control escape spellings could also modify an already escaped backslash.
The source writer now encodes Gooo string tokens directly.

`escaped-rename.json`, `escaped-renamed.gooo.fixture` and `escaped-observation.json`
record a new transformation with that correction. The decoded body changes only
at its one activity call name; the string value stays intact. The first receipt
keeps its original compiler identity, and the escaped-string receipt identifies
the corrected development build. Neither observation executes the resulting
Gooo program.

`graph-identity-check.json` records a separate source-only integration check with
the local Unicode identifier correction recorded at SDK revision `4de73fd`.
Its public successor is decision-runtime PR #78, which also retains the extraction
history. The integration check renames a helper,
a fixed predicate and the assembly owner to Korean names. Stable IDs, ordered
predicate nodes, input/result roles and candidate source-feature bytes stay
identical. The source hash changes. The complete candidate-plan digest can change
because it records the source name, while its structural feature input retains
the same meaning. No model prediction or program execution was performed.

See [the scope and reference rules](../../docs/language/activity-refactoring.md)
before renaming an activity used by another file or package.
