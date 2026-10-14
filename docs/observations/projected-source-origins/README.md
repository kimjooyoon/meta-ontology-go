# Declaration origins after source projection

The original PR1568 projection job115941601067 failed in the source-rendering
check: the original role-verification file had zero declarations after extraction.
Its own uploaded patch shows the four required functions in three new physical
files. Artifact11907244532 was downloaded once and matched its API SHA256 digest.
This is a source-owner counterexample. The nine semantic roles and physical
write-set assertions remain intact.

The revised check finds each required free-function declaration in its actual
package files. Missing, repeated or foreign owners fail. All four declarations
retain their rendered75-line constraints. Actual extraction is checked once for
each observed owner, and emitted helper files keep the same75-line limit. The
original filename and an assumed helper suffix are not identity inputs.

The same head's semantic job115941598791 completed extraction1038/1038, then
exited without printing the projected compilation result. The compilation
function restored shell errexit before returning its failing status, so the
caller could not finish its result handling. The revised function captures the
command through a conditional, leaves the caller's flags intact and exposes the
captured output and digest-bound observation on failure. The command and
nonzero acceptance behavior are unchanged.

## New observations

- [owners39503](owners39503/ended.json): a whole-file layout and an empty-original
  layout with four moved declarations both preserve all four owners and actual
  extraction. Duplicate and foreign/missing owners are rejected. These are new
  source fixtures. The completed36013 or34731 cohort is not rerun locally.
- [compile40117](compile40117/ended.json): two new synthetic command inputs return
  status0 and37. The caller receives both statuses; stdout/stderr byte counts and
  SHA256 identities agree. A fake command is used. No real module test, compiler
  construction or model study is repeated by this observation.
- [static40331](static40331/ended.json): final75 Go files produce no fix diff or
  stderr, with zero tests executed.

[Index and byte identities](index.json) retain each exact source archive and
the original compressed failure logs. The full submitted source must pass its
own original workflow family, exact proof and current six jobs before acceptance.
These observations establish their stated scopes; native behavior, release
publication and whole-language completeness require separate records.
