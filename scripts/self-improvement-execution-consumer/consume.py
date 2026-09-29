#!/usr/bin/env python3
"""Consume one verified grant through the existing Gooo execution CLI."""

import argparse
import hashlib
import json
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path


SCHEMA = "gooo/self-improvement-execution-consumption/v1"
SOURCE_PATH = "examples/language-value-witness/main.gooo"
ACTIVITY = "Increment"
EXPECTED_CORPUS = [
    ("negative", -2, -1),
    ("negative-to-zero", -1, 0),
    ("zero", 0, 1),
    ("positive", 41, 42),
    ("maximum-boundary", 9223372036854775806, 9223372036854775807),
]


def canonical_digest(value):
    payload = json.dumps(value, sort_keys=True, separators=(",", ":")).encode()
    return "sha256:" + hashlib.sha256(payload).hexdigest()


def load(path):
    return json.loads(Path(path).read_text(encoding="utf-8"))


def write(path, value):
    Path(path).write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")


def require(condition, message):
    if not condition:
        raise ValueError(message)


def project_v25(contract):
    return {
        "schema": contract["schema"],
        "contract_id": contract["contract_id"],
        "contract_digest": contract["digest"],
        "decision": contract["decision"],
        "resolution": contract["resolution"],
        "candidate_stable_id": contract["candidate_stable_id"],
        "candidate_digest": contract["candidate_digest"],
        "subject_sha": contract["subject_sha"],
        "observation_digest": contract["observation_digest"],
        "candidate_input_digest": contract["candidate_input_digest"],
        "operation_id": contract["operation_id"],
        "bounded_target": contract["bounded_target"],
        "evaluator_registry_digest": contract["evaluator_registry_digest"],
        "toolchain_test_contract_identity": contract["toolchain_test_contract_identity"],
        "max_executions": contract["max_executions"],
        "repository_writes_allowed": contract["repository_writes_allowed"],
        "execution_authorized": contract["execution_authorized"],
        "execution_grant_required": contract["execution_grant_required"],
        "valid": True,
    }


def validate_pair(grant, contract, metadata):
    resolution = grant["resolution"]
    source = grant["request"]["source_artifact"]
    contract_resolution = contract
    execution_input = contract_resolution["execution_input"]
    snapshot = execution_input["source"]
    corpus = execution_input["corpus"]
    receipt = resolution.get("receipt") or {}
    require(grant.get("schema") == "gooo/self-improvement-execution-grant/v2", "invalid v26 schema")
    require(grant.get("verification", {}).get("verified") is True, "v26 report is not verified")
    require(resolution.get("decision") == "CLOSED" and resolution.get("resolution") == "GRANTED_UNCONSUMED", "grant is not open")
    require(resolution.get("grant_allows_execution") is True and resolution.get("one_use_enforced") is False, "grant boundary is not exact")
    require(receipt.get("grant_allows_execution") is True and receipt.get("remaining_uses") == 1, "grant receipt is not one use")
    require(receipt.get("consumed_uses") == 0 and resolution.get("execution_count") == 0, "grant was already consumed")
    require(resolution.get("repository_writes") == 0 and resolution.get("local_test_executions") == 0, "grant includes effects")
    require(grant.get("decision_source") == "system-derived-exact-evidence", "grant decision is not system-derived")
    require(contract_resolution.get("schema") == "gooo/self-improvement-execution-contract/v1", "invalid v25 schema")
    require(contract.get("verification", {}).get("verified") is True, "v25 report is not verified")
    require(contract_resolution.get("decision") == "CLOSED" and contract_resolution.get("execution_grant_required") is True, "v25 is not grant-bound")
    require(source.get("repository") == metadata["repository"], "grant repository does not match workflow")
    require(source.get("workflow_run_id") == metadata["contract_run_id"], "v25 run identity mismatch")
    require(source.get("workflow_run_attempt") == metadata["contract_run_attempt"] == 1, "v25 attempt mismatch")
    require(source.get("artifact_id") == metadata["contract_artifact_id"], "v25 artifact identity mismatch")
    require(source.get("artifact_digest") == metadata["contract_artifact_digest"], "v25 archive digest mismatch")
    require(source.get("observed_artifact_digest") == source.get("artifact_digest"), "v25 observed digest mismatch")
    require(source.get("artifact_retrieved") is True and source.get("artifact_expired") is False, "v25 artifact is unavailable")
    require(source.get("artifact_expiry_known") is True and source.get("artifact_retrieval_error", "") == "", "v25 artifact retrieval is incomplete")
    require(grant["request"]["v25_pre_execution_contract"] == project_v25(contract_resolution), "v25 projection does not match v26 grant")
    require(execution_input.get("subject_sha") == metadata["subject_sha"] == contract_resolution.get("subject_sha"), "execution subject mismatch")
    require(snapshot.get("path") == SOURCE_PATH, "execution source path is not registered")
    source_bytes = snapshot.get("bytes", "").encode()
    require(hashlib.sha256(source_bytes).hexdigest() == snapshot.get("digest", "").removeprefix("sha256:"), "source snapshot digest mismatch")
    require(Path(SOURCE_PATH).read_bytes() == source_bytes, "authorized source differs from exact subject checkout")
    require(execution_input.get("max_executions") == 1 and execution_input.get("repository_writes_allowed") is False, "execution input exceeds the grant")
    require(execution_input.get("allowed_effects") == [], "execution input declares external effects")
    observed = [(case["id"], case["input"], case["expected_output"]) for case in corpus]
    require(observed == EXPECTED_CORPUS, "execution corpus is not the five-case registered corpus")
    require(metadata.get("grant_run_attempt") == 1 and metadata.get("contract_run_attempt") == 1, "workflow rerun is not consumable")
    return execution_input, receipt


def reservation_for(grant, metadata):
    reservation = {
        "schema": SCHEMA + "/reservation",
        "grant_id": grant["resolution"]["receipt"]["grant_id"],
        "request_digest": grant["request"]["digest"],
        "subject_sha": metadata["subject_sha"],
        "source_run_id": metadata["grant_run_id"],
        "source_run_attempt": metadata["grant_run_attempt"],
        "status": "CONSUMPTION_RESERVED",
        "execution_count": 0,
        "consumed_uses": 1,
        "remaining_uses": 0,
        "one_use_enforced": True,
        "reserved_at": datetime.now(timezone.utc).isoformat(),
    }
    reservation["digest"] = canonical_digest(reservation)
    return reservation


def validate_reservation(reservation, grant, metadata):
    expected = reservation_for_fields(reservation)
    require(reservation.get("digest") == canonical_digest(expected), "reservation digest mismatch")
    require(reservation.get("schema") == SCHEMA + "/reservation", "reservation schema mismatch")
    require(reservation.get("status") == "CONSUMPTION_RESERVED" and reservation.get("one_use_enforced") is True, "reservation is not active")
    require(reservation.get("consumed_uses") == 1 and reservation.get("remaining_uses") == 0, "reservation use count mismatch")
    require(reservation.get("grant_id") == grant["resolution"]["receipt"]["grant_id"], "reservation grant mismatch")
    require(reservation.get("request_digest") == grant["request"]["digest"], "reservation request mismatch")
    require(reservation.get("subject_sha") == metadata["subject_sha"], "reservation subject mismatch")
    require(reservation.get("source_run_id") == metadata["grant_run_id"], "reservation run mismatch")
    require(reservation.get("source_run_attempt") == metadata["grant_run_attempt"], "reservation attempt mismatch")


def reservation_for_fields(reservation):
    return {key: value for key, value in reservation.items() if key != "digest"}


def consume(grant, contract, metadata, reserved, output_path):
    started = time.perf_counter_ns()
    execution_input, receipt = validate_pair(grant, contract, metadata)
    validate_reservation(reserved, grant, metadata)
    with tempfile.TemporaryDirectory(prefix="gooo-bounded-consumer-") as temporary:
        executable = Path(temporary) / "gooo"
        build_started = time.perf_counter_ns()
        build = subprocess.run(["go", "build", "-o", str(executable), "./cmd/gooo"], capture_output=True, text=True)
        build_elapsed = time.perf_counter_ns() - build_started
        cases = []
        if build.returncode != 0:
            build_error = build.stderr.strip() or "go build failed"
            cases = [{"id": case["id"], "input": case["input"], "expected_output": case["expected_output"],
                      "passed": False, "elapsed_nanoseconds": 0, "error": build_error}
                     for case in execution_input["corpus"]]
        else:
            build_error = ""
            for index, case in enumerate(execution_input["corpus"]):
                input_path = Path(temporary) / f"input-{index}.json"
                input_path.write_text(json.dumps({"value": case["input"]}) + "\n", encoding="utf-8")
                case_started = time.perf_counter_ns()
                run = subprocess.run([str(executable), "run", "--json", "--entry", ACTIVITY,
                                      "--input", str(input_path), SOURCE_PATH], capture_output=True, text=True)
                elapsed = time.perf_counter_ns() - case_started
                item = {"id": case["id"], "input": case["input"], "expected_output": case["expected_output"],
                        "passed": False, "elapsed_nanoseconds": elapsed}
                if run.returncode != 0:
                    item["error"] = run.stderr.strip() or f"Gooo execution exited {run.returncode}"
                else:
                    try:
                        result = json.loads(run.stdout)
                        actual = result["execution"]["results"][ACTIVITY]
                        item.update({"actual_output": actual["value"],
                                     "execution_digest": result["execution"]["execution_digest"],
                                     "result_digest": actual["result_digest"]})
                        item["passed"] = (
                            result.get("schema") == "gooo/value-execution-plan/v1" and
                            result.get("decision") == "PASS" and result.get("entry") == ACTIVITY and
                            result.get("source_digest") == execution_input["source"]["digest"] and
                            result.get("semantic_fingerprint") == execution_input["activity"]["semantic_fingerprint"] and
                            result["execution"].get("apply_calls") == 1 and
                            result["execution"].get("deliveries") == 0 and
                            actual.get("value") == case["expected_output"] and
                            actual.get("producer_activity") == ACTIVITY and
                            actual.get("result_digest", "").startswith("sha256:")
                        )
                        if not item["passed"]:
                            item["error"] = "Gooo result or execution evidence did not match the authorized case"
                    except (KeyError, TypeError, ValueError) as error:
                        item["error"] = f"invalid Gooo execution receipt: {error}"
                cases.append(item)
    passed = sum(1 for case in cases if case["passed"])
    report = {
        "schema": SCHEMA,
        "grant_id": receipt["grant_id"],
        "grant_request_digest": grant["request"]["digest"],
        "grant_receipt_digest": receipt["digest"],
        "reservation_digest": reserved["digest"],
        "subject_sha": metadata["subject_sha"],
        "grant_run_id": metadata["grant_run_id"],
        "grant_run_attempt": metadata["grant_run_attempt"],
        "grant_artifact_id": metadata["grant_artifact_id"],
        "grant_artifact_digest": metadata["grant_artifact_digest"],
        "contract_run_id": metadata["contract_run_id"],
        "contract_run_attempt": metadata["contract_run_attempt"],
        "contract_artifact_id": metadata["contract_artifact_id"],
        "contract_artifact_digest": metadata["contract_artifact_digest"],
        "candidate_stable_id": execution_input["candidate_stable_id"],
        "candidate_digest": execution_input["candidate_digest"],
        "execution_input_digest": execution_input["digest"],
        "observation_digest": execution_input["observation_digest"],
        "operation_id": execution_input["operation_id"],
        "execution_count": 1 if cases and not build_error else 0,
        "evaluator_invocations": len(cases) if not build_error else 0,
        "consumed_uses": 1,
        "remaining_uses": 0,
        "one_use_enforced": True,
        "repository_writes": 0,
        "external_effects": [],
        "plan_source_digest": execution_input["source"]["digest"],
        "plan_identity_digest": canonical_digest({
            "source_digest": execution_input["source"]["digest"],
            "semantic_fingerprint": execution_input["activity"]["semantic_fingerprint"],
        }),
        "executor_build_elapsed_nanoseconds": build_elapsed,
        "evaluation_elapsed_nanoseconds": sum(case["elapsed_nanoseconds"] for case in cases),
        "elapsed_nanoseconds": time.perf_counter_ns() - started,
        "case_count": len(execution_input["corpus"]),
        "passed_cases": passed,
        "cases": cases,
        "performance_improvement": "UNKNOWN",
    }
    if build_error:
        report["executor_build_error"] = build_error
    report["digest"] = canonical_digest(report)
    write(output_path, report)
    require(passed == report["case_count"], "bounded execution produced failing or incomplete value cases")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=("reserve", "consume"))
    parser.add_argument("--grant", required=True)
    parser.add_argument("--contract", required=True)
    parser.add_argument("--metadata", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--reservation")
    args = parser.parse_args()
    try:
        grant = load(args.grant)
        contract = load(args.contract)
        metadata = load(args.metadata)
        validate_pair(grant, contract, metadata)
        if args.mode == "reserve":
            write(args.output, reservation_for(grant, metadata))
            return 0
        require(args.reservation is not None, "--reservation is required for consume mode")
        consume(grant, contract, metadata, load(args.reservation), args.output)
        return 0
    except (OSError, ValueError, KeyError, TypeError) as error:
        print(str(error), file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
