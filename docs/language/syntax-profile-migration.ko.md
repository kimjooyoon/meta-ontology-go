# 같은 사례를 유지하며 문법 프로필 바꾸기

Gooo 원문을 확장하다 보면 같은 프로그램에 더 넓은 필드 타입이 필요할 수 있습니다.
문자열 필드만 쓰던 프로그램에 불리언 필드를 넣는 경우가 그 예입니다.
이때 새 사례를 추가해 평가 대상 수를 늘리는 대신, 기존 사례와 명시적인 해석
프로필을 연결하는 변경 후보를 만듭니다.

보기 힌트는 이 연산의 입력이 아닙니다. 보기 오류에는 기본 보기를 사용하지만,
문법 프로필 오류는 원문을 어떻게 해석할지 정할 수 없다는 진단으로 남습니다.
프로필 변경은 별도의 명령이며 컴파일·생성·실행 때 자동으로 수행하지 않습니다.

## 에이전트가 사용하는 순서

하나의 실행 파일을 만들어 관찰·생성·검증에 사용합니다. 실행 파일 자체의 해시도
입력 정체성에 포함되므로, 다른 소스로 다시 빌드하면 요청을 새로 관찰해야 합니다.
현재 저장소는 Go 1.27.2를 사용합니다.

```sh
go build -o /tmp/gooo-syntax-profile-migration ./cmd/syntax-profile-migration
```

호스트는 `request.json`에 다음 정보를 명시합니다.

- `schema`: `gooo/syntax-profile-migration-request/v1`.
- `case`: 현재 등록 JSON의 대상 `CaseDefinition` 전체. 기존 ID·경로·플래그를 포함합니다.
- `target_entity_fields_profile`: 정확한 커널 프로필 ID. 예: `gooo.entityfields.go-projection.v2`.
- `base_version`: 현재 분모 이력의 마지막 버전.
- 입력 관찰 전에는 `snapshot_digest`, `source_digest`, `replacement_digest`, `toolchain`,
  `execution_identity`를 빈 값으로 둘 수 있습니다. 생성에는 실제 관찰값이 필요합니다.

교체할 Gooo 원문은 최대 128KiB의 별도 파일로 전달합니다.

```sh
/tmp/gooo-syntax-profile-migration --root . --request request.json \
  --source-after after.gooo --inspect
```

출력의 `request` 객체에는 호출자가 준 사례·대상 프로필과 실제 관찰한 고정값이
함께 있습니다. 에이전트는 이 객체를 다음 요청 JSON으로 저장하면 됩니다.
사람이 값을 승인하거나 해석해서 재조립하는 단계는 없습니다. `--inspect`는 바이트와
도구 정체성의 관찰이며, 해당 소스를 수용하거나 적용했다는 결과는 아닙니다.

```sh
/tmp/gooo-syntax-profile-migration --root . --request pinned-request.json \
  --source-after after.gooo --output /tmp/gooo-profile-proposal

/tmp/gooo-syntax-profile-migration --root . --request pinned-request.json \
  --source-after after.gooo --verify-candidate /tmp/gooo-profile-proposal/candidate.json
```

출력 디렉터리는 저장소 밖의 새 디렉터리여야 합니다. 기존 디렉터리를 덮어쓰지
않습니다. 출력 실패 중 일부 파일이 남을 수 있으므로 `PROPOSAL_EXPORTED` 응답과
후보 검증이 모두 없는 디렉터리를 완성된 후보로 취급하지 않습니다.

## 후보의 내용과 보존 범위

| 산출물 | 변경하는 부분 | 보존하는 부분 |
| --- | --- | --- |
| 등록 JSON | 대상 사례의 명시적 프로필과 이전 플래그 | 사례 ID·경로·순서·개수, 다른 JSON 바이트 |
| 실제 네이티브 등록 소스 | 해당 사례 리터럴의 프로필과 이전 플래그 | 다른 사례, 리터럴 밖의 원래 바이트 |
| 프로그램 원문 | 호출자가 지정한 교체 바이트 | `before/`에 변경 전 바이트, 원문별 해시 |

파일 위치는 실제 `Cases` 표현식을 따라 찾습니다. 파일명이나 행 번호를 등록의
근거로 삼지 않습니다. 등록에 쓰이지 않은 부수적인 리터럴은 수정하지 않습니다.
상수의 해석이 모호하거나 두 등록의 사례 집합이 다르면 후보를 만들지 않습니다.

`candidate.json`과 `before/`, `members/`를 함께 내보냅니다. 후보 JSON에도 변경 전후
바이트가 있어 원본과 각각의 해시를 확인할 수 있습니다. 이는 원문 보관이며 압축이나
디스크 절감 기능으로 설명하지 않습니다. 큰 이력의 비손실 보관은
[출처 저장 경로](advisory-origin-views.ko.md)의 별도 계약입니다.

변경 전후 소스는 선택한 프로필에서 각각 왕복 법칙을 검사합니다. 후보 검증은
현재 입력과 요청·실행 정체성을 다시 확인하고 예상 산출물 전체와 비교합니다.
기존 분모 이력은 산출물에 포함하지 않습니다. `case_population`은 등록된 사례 수이며
실행한 프로그램 수나 언어의 완전성 백분율이 아닙니다.

## 결과를 해석하기

- `PROPOSAL_EXPORTED`: 외부 디렉터리에 세 역할의 후보를 작성했습니다.
- `EXACT_SOURCE_CANDIDATE`: 현재 고정 입력에서 해당 후보 전체가 예상 결과와 같습니다.
- `semantic_admission=UNASSESSED`, `state=PROPOSAL_ONLY`: 의미 수용과 실제 적용은 별도입니다.
- `repository_writes=0`, `apply_authorized=false`, `promotion_authorized=false`를 유지합니다.

이 명령에는 저장소 적용 기능이 없습니다. 호스트가 후보를 변경으로 제출하면 그
정확한 소스의 원래 CI가 기존 계약과 함께 판단합니다. 일반 파서의 기본 지원 범위를
넓히거나, 새 사례를 등록하거나, 잘못된 프로그램을 보조 힌트로 낮추는 기능도 없습니다.

다음 실제 사용 과제는 미등록 연산의 여섯 필드 결과를 Gooo 레코드가 직접 정하고,
네이티브 연결기가 안정 필드 ID로 전송하게 하는 것입니다. 현재 이 명령의 후보
검증은 그 정책의 실제 실행·독립 소비를 대신하지 않습니다.
