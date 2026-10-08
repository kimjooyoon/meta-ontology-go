# Keep an input-dependent fit inside a small candidate cap

This example has ten construction inputs. Each requires a different value for
the expression hole. The v1 quadratic grammar enumerates each input's constant
roots first, so its 16 retained candidates omit the affine expression that can
satisfy every input. The best retained v1 body matches one of the ten cases.

`integer-hole-quadratic/v2` uses the same candidate universe. Before applying the
cap, it places root-derived constants and affine fits compatible with all
available integer root observations first, in a deterministic order. Remaining
v1 expressions follow in their original order. An unavailable or non-integral
probe supplies no root constraint. Full-body scoring still checks the source's
actual computation, including branches and higher-degree expressions.

The original source declares v2 as an alternative. The existing Gooo policy
selects it after the current attempt allowance is spent:

```sh
gooo body-refine --source examples/quadratic-fit-feedback/source.gooo.fixture \
  --activity Build --feedback-cases examples/quadratic-fit-feedback/feedback-cases.json \
  --evaluation-cases examples/quadratic-fit-feedback/evaluation-cases.json \
  --policy examples/search-feedback/policy.gooo.fixture --search-policy \
  --max-attempts 8 --max-rounds 4 --out out/quadratic-fit-feedback
```

Both grammars enumerate 108 expressions and retain 16. The v1 round matches
1/10 feedback outputs. The v2 round matches 10/10, then the separately evaluated
inputs `0`, `11` and `-2` match 3/3. The selected factor is `input * -2 + -1`;
squaring it gives the requested adjustment. Two rounds complete this example.

The candidate coverage stays 16/108 even after the supplied functional cases
pass. The grammar still reports truncation. A three-point fit supplies ordering
hints; the cubic regression demonstrates that compatible fitted roots can fail
the original body. Holdouts stay outside candidate construction and ordering.

The named v1 grammar retains its original enumeration for saved-record replay.
The compiler regression suite replays the earlier published v1 native program
and checks its three evaluation outputs with zero model calls.
