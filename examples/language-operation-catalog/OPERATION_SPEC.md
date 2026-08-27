# Typed operation specification

`main.gooo` keeps the existing `activity ... computes "..."` syntax. The
compiler lowers each accepted value program into two closed semantic objects:

- `OperationSpec` describes the registered operation independent of a call.
- `OperationIR` binds one activity and typed operand to that specification.

The first catalog contains one operation, `int.add` version 1. It accepts one
`Integer`, parses an `INT64_LITERAL`, returns `Integer`, performs a
`PURE_VALUE` effect, and is `DETERMINISTIC`. Its runtime failure set contains
only input arity mismatch and signed integer overflow. Repository writes,
external calls, and promotion authority are all false.

## OS9 fixed denominator

The CI receipt reports exactly nine indicators:

1. one canonical specification is resolved;
2. operation identity and version are explicit;
3. input and output entities are typed;
4. the literal operand kind is typed;
5. the effect class is explicit;
6. determinism is explicit;
7. runtime failures form a closed set;
8. operation authority is zero;
9. every present invocation binds the canonical specification.

Each indicator has one durable claim. Evidence changes a claim from `OPEN` to
`DISCHARGED`; it does not delete the claim. An unresolved operation records one
`stage/step/reason` coordinate and leaves all unsupported claims open.

This example does not claim arbitrary operations, general expressions, other
value types, runtime performance bounds, or authority to mutate the repository.
