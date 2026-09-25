# Operation manifest experiment

The package is a Gooo-only definition: two `.gooo` files and zero `.go`
definition files. The compiler can project its selected activity as a
versioned, non-Go operation manifest:

```sh
gooo emit --kind operation-manifest --entry PayOrder examples/billing-package
```

The adjacent contract and golden file close one small experiment. A passing
receipt means only that this fixed projection was observed. It does not claim
business correctness, production readiness, general code generation, or
performance outside the recorded runner samples.

The emitter registry contains three kinds. That count is an extension-surface
measurement, not a language-quality score.
