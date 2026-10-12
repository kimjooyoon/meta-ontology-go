# An Integer function chooses a path using its own conditions

The goal is to subtract eleven from an input. At input 11, either operand order
returns zero. The source declares input 17 as a question and a Gooo Boolean
activity as the criterion. No expected result is supplied for either criterion
input. The compiler evaluates each candidate's actual Integer result.

- `main.gooo.fixture` can choose between `11-input` and `input-11`.
- `insufficient.gooo.fixture` only swaps the outer operands of `(11-input)+0`.
  Both alternatives have the same wrong behavior away from input 11.
- `construction-cases.json` checks the caller at input 23.
- `evaluation.json` supplies an additional native evaluation at input -31.
- The source reserves input 20 for held criterion evaluation after selection.

The signature stays `Integer -> Integer`. The full declaration and supported
choices are described in [the language guide](../../docs/source-scalar-criteria.md).

## Recorded results

Each variant was constructed and evaluated once using a fresh development
compiler. Both acquired question 17. The question strategy first looks for
output disagreement, then checks an observable shared answer against the
criterion if needed.

| Observation | Correct path available | Both paths have the same error |
| --- | --- | --- |
| Selected criterion/input pairs | 2/2 | 1/2 |
| Question outcome | One surviving candidate | No surviving candidate |
| Held criterion at 20 | 1/1 | 0/1 |
| Caller output at 23 | 1/1 on first attempt | 0/1 on both attempts |
| Additional native output at -31 | 1/1 | 0/1 |
| Construction result | `COMPLETE_FINITE` | `PARTIAL_FINITE` |
| Model calls | 0 | 0 |

These counts describe the named criterion and finite inputs. In the second
program, the report links both rejected candidates to the false criterion at
17. That identifies a limitation of its declared choices. Additional choices or
changed requirements would create a new source contract.

[The successful report](positive-report.md) and
[the insufficient-choice report](insufficient-report.md) were rendered from
saved results without executing the programs or a model again. Their original
unverified-record warning is preserved. [Observation metadata](observation.json)
binds compiler, source, case, output and report digests. Exit zero records that
the operation finished; the functional results are shown separately above.

After these observations, an unchanged record-only validation block was moved
to a helper to keep the Go projection within its supported function size. The
scalar path returns before that block. Existing record compatibility and the
changed-source projection probe passed after the extraction. The observed
binary and manifest retain their original digests.

## Try it with a development build

From the repository root:

```sh
gooo body-construct \
  --source examples/source-scalar-criteria/main.gooo.fixture \
  --entry Fee \
  --construction-cases examples/source-scalar-criteria/construction-cases.json \
  --cases examples/source-scalar-criteria/evaluation.json \
  --attempts 2
```

Use `insufficient.gooo.fixture` for the second program. This command performs a
new construction and native evaluation. The saved Markdown reports can be read
on their own. Performance and model accuracy were not measured in this example.
