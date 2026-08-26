# Compiler-projected conformance vectors

This experiment adds two structural vectors to the compiler-emitted symbolic
invocation artifact. The vectors are derived from the same operation model as
the JSON Schema; they are not maintained as runtime fixtures.

## Fixed measurements

| Measurement | Before | Target |
| --- | ---: | ---: |
| Compiler-generated validation vectors | 1/2 | 2/2 |
| Generated ACCEPT vectors | 1/1 | 1/1 |
| Generated REJECT vectors | 0/1 | 1/1 |
| Handwritten runtime vectors | 1 | 0 |
| Repository writes | 0 | 0 |
| Mutation authorities | 0 | 0 |

`accept-exact` is a FOUNDATION projection of the declared activity and ordered
input identifiers. `reject-missing-activity` is a REGRESSION projection that
removes the schema-required activity coordinate. Each vector names the compiler
meta-operation that produced it.

The first increment is intentionally `STRUCTURAL_ONLY`. It does not claim that
an external validator observed either verdict. External validation, value-level
execution, domain correctness, production readiness, and generalized
performance remain outside this experiment until separate CI evidence exists.
