# Source cycle lifecycle

`policy.gooo` owns the distinction between bound, interrupted and unresolved
cycle observations. The host supplies the original CI conclusion and whether
all existing artifact/identity/link/metric obligations passed. The generated
library is imported by `scripts/self-improvement-cycle` without a runtime
compiler, model, external service or new wire format.

Regenerate with a development compiler supporting package library generation:

```sh
gooo package generate --go-package cyclecontext --out internal/meta/cyclecontext \
  internal/meta/cyclecontext/gooo.workspace.json
```

The checked-in file retains its actual compiler producer and input digests. A
source-consistency test reuses the common workspace loader and library generator;
independent cases check lifecycle behavior. Original source-CI cancellation does
not invalidate its available observations, complete missing work, or permit
downstream execution. The public envelope keeps the original conclusion and
reports `INTERRUPTED` only when the existing bindings are valid.
