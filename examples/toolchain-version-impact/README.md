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
```

`go.mod`의 `go` 지시문이 관찰 버전의 기준입니다. 어댑터는 Git에 등록된 파일만
읽고 파일 해시, 경로, 줄과 분류를 출력합니다. 현재 첫 범위는 `go.mod`와
`.github/workflows/*.yml|*.yaml`입니다. `go-version` 또는 `GOTOOLCHAIN`의 값으로
정확히 표기된 한 건, 모듈 지시문과 정확한 Go 버전 검사식을 **분류됨**으로 봅니다.
그 밖의 실행 명령·검사식·설명 문구는 **unresolved**로 남깁니다. 예제·스크립트·문서의
현 버전 표기는 **outside scope**로 따로 셉니다. `docs/releases/`의 다른 Go 버전은
과거 기록으로 세되 현재 변경 대상으로 간주하지 않습니다.
현행 모듈과 다른 버전을 선택하는 워크플로 값은 **stale selector**로 별도 표시합니다.
분류됨은 구문을 식별했다는 뜻이며, 생성이나 출처 관계가 입증됐다는 뜻은 아닙니다.

2026-10-21의 Go 1.27.3 기준 관찰은 활성 표기 184건 중 분류 172건,
미분류 12건입니다. Gooo는 이 경우 `ACTIVE_SCOPE_UNRESOLVED`와 9,347/10,000을
반환합니다. 과거 Go 버전 36건과 범위 밖의 현 버전 302건은 이 분모에 들어가지
않습니다. 숫자는 이 커밋의 분류기와 저장소 상태에 대한 관찰이며, 모든 파일이
자동 수정 가능하다는 뜻이 아닙니다. 네 개의 독립 입력 사례가
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
   별도의 결정론적 투영 계약입니다. 현재 어댑터는 해시와 구문 관찰만 기록하며
   RDF 관계나 쓰기 권한을 주장하지 않습니다.

이 순서의 첫 단계가 완성되면 새 버전 전환의 절약 시간을 측정할 수 있습니다.
`미분류 건수`, `누락된 활성 대상`, `원본 해시 충돌`, `후보 생성 시간`, `검증 시간`,
`수동 편집 시간`을 각각 기록해야 초기 계약 구축 비용의 회수 시점을 판단할 수
있습니다. PROV-O 어휘는 [W3C 권고안](https://www.w3.org/TR/prov-o/)을 따릅니다.
