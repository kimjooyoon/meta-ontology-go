# 고유 ID로 기원 조회하기

필드의 값이나 표시 이름이 같아도 작성한 개념은 다를 수 있습니다. 저장한 실행
기록에서 필드·활동의 고유 ID로 먼저 고르면 원하는 기원을 찾기 위해 행 번호를
외울 필요가 줄어듭니다. Go는 해당 기록을 전달하고, Gooo가 그 의미를 해석합니다.

```sh
saved="$(mktemp)"
gzip -dc internal/bodyexecution/testdata/choice22133-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" \
  --origin-field-id urn:assembledlength22133:total --origin-limit 3 \
  examples/record-choice-origins/gooo.workspace.json
rm "$saved"
```

ID 선택을 지정하고 종류를 생략하면 `field-input-locations`를 사용합니다.
`field-writes`, `field-uses`도 지정할 수 있습니다. 다른 종류와 함께 사용하면
지원하지 않는 조합이라는 진단을 냅니다. 필터를 조용히 무시하지 않습니다.

| 선택 | 찾는 주체 |
| --- | --- |
| `--origin-field-id` | 작성한 레코드 필드 |
| `--origin-activity-id` | 기록한 사건이 발생한 활동 |
| `--origin-root-activity-id` | 그 활동을 호출한 진입 활동 |

여러 ID를 주면 모두 일치하는 기록을 고릅니다. 도우미의 활동 ID와 진입 활동
ID는 다를 수 있습니다. 표시 이름을 추측하거나 대소문자를 바꾸지 않고 원래
ID 문자열로 비교합니다.

## 세 가지 행 수

선택을 지정하면 `origin_input`에 다음 분모가 있습니다.

- `source_rows`: 이 조회 종류의 원래 남아 있는 전체 행 수
- `total_rows`: 지정한 주체와 일치하는 행 수
- `returned_rows`: 현재 페이지로 전달한 행 수

예제 기록은 전체 27행 중 `total` 필드의 9행을 고르고 그중 3행을 전달합니다.
한 쓰기에 대상·값·선택 대안의 세 위치가 붙습니다. 9행은 독립적인 계산 9개의
정답률을 뜻하지 않습니다. 다음 페이지의 `offset`은 선택한 9행을 기준으로 하며,
같은 ID 조건을 유지해서 사용합니다.

## 원래 연결 번호는 유지

반환식을 평가하기 전에 쓰기를 시작하고, 그 안에서 도우미가 자기 필드를 쓸 수
있습니다. 그래서 기록 시작 순서와 완료 순서가 다릅니다. 선택 결과의 행 순서는
페이지를 위한 순서이며, 원래 사건·쓰기·부모 번호를 덮어쓰지 않습니다.

`field-lineage-columns`와 `field-local-columns`는 선택한 읽기의 사건 번호와
완전한 원래 케이스 저널을 함께 전달합니다. 다른 필드가 그 기원의 부모일 수
있으므로 주체 선택이 저널의 다른 필드까지 잘라내지는 않습니다.

ID를 지정하지 않은 예전 조회는 기존 직렬화와 분모를 유지합니다. 일치한 기록이
없는 일반 행 입력은 빈 결과 진단을 내며, 빈 표를 표현하는 내부 페이지 API는
전체 수·선택 수를 구별해서 반환합니다. 기록 누락, 깊이와 페이지 한계는 계속
별도로 살펴야 합니다. 원래 생산 프로그램과 모델은 이 조회로 실행하지 않습니다.

[기원 조회 시작 프로젝트](project-starters.md#start-an-origin-reader) ·
[Gooo로 선택의 기원 해석하기](../../examples/record-choice-origins/README.md)

[TypeQL·CUE·Soufflé에서 참고한 원리와 후속 범위](../research/origin-query-language-comparison.ko.md)
