# 문법에서 온 후보 읽기

`source.gooo`는 조건식과 정수식의 작은 문법을 선언한다. 먼저 소스가 보관한
후보 ID와 대입을 읽는다. 명령은 후보를 골라 본문을 만들거나 사례를 실행하지 않는다.

```sh
gooo body-origins --activity Choose --fill-candidates \
  examples/derived-candidate-origins/source.gooo
```

이 예제의 문법은 조건식 2개와 정수식 3개를 조합한다. 가능한 조합 6개 중
`max_candidates "4"`로 4개를 보관하고 2개를 생략한다. 이는 문법 조합의
범위이며, 의도에 대한 정답률은 별도다. 출력의 후보 ID 중 하나로 기원을 읽는다.

```sh
gooo body-origins --activity Choose --fill-candidate <출력된-ID> \
  examples/derived-candidate-origins/source.gooo
gooo body-origins --activity Choose --fill-candidate <출력된-ID> \
  --inputs-for origin/derived-candidates:Inspect \
  examples/derived-candidate-origins/source.gooo > derived-origins.json
gooo package execute --json --inputs derived-origins.json \
  examples/derived-candidate-origins/gooo.workspace.json
```

생성된 식에는 소스에 직접 쓴 식의 위치가 없다. 보고서는 실제 문법 선언과 빈자리,
본문 사용 위치를 남긴다. 후보 집합 해시와 계획 해시가 보관된 조합을 설명한다.
집합 밖의 ID는 조회가 거부되며, 그 사실만으로 문법 전체에서 불가능한 후보라고
판단하지 않는다.

Go가 전달하는 원시 행은 19개 필드다. `inspect.gooo`는 `origin_kind`, 실제 문법
범위와 본문 사용 여부를 읽어 `DERIVED_ASSIGNMENT`와 `UNRESOLVED_USE`를 구분한다.
직접 식의 좌표는 0이며, 문법 기원이라는 표지가 그 빈 좌표의 의미를 구분한다.
기존 직접 선언 후보의 v1 보고서와 15개 필드 입력은 그대로 유지한다.

입력 전용 소비자의 `finite_total=0`이다. 원래 후보 프로그램과 모델은 실행하지
않는다. 소스의 1..8개 빈자리·2..16개 보관 후보·128KiB 소스·32KiB 입력 내보내기
범위를 따른다. 없는 사용 위치는 비워 두고 미해결 이유를 보관한다.

[사용 범위](../../docs/language/derived-candidate-origins.ko.md) ·
[직접 선언 후보](../fill-candidate-origins/README.md)
