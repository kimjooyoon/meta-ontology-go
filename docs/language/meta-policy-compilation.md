# Meaning-preserving compilation of a Gooo meta-policy

This experiment asks whether a policy declared in Gooo can be compiled into a
deterministic decision kernel while retaining enough lineage to show that the
source, compiled artifact, generated execution, and independent execution made
the same decision. It is deliberately narrower than “support another policy
DSL”.

## Source, compiler, and proof boundary

`examples/meta-policy-compilation/policy.gooo` is the authority for the policy
steps. Each of its eight activities has a `computes` value carrying the
semantic fields `role`, `meta-operation`, `proof-choice`, `stage`, `step`,
`reason`, and `claim`. The compiler lowers this file through the semantic IR,
checks those fields against the fixed experiment ontology, and sorts by `step`.
Changing declaration order therefore cannot change the compiled meaning;
changing an activity, metadata value, or stable source identity is rejected.

The generated judge is a standalone Go program. It receives only structured
evidence and returns `PASS`, `FAIL_CLOSED`, or `UNKNOWN` plus a stage/step/reason
coordinate. The independent verifier is a separate implementation in the
repository package. A consumer replays both and compares the full decision
coordinate, not only the decision label.

The fixed denominator is eight policy obligations. The case denominator is
three: one pass, one fail-closed source drift, and one unknown missing-consumer
case. Every case gets eight `UNRECORDED -> OPEN` claim registrations followed
by eight persistent outcome transitions. The receipt therefore contains exactly
48 chained events, each carrying the previous event digest; no later event can
silently rewrite an earlier assertion.

## Research basis and limits

Two official bodies of work provide useful principles, but neither is a proof
of this experiment:

1. [Open Policy Agent documentation](https://www.openpolicyagent.org/docs) treats
   policy as declarative code and separates policy decision-making from policy
   enforcement. Its [policy-language documentation](https://www.openpolicyagent.org/docs/policy-language)
   emphasizes decisions over structured input. This motivates the producer /
   consumer split and the explicit decision coordinates here. OPA's
   [bundle documentation](https://www.openpolicyagent.org/docs/management-bundles)
   also describes bundle activation as eventually consistent and notes that
   signature verification has mode-specific limits. This experiment therefore
   binds local source and artifact digests, but does not claim freshness,
   authenticity, or production enforcement.
2. The INRIA [CompCert compiler proof documentation](https://compcertssa.gitlabpages.inria.fr/html/compcert.driver.Compiler.html)
   states semantic preservation through composed simulations, and its
   [correctness overview](https://compcertssa.gitlabpages.inria.fr/html/compcert.driver.Compiler.html)
   distinguishes a source semantics from the generated target semantics. This
   motivates comparing source and target decisions through an explicit modeled
   semantics. The limit is decisive: this repository has no machine-checked
   theorem for arbitrary Gooo policies. A digest match and two deterministic
   executions are evidence of this modeled contract, not a universal
   compiler-correctness theorem.

The experiment is falsifiable. It fails closed if the fixed denominator changes,
an activity's semantic metadata changes, the producer and artifact source
digests differ, the independent verifier disagrees with the generated judge,
the expected case result changes, or any claim event is edited, removed, or
reordered. It remains `UNKNOWN` when required evidence is unavailable. A future
extension would need to add a new semantic rule and proof obligation rather
than silently widening this ontology.
