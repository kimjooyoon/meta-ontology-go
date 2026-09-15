# Claim resolution tuple

`gooo claim resolve <file.gooo> --activity <name> --json` resolves the exact
`claim.resolve:v1` value program carried by one Gooo activity.

The six source fields are `state`, `stage`, `step`, `reason`, `unknown_class`,
and `next_operation`. Allowed states are `CLOSED`, `UNKNOWN`, and `REFUTED`.
`UNKNOWN` requires every explanatory field. An unrecognized value such as
`FIXED_POINT` fails closed rather than being treated as success.

The candidate id `gooo.primitive.claim-resolution-tuple.v1` occurs in the Gooo
source output entity and every JSON receipt. Indicators retain the selected
activity, so the metric cannot detach from its meta operation.

The implementation claims no dependency propagation, external truth, core
mutation authority, repository write, or automatic merge.
