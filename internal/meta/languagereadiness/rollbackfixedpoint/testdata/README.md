# Conformant sandbox terminal regression

`applied-original-8f45c78e.json.gz` contains the unchanged `rollback-fixed-point-a.json`
from original [workflow 38137096323](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38137096323),
at source `8f45c78e1797514857f7380bc8a05790c7b2aa45`, attempt 1.
Artifact 11701920913 was downloaded once and its ZIP SHA256 checked before extraction:
`73aa84d0cb58bf48eb2e997fb561efb692451708dd6c9953aee3ad6a073bb70c`.

Uncompressed report SHA256:
`7b3a9dd5b012016b34157b51c4ebef0618d470ba2c7b0dd76a8e09cc71e5765a`.
The gzip wrapper uses no timestamp. Tests preserve the original report bytes and
retain its `FAIL_CLOSED` recovery decision, 8/10 satisfied coordinates, two rejected
recovery coordinates, two applied sandbox effects, and zero recoveries or promotions.

The original workflow failed to classify this observed outcome. Recognition as a
non-promoting terminal permits its result to be recorded and consumed. The original
recovery requirements, source boundaries and authorization counts remain unchanged.
The downstream guard-envelope projection used by the CLI regression is synthetic;
its role is to exercise the existing input reader alongside this full original report.
No original transformation, model study or workflow is rerun by these tests.
