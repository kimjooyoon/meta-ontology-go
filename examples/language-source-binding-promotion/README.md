# Source binding promotion

This example promotes one bounded claim only: the selected source-execution receipt is bound to the declarations in its `.gooo` source.

The policy is declared in `examples/self-improvement/main.gooo`. CI generates it twice, compares the generated artifacts, and binds that artifact digest to the promotion receipt.

The three claims remain in every case. Missing or unknown evidence leaves claims `OPEN`; contradictory links make the promotion `REFUTED`; only linked structural and independent evidence makes it `DISCHARGED`.

This does not claim full compiler correctness, an independent toolchain implementation, value-level computation, or repair of the producer validator.
