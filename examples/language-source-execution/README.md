# Language source execution

The fixed four-case contract executes `examples/billing/main.gooo` through the
real `gooo run` command. It proves one symbolic activity transition, one exact
byte replay, and two fail-closed diagnostics.

Gooo activities currently describe typed ontology transitions. Execution binds
the declared input entities, invokes the selected activity, and produces the
declared output entity in a four-event receipt. This does not claim value-level
computation, side effects, multi-file execution, or external dependencies.
