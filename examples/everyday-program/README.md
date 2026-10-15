# Variables, a condition and the origins of the result

This is a small ordinary Gooo program. It adds 11 to an integer input and doubles
the result when the Boolean input is true. `main.gooo` owns those operations:

```gooo
activity Adjust(Integer, Boolean) -> Integer id "urn:gooo:everyday:adjust" computes `let value=input0+11
if input1 {value=value*2}
return value`
```

`let` introduces a local. `input0` and `input1` are the activity's declared input
ports. A write such as `value=value*2` assigns the next local value. The activity's
stable ID identifies this behavior independently of its display name.

## Run the program

Use a compiler and Go toolchain matching the source you are working with. From
this directory:

```sh
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
gooo body-origins --activity Adjust main.gooo > origins.json
```

The first command generates Go, executes three declared finite rows and saves a
receipt. The second reads possible value and control dependencies from the source.
The source inspection contains no execution inputs. It cannot identify which
branch ran for an arbitrary input.

| Integer input | Boolean input | Declared expected result |
| ---: | --- | ---: |
| 5 | true | 32 |
| -7 | false | 4 |
| 0 | true | 22 |

Change the inputs or declaration to make a new program observation. Keep the
source with its receipt: an edited source needs a new receipt.

## Recorded development observation

The [source manifest](observations/manifest.json),
[compressed receipt](observations/execution.json.gz) and
[summary](observations/summary.json) record one new source/input cohort compiled
from `10067b26de6d3fa75040173a610fdc522b5da62a` with Go 1.27.2 and SDK
`v0.2.51-experimental`. It was a development build with version label
`0.6.34-dev`; the [build metadata](observations/binary-build-metadata.txt.gz)
identifies the actual producer. It was not the public v0.6.34-dev archive.

All three expected outputs passed. Model calls were zero. The whole execution
command took 0.6577 seconds once on the recorded machine. This is a single
observation, with no throughput or performance comparison. The compiler build
and source inspection are separate commands.

The [origin report](observations/origins64317/origins.json) resolved 16 graph nodes,
returned root 16, and retained both declared input identities and original source
byte ranges. The Boolean is a control dependency; the integer supplies the data
being adjusted. Parent and control edges describe possible source dependence.

These three rows measure the declared finite examples. All-input correctness,
overflow behavior and the caller's complete intent are wider requirements.
This walkthrough uses existing language features and adds no denominator claim.
