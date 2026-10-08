# Continue after a rejected fill assignment

The five assignments preserve the three valid assignments from
`../caller-source-fill`, adding a Boolean in an Integer position and an expression
that divides by zero on the local training case. All original cases are unchanged.

```sh
go run ./cmd/gooo body-construct \
  --source examples/caller-fill-rejection/budget.gooo.fixture --entry Main \
  --construction-cases examples/caller-fill-rejection/construction-cases.json \
  --cases examples/caller-fill-rejection/evaluation-cases.json \
  --attempts 5 --out /tmp/gooo-fill-rejection-example
```

Use a fresh output directory. Local preparation scores three assignments and
records two rejections. Whole-program search retains all five assignments, with
rejections consuming its program budget. Rejected attempts have no caller
execution or local score; already scored earlier activities remain observable.
The separate local holdout does not rank candidates. The final caller evaluation
retains the original four expectations, including the integer above 2^53.

The mixed fixture adds three record choices, giving 40 declared combinations.
An optional `--model` orders those record choices. The fill assignments use
deterministic selection and source order independently of that graph model.

These records use `gooo/joint-construction/v5`. Valid fill histories without
rejections keep v4. Saved replay reconstructs the rejected assignments and their
reasons without new inference. This bounds a finite source-owned search; it does
not establish behavior outside the supplied cases.
