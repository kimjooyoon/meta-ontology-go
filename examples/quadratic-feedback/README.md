# Construct a value used twice in the body

This activity keeps negative inputs unchanged and adds a squared adjustment to
other inputs. The source includes a branch, local values and one expression hole.
The expected adjustment is nine. A zero/one residual proposal guesses nine for
the hole, and the full body rejects that guess because its square is 81.

The new `integer-hole-quadratic/v1` grammar evaluates the body with the hole set
to -1, 0 and 1 on each construction input. Exact integer arithmetic fits a
quadratic to those observations and proposes its integer roots. Both -3 and 3
produce the requested square in this example. Full-body scoring chooses a value
from the retained candidate list.

The original source declares this grammar as an alternative. The existing Gooo
feedback policy can move to it after reaching its current attempt allowance:

```sh
gooo body-refine --source examples/quadratic-feedback/source.gooo.fixture \
  --activity Adjust --feedback-cases examples/quadratic-feedback/feedback-cases.json \
  --evaluation-cases examples/quadratic-feedback/evaluation-cases.json \
  --policy examples/search-feedback/policy.gooo.fixture --search-policy \
  --max-attempts 8 --max-rounds 4 --out out/quadratic-feedback
```

The example policy advances to `quadratic`, then stops after all three feedback
outputs match. The three final inputs are evaluated after source selection.
The selected source/composition can replay with zero inference. Model-free
candidate enumeration and the Gooo decision policy are deterministic.

## Candidate and observation boundaries

- The grammar accepts 1..128 scalar construction cases and performs at most
  three body probes per case. Holdout cases do not enter the probes or candidates.
- It enumerates constant roots, affine fits between roots of the first usable
  input and subsequent inputs, input offsets, then residual-grammar seeds.
  Repeated expressions are removed in that order. At most 16 candidates remain;
  the receipt records the full enumerated count and any truncation.
- `gooo/integer-hole-context/v2` adds the observed minus-one output and fitted
  roots. Existing residual-grammar receipts retain their v1 representation.
- A probe can fail, have no observed sensitivity, or yield no integral root.
  These observations stay in the receipt; residual candidates remain available.
- Branch changes and higher-degree arithmetic can disagree with a three-point
  fit. The original typed body checks every attempted candidate. A square that
  must equal five and the supplied cubic regression stay partial in the tests.

The fit proposes values for a declared region of the program. The source owns
the body, permitted grammar and expectations. Finite scores describe the supplied
inputs, including when a polynomial proposal fails their checks.
