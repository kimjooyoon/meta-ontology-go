# Language artifact oracle

The fixed four-case contract compares a real `gooo run` artifact with an
independent, bounded projection of `examples/billing/main.gooo`.

The oracle imports neither the compiler parser nor the source-execution receipt
package. It accepts one genuine artifact, rejects one digest-valid output
forgery, rejects an unknown artifact decision, and lowers resolution when the
source is outside its declared grammar. It does not claim full compiler
semantic correctness or general parsing.
