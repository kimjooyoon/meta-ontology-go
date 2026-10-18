# 선택한 후보가 어디에 쓰였는지 읽기

조립한 프로그램을 살펴볼 때 “이 후보에서 고른 값이 어떤 조건과 반환값으로 이어졌나”를
물어볼 수 있다. 개발 버전의 `gooo package origins`는 저장한 실행 기록에서 그 연결을 읽는다.

```sh
gooo package origins --receipt execution.json
gooo package origins --receipt execution.json --candidate selected-rule
gooo package origins --receipt execution.json --activity urn:example:calculate --json
gooo help package origins
```

`--activity`에는 조립 기록의 activity ID를 넣는다. 후보 ID가 여러 activity에 같게 쓰인
경우 두 조건을 함께 지정하면 범위를 좁힐 수 있다. 기본 출력은 32행, `--limit`은 1~128행이다.

## 큰 실행 기록을 그대로 전달하기

개발 소스의 `package execute`, `replay`, `resume`에는 선택적인 gzip 전송이 있다.
같은 JSON 인코더의 출력을 압축하므로 결과, 원본 위치, 관측 누락 수와 `PARTIAL` 상태가
모두 남는다. 기본 `--json` 출력은 기존 JSON이다.

```sh
gooo package execute --json --receipt-compression gzip \
  --cases cases.json gooo.workspace.json > execution.json.gz
gooo package replay --json --receipt-compression gzip \
  --receipt execution.json.gz --cases cases.json gooo.workspace.json > replay.json.gz
gooo package origins --receipt execution.json.gz --json
```

`--receipt-compression gzip`은 `--json`과 함께 사용한다. 읽기는 확장자가 아닌 gzip
헤더로 구분한다. `--origin-receipt`, `--construction-receipt`도 같은 읽기를 지원한다.
소비 에이전트는 압축 출력임을 명시적으로 선택하고 종료 코드를 먼저 확인해야 한다.
`gooo version --build --json`의 `package_receipt_transport`에는 지원 압축 형식,
JSON 모드 요구사항, 두 바이트 한도와 정체성 기준이 들어 있다. 이 필드가 없는 이전
컴파일러에 압축 옵션을 추정해서 전달하지 않는다.

- gzip 전송 한도는 **2MiB**, 복원한 JSON 한도는 **32MiB**다. 두 한도를 별도로 검사한다.
- 출력과 입력을 읽는 중에도 누적 크기를 제한한다. gzip은 한 멤버만 받고 체크섬,
  끝맺음, 뒤에 붙은 바이트를 확인한 후 기존 JSON·스키마·실행 정체성 검사를 한다.
- 한도를 넘거나 쓰기가 실패하면 명령이 실패한다. 남아 있는 부분 출력은 완성된 기록이 아니다.
- `replayed_from_sha256`과 출처 조회의 `receipt_sha256`은 **복원한 원래 JSON 바이트**를
  가리킨다. 전송 파일의 해시가 필요하면 별도로 보관한다.

이 기능은 실행을 더 많이 관측하거나 이미 누락된 이벤트를 복구하지 않는다. JSON 객체와
실행 기록은 여전히 메모리에 올라가므로 32MiB는 전체 프로세스 메모리 상한이 아니다.
큰 반복 기록을 저장·전달하는 비용을 줄이는 기능이며, 관측 예산과 기록 구조 개선은 별도다.

## 경로를 읽는 방법

예를 들어 `let eligible = input > 53`이라는 대입과 `if eligible`이라는 조건이 있으면,
조회는 조건에서 출발해 `eligible`의 대입과 선택한 기준값으로 돌아가는 경로를 보여준다.
반환값이 그 분기 안에서 바뀌었다면 반환식에서 분기 조건까지 이어지는 경로도 나타난다.

- `input`: 값의 계산에 사용한 선행 대입
- `control`: 해당 대입이 속한 분기를 정한 조건
- `direct`: 조건·반환식 안에 후보 조각이 직접 들어 있음
- `static_dependency`: 선행 대입이나 제어 조건을 거쳐 이어짐

조건·반환식과 빈칸의 쌍마다 **기록된 최단 경로 하나**를 보여준다. 분기 양쪽처럼
가능한 경로가 여러 개일 수 있다. 실제 입력이 지나간 경로는 실행 기록에서 별도로 읽는다.
같은 변수 이름이 다른 블록에 있어도 선언의 구문 정체성으로 연결한다.

각 위치에는 문서의 종류와 해시가 함께 있다. `workspace_input_file`은 최초 패키지 파일,
`assembly_input`은 해당 조립 단계에 들어온 문서를 가리킨다. 원본 파일의 위치가 조건에
연결돼 있으면 같은 activity·후보·빈칸의 반환 경로에도 그 선언 위치를 함께 보여준다.
개발 소스는 반환식만 가진 본문에도 최초 파일의 후보 위치를 보관한다. 이전 실행 기록에
원본 연결이 없으면 기록된 조립 위치까지만 표시한다.

함수 호출과 `bind`를 따라가려면 `--follow`를 사용한다.
[활동 사이의 값 기원 안내](value-origins-across-activities.ko.md)에 출발점 선택과 분석 범위를 설명한다.

## 상태와 범위

`analysis_status`와 `issues`는 저장된 의존 관계 분석의 제한 사항이다. 반복문 종료 경로,
다른 함수 내부, 별칭 등을 완전히 펼치지 않은 경우 `PARTIAL`과 이유를 유지한다.
부분 분석에서는 확인된 연결과 미해결 이유를 함께 읽는다.

과거 기록에 그래프가 없으면 `NO_RETAINED_VALUE_GRAPH`를 보여준다. 조회 과정에서
과거 본문을 다시 조립해 기원을 덧붙이지 않는다. 출력이 잘리면 `truncated`가 `true`가 된다.
생성 단계는 최대 128개, 개별 그래프는 256개 노드, 반환 경로의 전체 길이는 2,048단계로 제한한다.
긴 식의 표시 문자열은 256바이트까지 보여주고 원래 범위는 유지한다.

JSON의 `receipt_sha256`은 읽은 기록 전체를 가리킨다. 선택된 본문의 저장 소스·해시와
경로의 바이트 범위를 대조해 식을 표시한다. 이 명령은 파일 조회와 구문 해석만 수행한다.
프로그램·후보·모델 실행은 0회다. 새 소스의 변경 영향은 기존 `gooo impact`로 살펴볼 수 있다.

## Gooo 메타프로그램으로 이어 쓰기

[`OriginUseFacts`](../../internal/packageruntime/workspaceexecution/origin_query_facts.go)는
루트와 출처 위치 한 쌍을 13개 필드의 typed record로 전달한다. `kind`, `relation`,
`analysis_status`, `source_view`와 후보·빈칸·activity ID를 Gooo 활동에서 읽을 수 있다.
원래 보고서의 `truncated`와 `issues`도 함께 확인해야 한다.

[실제 사례](../../examples/recorded-origin-queries/README.md)의 Gooo 프로그램은
조건에 쓰인 빈칸과 반환값으로 이어진 빈칸을 서로 다른 문장으로 설명한다.
출처에서 확인한 사실을 읽고, 그 의미를 언어 안에서 다루는 작은 메타프로그램이다.
