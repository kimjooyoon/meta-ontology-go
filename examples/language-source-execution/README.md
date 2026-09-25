# Language source execution

The fixed four-case contract executes `examples/billing/main.gooo` through the
real `gooo run` command. It proves one symbolic activity transition, one exact
byte replay, and two fail-closed diagnostics.

Gooo activities currently describe typed ontology transitions. The receipt
scope is `DECLARATION_RESOLUTION_ONLY`: it binds the declared input entities
and records the selected activity and declared output entity in four events.
It does not claim registered-value operation execution, handwritten Go-body
execution, external effects, multi-file execution, or external dependencies.
