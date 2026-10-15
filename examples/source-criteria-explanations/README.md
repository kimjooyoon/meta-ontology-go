# Explain why the declared choices did not work

A quote must cover a cost of 18 while staying within a budget of 25. The Gooo
source supplies two choices: the original input and that input plus 20.
For an input of 10, the possible prices are 10 and 30.

| Candidate price | Covers cost 18 | Within budget 25 |
| --- | --- | --- |
| 10 | false | true |
| 30 | true | false |

Both source conditions can be satisfied by a price of 20. The declared choices
leave that value out. The useful explanation therefore names the condition
rejecting each candidate and keeps its conclusion within this candidate set.

Gooo records the conditions in the assembling activity:

```gooo
criteria {
 predicate "within-budget" activity "WithinBudget"
 predicate "cost-covered" activity "CostCovered"
 input "[10]"
 holdout_input "[11]"
}
```

The fixed Gooo activities define the comparisons. `main.gooo.fixture` is the
complete source. `expanded.gooo.fixture` replaces the second expression with
`input+10`, permitting price 20 at the initial input. The number of choices stays
two; the source conditions stay unchanged.

## Recorded construction and separate evaluation

| Observation | Original alternatives | Revised second alternative |
| --- | --- | --- |
| Construction decision | `PARTIAL_FINITE` | `COMPLETE_FINITE` |
| Selected local criteria | 1/2 | 4/4 |
| Question-history candidates rejected / explained | 2/2 | 1/1 |
| Caller selection output | 0/1 for both attempts | 1/1 on the first attempt |
| Final held criteria | 1/2 | 2/2 |
| Separate native output at input 11 | 0/1 | 1/1 |
| Model calls | 0 | 0 |

The revised source also acquires question `[12]`, so its two predicates are
checked at two inputs: four predicate/input pairs. The different denominators
describe different acquired obligations. They are finite observations of this
program. The source revision in this example was authored directly.

The report identifies `cost-covered` for candidate 0's price 10 and
`within-budget` for candidate 1's price 30. After the source change, it identifies
the rejected price-10 candidate while retaining the surviving alternative.

## Read the explanation without another execution

From a checkout containing this implementation, create a new result:

```sh
go run ./cmd/gooo body-construct \
  --source examples/source-criteria-explanations/main.gooo.fixture \
  --entry Price \
  --construction-cases examples/source-criteria-explanations/construction-cases.json \
  --cases examples/source-criteria-explanations/evaluation.json \
  --attempts 2 > /tmp/gooo-quote-result.json
```

Then read its stored observation:

```sh
go run ./cmd/gooo body-construct --report /tmp/gooo-quote-result.json --format text
```

With a built `gooo` binary, the report command needs no Go toolchain. Reading the
saved report makes zero model calls and zero program executions. The saved JSON
bytes remain unchanged. The report labels its recorded claims as unverified;
reading a record does not re-establish its source or execution evidence.

Each rejected candidate contributes one linked decisive failure, with the source
criterion's identity, input, actual output and recorded evaluation indices. Exact
value-case failures are included when the stored expectation is available. The
human report shows up to six reason groups and states how many further groups
remain in the recorded question evaluations.

Unexamined candidates, unavailable outputs and missing reason links remain
separate. After a predicate edit, the explanation still refers to the initial
question history and its source digest; current selected criteria have their own
report rows. The explanation does not modify the source or select a new body.

`observation.json` binds both once-recorded native runs and their saved reports.
The public reports retain both the failed example and the revised source's
results. These measurements establish neither general input coverage nor a
model-quality improvement.
