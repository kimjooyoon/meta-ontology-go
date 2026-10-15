# 코드에서 값의 기원을 읽기

Gooo의 값에는 작성한 소스와 입력, 그 값을 만든 연산의 연결이 있습니다.
`body-origins`의 텍스트 보기는 반환값에서 이 연결을 따라가기 위한 입구입니다.

```sh
gooo body-origins --activity Adjust --format text examples/compound-updates/main.gooo
```

출력에서 `return <- #18` 같은 줄은 반환값의 정의를 가리킵니다. 각 번호에는
원래 소스의 변수와 식, 행·열 위치가 함께 나옵니다. `data <-`는 값의 입력을,
`control <-`와 `condition <-`는 분기를 고르는 조건을 보여줍니다. 입력 포트는
자신의 이름으로 표시하며, 본문 위치가 없는 이유도 유지합니다.

이는 소스에서 가능한 의존성입니다. 실제 실행 입력이 없는 조회이므로 조건
양쪽의 정의가 함께 나타날 수 있습니다. 실행한 경로와 값의 기록은 실행 출처에서
읽어야 합니다. 이 조회는 모델 선택이나 네이티브 실행을 수행하지 않습니다.

## Gooo가 설명 규칙도 갖는다

입력, 지역 정의, 갱신, 반환, 합류 등의 표시 이름은
[Gooo 프로그램](../../internal/meta/originview/roles.gooo)이 정합니다.
컴파일러가 이 프로그램을 Go로 생성하고, 생성 파일의 해시와 원본 출처를
[생성 기록](../../internal/meta/originview/generation-origin.json)에 연결합니다.
새 종류를 해석하지 못하면 원래 종류와 미분류 상태를 표시합니다.

JSON은 기존 기본 형식입니다. `--format text`는 같은 그래프를 읽는 보기이며,
원본 기원이나 의미 상태를 바꾸지 않습니다. 긴 식은 텍스트에서 160자 뒤를
`…`로 줄여 표시합니다. 전체 식과 바이트 범위는 JSON에서 확인할 수 있습니다.
후보 목록과 입력 내보내기는 JSON 형식을 사용합니다.

## PROV-O와 이어갈 방향

[PROV-O](https://www.w3.org/TR/prov-o/)의 Entity·Activity와 생성·파생 관계는
서로 다른 시스템의 기원을 표현하는 공통 어휘입니다. Gooo에서는 작성한 소스,
컴파일로 만든 표현, 실제 실행의 값 버전을 각각 식별하는 데 이 관점을 사용합니다.
현재 텍스트 보기의 대상은 소스 그래프입니다.

다음 언어 설계는 입력의 읽기 쉬운 이름과 안정된 포트 ID, 지역 변수의 갱신 버전,
실행 출처를 함께 다루는 방향입니다. 작성자가 붙인 이름으로 코드를 이해하면서도
그 의미와 기원이 이어져야 합니다.

[현재 소스 보기](../observations/readable-origins/compound-source-view.txt)와
[새 입력 관찰](../observations/readable-origins/readable133613.json)을 함께 제공합니다.

## 컴파일러 저장소의 등록 누락을 먼저 확인하기

컴파일러 저장소에서 새 `.gooo` 프로그램을 추가할 때는 다음 조회를 먼저 사용합니다.

```sh
go run ./cmd/syntax-registration-candidate --root . --inventory > /tmp/gooo-source-inventory.json
jq -e '(.unregistered | length) == 0 and (.missing | length) == 0' /tmp/gooo-source-inventory.json
```

각 파일의 기원 해시와 등록 목록의 차이를 읽습니다. 문법 사례 재생, 네이티브 실행,
저장소 쓰기는 수행하지 않습니다. 문법과 의미 판정은 정식 검사에서 이어갑니다.
표시 규칙 소스는 메타 프로그램 영역으로 옮겼습니다. 내용과 생성된 Go 바이트를
유지하고 [위치 변경 출처](../../internal/meta/originview/source-relocation.json)를 남겼습니다.
등록은 기존 프로토콜의 아홉 산출물로 수행했고 과거 분모를 보존했습니다.

등록 검사는 현재 목록의 정확성과 과거 사례의 보존을 함께 확인합니다. 과거의
사례 수는 당시 분모 파일에 고정합니다. 과거 기능의 회귀 검사는 그 기능의 ID,
위치, 앞선 사례를 확인하고 이후 등록을 허용합니다. 새 소스를 추가할 때마다
이전 기능의 사례 수까지 바꾸는 반복을 줄이기 위한 구분입니다.
