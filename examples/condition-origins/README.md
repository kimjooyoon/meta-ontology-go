# 출처를 Gooo로 읽기

`classify.gooo.fixture`는 실행 위치의 출처 정보를 받고 짧은 설명을 돌려준다.
실행 기록을 다시 언어의 입력으로 쓰는 작은 메타프로그래밍 예제다.

`bodyexecution.ConditionOriginFacts`는 출처 링크의 유무와 필드 참조 수를
평평한 레코드 배열로 옮긴다. 분류 조건과 문자열은 Gooo 소스가 정한다.

| 입력 필드 | 뜻 |
| --- | --- |
| `site_id` | 순수 투영의 실행 위치 ID |
| `selected` | 선택된 Gooo 조건의 위치가 연결됨 |
| `original` | 원본 패키지 파일의 조건 위치가 연결됨 |
| `declaration` | 선택된 activity 선언이 연결됨 |
| `ref_count` | 정적으로 참조한 필드 수 |
| `located_count` | 선택된 소스의 필드 선언에 연결된 참조 수 |

원본 연결이 있고 선언이 모두 연결되면 `source-retained`, 선택된 소스까지
연결되면 `selected-source`, 필요한 연결이 부족하면 `origin-unresolved`를 반환한다.
새로운 분류나 후속 동작은 이 소스의 조건문을 바꿔 구성할 수 있다.

현재 개발 소스에서 이 예제를 사용하는 새 테스트는 컴파일러가 만든 출처 입력 한 개와
명시적인 분류 입력 네 개를 구분해 기록한다. [관찰 및 생산자](observations.json)에
최초 실패도 함께 남겼다. 테스트 기록은 실행된 입력에 대한 관찰이다.
`fields`는 현재 문법의 예약어여서 이 예제는 `ref_count`를 필드 이름으로 사용한다.
문자열 결과는 표준 문자열 ID를 가진 `Text`로 선언한다.

개발 소스의 저장소 루트에서 실행하는 명령은 다음과 같다. 이 묶음은 API로 관찰한
동일한 Gooo 본문과 입력을 CLI 파일 형식으로 제공한다. CLI 묶음을 로컬에서 재실행한
관찰은 포함하지 않는다.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo package execute --json \
  --cases examples/condition-origins/cases.json \
  examples/condition-origins/workspace.json
```

[출처와 기원 설계](../../docs/language/source-origins.ko.md)
