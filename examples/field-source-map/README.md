# 필드의 원래 Gooo 위치 읽기

필드의 실행 기록에서 원본 `.gooo` 파일로 돌아가는 작은 메타 프로그램입니다.
`inspect.gooo`가 전달받은 위치를 필드 대상, 필드 값, 조립 템플릿, 선택 후보로
구분합니다. 분류 규칙은 이 Gooo 파일에서 바꿉니다.

저장된 두 케이스의 기록을 읽습니다. 원래 생산 프로그램과 모델은 실행하지 않습니다.
저장소 루트에서 Go 1.27.3를 사용합니다.

```sh
saved="$(mktemp)"
gzip -dc internal/bodyexecution/testdata/field-files16301-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-kind field-input-locations \
  --origin-offset 6 --origin-limit 6 \
  examples/field-source-map/gooo.workspace.json
rm "$saved"
```

`result.runtime.traces[].deliveries[0].actual`에서 원래 파일 이름, UTF-8 바이트
범위, 식, 필드의 고유 ID와 쓰기 번호를 볼 수 있습니다. `ORIGINAL_FILE`은
그 행에 원본 위치가 있다는 뜻입니다. `UNRESOLVED` 행은 이유를 함께 보존합니다.
이 입력 실행에는 예상값을 붙이지 않으므로 `finite_total=0`입니다.

조립된 값의 예도 같은 소비자로 읽을 수 있습니다.

```sh
saved="$(mktemp)"
gzip -dc internal/bodyexecution/testdata/field-fill16541-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-kind field-input-locations \
  --origin-limit 32 examples/field-source-map/gooo.workspace.json
rm "$saved"
```

`input+17`이 템플릿 `input+`와 후보 `17`에서 만들어졌다면, 두 조각의 원래
위치를 따로 전달합니다. 조립된 식 전체의 단일 원본 위치는 미해결로 남깁니다.

현재 좌표를 붙이는 구문은 명시적 필드 대입, 생성자의 필드 값, 필드 선택 읽기입니다.
파라미터 전달, 레코드 전체 복사·읽기 등의 필드 본문 좌표는 후속 범위입니다.
오래된 기록에 좌표가 없으면 `ORIGINAL_FILE_ANNOTATION_ABSENT`로 표시합니다.

[언어 설명](../../docs/language/field-source-locations.ko.md) ·
[실제 관찰과 실패 기록](../../docs/observations/field-source-locations/README.md)
