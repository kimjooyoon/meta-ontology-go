# Go 도구 버전 변경의 영향 범위

한 버전을 올릴 때 현재 설정과 과거 기록을 함께 치환하면 원본 기록이 훼손됩니다.
여기서는 먼저 **현재 변경 대상의 발견률**을 Gooo 소스로 계산합니다. 파일 조회는
Go 어댑터가 수행하고, 분류가 충분한지에 대한 판단은 [impact.gooo](impact.gooo)의
`AssessImpact`가 수행합니다. 모델 호출이나 저장소 수정은 없습니다.

저장소 루트에서 실행합니다.

```sh
go run ./scripts/toolchain-version-impact > /tmp/gooo-version-impact.json
go run ./scripts/toolchain-version-impact --gooo-inputs > /tmp/gooo-version-inputs.json
go run ./cmd/gooo package execute --inputs /tmp/gooo-version-inputs.json \
  examples/toolchain-version-impact/gooo.workspace.json
go run ./cmd/gooo package execute --json --cases \
  examples/toolchain-version-impact/cases.json \
  examples/toolchain-version-impact/gooo.workspace.json
go run ./scripts/toolchain-version-impact --plan-version 1.27.4 \
  > /tmp/gooo-version-plan.json
go run ./scripts/toolchain-version-impact --plan-version 1.27.4 --gooo-inputs \
  > /tmp/gooo-version-plan-inputs.json
go run ./cmd/gooo package execute --inputs /tmp/gooo-version-plan-inputs.json \
  examples/toolchain-version-impact/plan.workspace.json
go run ./cmd/gooo package execute --json --cases \
  examples/toolchain-version-impact/plan.cases.json \
  examples/toolchain-version-impact/plan.workspace.json
# 활성 범위의 미분류 소비 지점이 0일 때만 성공합니다.
go run ./scripts/toolchain-version-impact --plan-version 1.27.4 \
  --materialize-dir /tmp/gooo-version-candidate
```

`go.mod`의 `go` 지시문이 관찰 버전의 기준입니다. 어댑터는 Git에 등록된 파일만
읽고 파일 해시, 경로, 줄과 분류를 출력합니다. 현재 활성 범위는 `go.mod`,
`.github/workflows/*.yml|*.yaml`, 추적 중인 `.gooo` 소스입니다.
`go-version` 또는 `GOTOOLCHAIN`의 값으로
정확히 표기된 한 건, 모듈 지시문과 정확한 Go 버전 검사식을 **분류됨**으로 봅니다.
Gooo의 `compatibility-test-contract`, `workflow-go-version`, `go-toolchain` 배지 행도
필드 구조와 버전 일치 여부를 확인한 뒤 분류합니다. 그 밖의 실행 명령·검사식·설명 문구는
**unresolved**로 남깁니다. Go 코드·테스트·문서의 현 버전 표기는 **outside scope**로
따로 셉니다. `docs/releases/`의 다른 Go 버전은
과거 기록으로 세되 현재 변경 대상으로 간주하지 않습니다.
현행 모듈과 다른 버전을 선택하는 워크플로 값과 Gooo 필드 값은 **stale selector**로
별도 표시합니다.
분류됨은 구문을 식별했다는 뜻이며, 생성이나 출처 관계가 입증됐다는 뜻은 아닙니다.

`--plan-version`은 분류된 표기의 정확한 바이트 구간, 변경 전후 값,
파일 원본 SHA-256, 관찰에 쓰인 구문 분류기 ID와 버전을 한 계획에 묶습니다. 나머지 표기는
`unresolved`에 남습니다. [plan.gooo](plan.gooo)의 `AssessProjection`은
후보 수와 관찰 수가 일치하고 미분류·중복 구간·과거 기록 수정이 없을 때만
`READY_FOR_DETACHED_PROJECTION`을 반환합니다. `--plan-version`만으로는 후보 파일을
만들거나 원본을 수정하지 않습니다. `--materialize-dir`은 Gooo 소스를 실행해
정확히 이 계획에 대한 `READY_FOR_DETACHED_PROJECTION`을 받았을 때만
이미 존재하는 부모 아래의 새 디렉터리에 변경 파일과 영수증을 만듭니다.
쓰기 직전에 전체 관찰과 각 파일 원본 해시를 다시 확인하고 원본 트리는 수정하지
않습니다. 실패하면 생성 중인 후보 디렉터리를 지웁니다. 현재 활성 범위의
미분류 항목은 0개이므로 동시 후보 생성이 가능합니다. 새 버전 문자열은 조사용
입력일 뿐 실제 릴리스 존재 여부를 검증하지 않습니다.
`observation_sha256`은 직렬화된 관찰 기록의 해시입니다. `plan_core_sha256`은
그 값을 포함하되 자기 자신의 해시 필드는 빈 문자열이고 `provenance`가 없는
계획 본문의 해시입니다. Gooo 판단 입출력이 이 값을 그대로 전달합니다.
구문 분류기 ID는 Go 어댑터의 관찰 방법을 식별하며, 아직 자동 수정 규칙의
의미 ID가 아닙니다. 계획의 JSON-LD에는 실제 계획 생성 활동이 관찰 기록과
해시로 식별한 `go.mod`를 `prov:used`했고, 계획을 `prov:wasGeneratedBy` 했다는
관계를 둡니다. 분리된 후보가 실제로
만들어진 경우에만 그 영수증은 각 파일 입력·출력에 다른 ID를 부여하고,
후보 생성 활동의 `prov:used`에 입력 파일·계획·Gooo 결정을 기록합니다.
파일을 실제로 고치지 않았으면 후보 파일의 생성 관계를 주장하지 않습니다.

초기 관찰(#1737)에서는 활성 표기 184건 중 분류 172건, 미분류 12건이었습니다.
직전 체크아웃의 루트 `go.mod`를 참조하도록 `setup-go` 106곳을 바꾼 뒤에는
활성 표기 78건 중 분류 66건, 미분류 12건이었습니다. 현재는 실행 기대값·영수증·
테스트용 모듈도 `go.mod`에서 읽도록 바꾸어 활성 표기 62건 중 분류 62건,
미분류 0건이었습니다. Gooo 소스의 현행 도구체인 선언을 편입한 뒤에는 활성 표기
66건 중 분류 66건, 미분류 0건입니다. 활성 범위 분류율은 10,000/10,000입니다.
이는 구문 발견률이며 자동 수정의 전역 완전성을 뜻하지 않습니다. 과거 Go 버전
36건과 범위 밖의 현 버전 표기는 분모에 들어가지 않습니다. 해당 수는 테스트·문서
변경에도 움직이므로 보고서의 `outside_scope_occurrences`로 확인합니다. 숫자는 이 커밋의
분류기와 저장소 상태에 대한 관찰이며, 모든 파일이 자동 수정 가능하다는 뜻이
아닙니다. 네 개의 독립 입력 사례가
완전·미분류·버전 불일치·모순 수치를 판정합니다.

## 다음 변경 단위

1. 각 현행 소비 지점에 `source version ID`, `projection rule ID/version`, 대상 경로와
   구문 위치, 원본 해시, 검증 의무를 저장합니다. 이름 변경과 위치 이동은 의미 ID로
   식별하며, 분류되지 않은 새 소비 지점은 계획의 미해결 항목으로 남깁니다.
2. 한 후보 버전에서 영향 범위를 닫고, 모든 대상의 원본 해시를 다시 확인한 뒤
   임시 트리에 함께 투영합니다. 결과 차이와 검사 결과를 확인한 후 하나의 변경
   단위로 제출합니다. 과거 릴리스 기록과 이미 발행된 영수증은 새 revision으로만
   이어집니다.
3. 실제 조회 활동이 읽은 파일에는 `prov:used`, 생성된 계획에는
   `prov:wasGeneratedBy`, 새 투영 결과에는 확인된 경우에만 `prov:wasDerivedFrom`을
   붙입니다. 이 어휘 자체는 “같이 수정해야 한다”는 규칙이 아닙니다. 변경 의무는
   별도의 결정론적 투영 계약입니다. 어댑터의 계획과 실제 후보 영수증은 각각
   관찰된 읽기와 실제 쓰기에 대한 PROV-O 관계만 기록합니다.

현재 `1.27.4`라는 가상 다음 버전의 분리 후보를 생성하면, Gooo가 승인한 계획에서
30개 파일의 66개 정확한 구간이 한 번에 투영됩니다. 두 Gooo 소스도 같은 후보에
포함되며, 후보의 Gooo 구문 검사는 통과했습니다. 이 수는 실제 버전 출시나
업그레이드 성공을 의미하지 않습니다. CI·스크립트·외부 설정까지 닫힌 영향 범위를
증명한 것도 아닙니다. 뒤이어 후보 빌드·검사와 범위 밖 영향 확인이 필요합니다.
두 Gooo 정책의 Go 상수는 생성된 평가기의 값에서 읽도록 바뀌었으므로 실제 전환 시
후보 소스에서 평가기를 다시 생성해야 합니다. 공개 배지의 README 표현도 같은
소스에서 다시 투영하고 검증해야 합니다.

미래 비용을 미리 지불할 가치가 있는지는 다음 전환부터 실측합니다. 한 전환에 대해
`발견된 대상/실제 영향 대상`, `미분류 건수`, `원본 해시 충돌`, `후보 생성 시간`,
`검증 시간`, `수동 편집 시간`, `사후 누락 건수`를 같은 범위와 함께 기록합니다.
초기 계약 구축 시간과 각 전환에서 절약한 수동 시간의 누계를 비교해야 회수 시점을
말할 수 있습니다. PROV-O 어휘는 [W3C 권고안](https://www.w3.org/TR/prov-o/)을 따릅니다.
