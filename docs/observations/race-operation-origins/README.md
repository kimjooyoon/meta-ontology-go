# Race operation origins

PR1569's original race job115961481366 was cancelled by its20-minute job limit.
Its annotation gives that exact reason. The final cleanup names extractor.test;
the original ordinary unit job115961481562 passed, including extractor237.681s.
The race log contains180 completed ok-package lines; the unit log contains244.
These are package observations from different jobs. They are not case accuracy,
controlled speed ratios, proof of deadlock or a diagnosis of one slow test.

The prepared workflow adds the Go test JSON output flag. The full ./... scope,
race detector, canonical check name,20-minute limit, events, permissions and all
other workflow declarations remain unchanged. JSON logs identify individual
package/test start, completion and output events, including active tests when a
job ends early. Go1.27.2 documents this flag in cmd/go/internal/test/test.go.
No repository test branches on testing.Verbose; the source search is empty.

The previous complete Gooo source feature is retained. Current semantic and
projection originals have their own success records, but cannot authorize this
new head. The new head needs its own full original CI family and exact proof.
The cancelled original is never rerun or manually cancelled. No runtime speed,
race completion, model quality, CPU/RSS improvement or release is claimed here.

[Index and original job coordinates](index.json) ·
[Exact one-flag source comparison](workflow-observation.json)
