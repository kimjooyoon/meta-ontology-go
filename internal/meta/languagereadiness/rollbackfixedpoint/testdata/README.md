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

## Known refusals retained in a collected guard

Three timestamp-free gzip fixtures preserve original JSON bytes from
[workflow 39049027405](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/39049027405),
attempt 1, whose actual subject is `23b83809300c153247e96bbf357fa8fc4635bc8a`.
The workflow metadata head is `6d37a33b0af4a4541df44f9e214bc955dba6670e`;
the receipt subject and its original guard/ledger links define these inputs.

| Fixture | Artifact | Raw JSON SHA256 |
| --- | --- | --- |
| `guard-original-23b83809.json.gz` |12106917319|`60004415083059925c2464aa78ddd4ec6f10c5097f22a95b2ed9ade6de258565`|
| `transformation-original-23b83809.json.gz` |12106672176|`5071ae242cdb0bf596f0b36035796d138de96ce3b8b9dadb992457976ae86a73`|
| `rollback-lossy-original-23b83809.json.gz` |12107097282|`885dae59d83de26bc471bcf9d5821367decedcab50727ba6c165898e0c590f2c`|

The downloaded ZIP digests were respectively
`f368eb311d15d1da91c1dadbd9498c135a7955b7311b57860f9c5f1a44a7a7af`,
`5608c88cc682ecc1c7ca456860a75df0dd94636b6ac0e5afd311ec2bfcf8d3b4`, and
`3ebba83d7c20851acd6fcdc7acda68bb5e04e3252867d4ab8e53511327d45dd7`.
The 71,493 original bytes occupy 21,257 gzip bytes. This is fixture storage only.

The full guard has 9 satisfied, 1 not-satisfied and 2 unresolved coordinates.
Its known refusal is the default-branch boundary. The historical rollback
projection omitted the known-refusal count, and its bytes stay unchanged.
New collection retains the optional positive `guard.not_satisfied` field.
Zero stays omitted, preserving earlier reports. An older reader that drops a
positive field cannot validate its new digest by silently discarding the field.

The new projection still reports `FAIL_CLOSED`, 8/10 coordinates, two rejected
recovery coordinates, two completed sandbox effects and zero recoveries,
promotions or repository writes. Contradictory populations and mutation
permissions remain refused. Tests decode original data and exercise the current
collector/observer; no historical workflow, transformation or model is rerun.
