# Claim resolution tuple

`claim.resolve:v1` is a minimal Gooo value-program operation derived from three
independent consumer releases. It does not introduce a second parser or a general
policy language.

An activity declares exactly six fields:

```gooo
activity Resolve(Claim) -> Claim computes "claim.resolve:v1;state=UNKNOWN;stage=SOURCE;step=OBSERVE_INPUT;reason=INPUT_UNAVAILABLE;unknown_class=DIRECT_MISSING;next_operation=PROVIDE_INPUT"
```

The allowed states are `CLOSED`, `UNKNOWN`, and `REFUTED`. `NONE` is the explicit
source spelling normalized to JSON `null` for `stage`, `step`, and
`unknown_class`.

- `CLOSED` requires `NONE` stage, step, unknown class, and next operation.
- `UNKNOWN` requires stage, step, reason, unknown class, and next operation.
- `REFUTED` requires stage, step, reason, and next operation while unknown class
  must be `NONE`.

The core command is:

```text
gooo claim resolve <file.gooo> --activity <name> --json
```

Valid UNKNOWN is returned without being converted to false or fixed point. An
unrecognized state such as `FIXED_POINT`, an incomplete UNKNOWN tuple, a missing
program, or non-unique activity cardinality returns a JSON `FAIL_CLOSED` report
and a non-zero command exit.

Every indicator names the source Gooo activity. The operation authorizes no core
mutation, repository write, dependency propagation, external truth claim, or
automatic merge.
