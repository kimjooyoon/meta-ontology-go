# Gooo가 선언한 세 가지 조립 선택

`summary.gooo`는 라벨 수를 기준값에 더하고, 처음 여덟 라벨의 UTF-8 바이트 수와
전체 처리 여부를 반환합니다. 반복문과 조건, 출력 필드, 허용하는 대안은 Gooo가
선언합니다. 현재 예제는 모델 없이 여덟 조합을 정해진 순서로 살펴봅니다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/assembled-length-summary/cases.json \
  examples/assembled-length-summary/gooo.workspace.json
```

선택용 사례는 1개, 별도 보류 사례는 1개이며 `cases.json`의 실제 입력은 3개입니다.
세 분모를 합쳐 일반적인 정답률로 표시하지 않습니다. 실제 입력에는 빈 라벨 목록,
한글·이모지, 처리 상한을 넘는 열 개의 라벨이 들어 있습니다. 마지막 입력의
`complete=false`는 여덟 개까지만 바이트를 합산했다는 뜻입니다.

새 반환식 기록은 `return Summary{...}` 자체에 실행 주체를 부여합니다. 관찰을
위해 사용자가 별도의 지역 변수를 만들 필요가 없습니다. 각 필드에는 선택된
식의 범위와 원래 `choice ... alternative`의 파일 위치가 함께 남습니다.

[Gooo로 선택의 기원 읽기](../record-choice-origins/README.md) ·
[언어 설명과 현재 범위](../../docs/language/return-choice-origins.ko.md)
