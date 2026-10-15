# 같은 문자열, 서로 다른 입력의 기원

개발 소스용 예제입니다. JSON에는 `🧭`를 그대로 쓰거나 `\ud83e\udded`로
쓸 수 있습니다. Gooo가 받는 문자열은 같고, 작성자가 보낸 바이트는 다릅니다.
이 두 가지를 따로 보존합니다.

[main.gooo](main.gooo)는 입력을 그대로 돌려주고 UTF-8 바이트 수를 계산합니다.
다음 두 입력은 각각 `{"text":"기원:🧭�","utf8_bytes":14}`를 만듭니다.
실행 영수증의 `inputs_digest`는 파일 원문을 가리키므로 두 입력에서 서로 다릅니다.

```sh
go run ./cmd/gooo package execute --json \
  --inputs examples/text-input-origins/raw-input.json \
  examples/text-input-origins/gooo.workspace.json

go run ./cmd/gooo package execute --json \
  --inputs examples/text-input-origins/escaped-input.json \
  examples/text-input-origins/gooo.workspace.json
```

## 입력 규칙

| 입력 | 처리 |
| --- | --- |
| 올바른 원문 UTF-8 | 문자열 값을 보존 |
| BMP 문자 이스케이프와 올바른 surrogate 쌍 | 해당 Unicode 문자로 해석 |
| 명시적인 `\ufffd` 또는 `�` | 유효한 문자 U+FFFD로 보존 |
| 단독 high/low surrogate | 원래 JSON 바이트 위치와 이유를 보고 |
| 문자 그대로의 `\\ud802` | 역슬래시와 `ud802` 문자로 보존 |
| 잘못된 원문 UTF-8 | 기존 입력 검사에서 거부 |

Text의 한도는 **디코딩한 UTF-8 1,024바이트**입니다. 줄 끝, 결합 문자, 대소문자와
Unicode 정규화 형태는 그대로 유지합니다. JSON 이스케이프를 문자로 해석한 뒤에도
원래 입력 파일의 해시는 유지합니다.

판정은 [Gooo 규칙](../../internal/meta/textinput/escapes.gooo)에 있습니다.
Go 스캐너는 문법 검사를 통과한 JSON에서 16비트 단위와 인접한 단위를 읽고,
이 활동이 정상 문자·쌍·단독 high·단독 low를 구분합니다. 그 Go 투영의 소스와
생성 기록은 함께 보관합니다.

## 적용 범위와 이전 동작

이 규칙은 composition 입력, 유한 사례, 사례 묶음의 원문 JSON 문자열과 이름에
적용합니다. 프로그램 안에서 직접 전달한 Text 값과 반복 Text 필드도 같은
문자 값 검사를 거칩니다. 패키지 실행에서 거부된 입력은 원문 `inputs_digest`를
남기고 실행 결과를 만들기 전에 멈춥니다.

[#1580](https://github.com/kimjooyoon/meta-ontology-go/issues/1580)은 이전 버전에서
단독 surrogate가 U+FFFD로 바뀌는 동작을 관찰했습니다. 이번 개발 소스는 그 입력을
명시적으로 거부합니다. 이전 버전으로 만든 기록은 당시의 관찰로 보존합니다.

[RFC 8259 §8.2](https://www.rfc-editor.org/rfc/rfc8259#section-8.2)는 JSON 문법에
단독 surrogate가 나타날 수 있고 처리 결과가 서로 다를 수 있음을 설명합니다.
[Go JSON 디코더](https://pkg.go.dev/encoding/json#Unmarshal)는 이런 입력을 U+FFFD로
대체할 수 있습니다. Gooo의 composition 경계는 위 표의 명시적인 규칙을 사용합니다.

## 기원과 값의 동등성

`inputs_digest`는 **받은 JSON 문서**의 기원입니다. 공백이나 이스케이프 표기가
바뀌면 그 해시도 바뀝니다. 디코딩한 문자열의 동등성은 실행이 다루는 **값**의
관계입니다. 같은 값에서 서로 다른 입력과 실행 회차를 하나로 합치지 않습니다.

[지역 값 버전 예제](../local-version-origins/README.md)는 실행 이후의 선언·호출·갱신
회차를 구분합니다. 이 예제는 그 앞단인 입력 바이트와 Text 해석의 경계를 다룹니다.

최초 새 검사는 단위 범위 12건, 정상·비정상 문자열, composition 경계 세 곳,
두 원문 문서의 실제 실행, 잘못된 입력 네 건의 조기 거부와 원문 해시 보존을
확인했습니다. 두 실제 실행은 입력 관찰이며 별도 유한 사례 점수는 없습니다.
정확한 소스와 시간은 [관찰 기록](../../docs/observations/text-input-origins/observations.json)에 남깁니다.
