# 배열 길이와 정수 계산을 함께 쓰기

이 Gooo 프로그램은 문자열 배열의 원소 수, 앞 8개 원소의 UTF-8 바이트 수,
첫 문자열과 처리 범위를 반환합니다. 카운터·길이·레코드 필드·보조 활동을
연결하는 부분을 Gooo 소스에서 읽을 수 있습니다.

저장소 루트에서 Go 1.27.3로 실행합니다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/length-aware-summary/cases.json \
  examples/length-aware-summary/gooo.workspace.json
```

세 입력은 다음 범위를 다룹니다.

| 입력 | 행 | 앞 8행 바이트 | 전체 처리 |
| --- | ---: | ---: | --- |
| `출처`, `Gooo!`, `🦉` | 3 | 15 | true |
| 빈 배열 | 0 | 0 | true |
| 한 바이트 문자열 9개 | 9 | 8 | false |

`total`은 보조 활동 `Offset`이 행 수에 20809를 더한 값과 입력 `offset`의 합입니다.
이 상수는 새 관찰 입력을 구분하기 위해 사용합니다. 처리할 업무에 맞춰 보조
활동과 집계 규칙을 변경할 수 있습니다.

원래 길이 식과 필드 대입은 원본 Gooo 파일의 위치로 연결됩니다. 저장된
실행 기록은 [20809 관찰](../../docs/observations/length-integer-contexts/native20809/length-native20809.json.gz)에 있습니다.
이 실행에는 모델 호출이 없습니다.

[길이와 Integer 문맥](../../docs/language/length-integer-contexts.ko.md) ·
[관찰과 남은 범위](../../docs/observations/length-integer-contexts/README.md)
