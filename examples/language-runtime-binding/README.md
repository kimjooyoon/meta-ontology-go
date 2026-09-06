# Language runtime binding execution

This fixture is the minimal value-plan execution example. The `Produce`
activity receives an explicit integer root input, applies the registered
`int.add:1` implementation once, and the two explicit `bind` edges deliver
that sealed result to `ConsumeA` and `ConsumeB`. Each consumer applies the
same registered operation once, yielding `43` from an input of `41`.

The plan records three real `Apply` calls and two real token deliveries. The
generated-Go and source-execution boundaries remain unsupported for runtime
bindings; this example exercises only the native value plan path.
