# Gooo language test experiment

This example closes one development loop using only `.gooo` declarations.

```gooo
entity PayOrderProducesPayment id "gooo://test/activity/PayOrder/output/Payment"
```

`gooo test examples/language-test/main.gooo` discovers the marker, executes
`PayOrder`, and compares the produced entity name and stable ID with `Payment`.

The experiment does not claim value-level assertions, side-effect assertions,
external dependency execution, or a final dedicated test syntax. The marker is
a versioned metaprogramming seam for measuring whether those extensions earn
their complexity.
