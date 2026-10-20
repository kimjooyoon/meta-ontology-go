# 선택한 식이 어디서 왔는지 Gooo로 읽기

`inspect.gooo`는 실행 기록의 필드 위치를 읽어 선택된 대안, 필드 이름과 미해결
값을 구분합니다. 분류 규칙과 반환 형식은 이 Gooo 파일에 있습니다.

다음 명령은 저장된 기록의 첫 아홉 행을 전달합니다. 원래 집계 프로그램이나
선택 모델을 실행하지 않습니다. Go 1.27.3와 저장소 루트를 기준으로 합니다.

```sh
saved="$(mktemp)"
gzip -dc internal/bodyexecution/testdata/choice22133-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-kind field-input-locations \
  --origin-limit 9 examples/record-choice-origins/gooo.workspace.json
rm "$saved"
```

`result.runtime.traces[].deliveries[0].actual`에서 결과를 봅니다.

| status | 이 행의 의미 |
| --- | --- |
| `DECLARED_ALTERNATIVE` | 선택된 식의 원래 선언 위치 |
| `FIELD_TARGET` | 원래 반환 생성자의 필드 이름 위치 |
| `UNRESOLVED` | 완성된 식에 단일 템플릿 위치를 붙이지 못한 행 |

첫 페이지는 아홉 행이고 전체는 27행입니다. 각 필드 쓰기에 대상·템플릿 값·선택
대안 행이 붙으므로, 27행을 독립적인 계산 27개로 세지 않습니다. 다음 페이지는
`--origin-offset 9`로 읽을 수 있습니다. 이 입력 실행은 예상 결과를 제공하지
않으므로 `finite_total=0`입니다.

좌표는 UTF-8 바이트 범위입니다. 필드와 활동의 고유 ID는 표시 이름과 별도로
보존합니다. 저장 기록의 일관성 검사는 원래 생산자의 실행을 다시 입증하는
절차가 아니며, 일부만 남은 기록은 그 범위를 유지합니다.

[조립 예제](../assembled-length-summary/README.md) ·
[언어 설명](../../docs/language/return-choice-origins.ko.md) ·
[실행·실패 기록](../../docs/observations/return-choice-origins/README.md)
