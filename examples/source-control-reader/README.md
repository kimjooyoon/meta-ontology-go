# 조건 분기의 출처를 Gooo로 읽기

`source.gooo`에는 두 조건이 있습니다. 첫 번째는 `false` 상수이고,
두 번째는 실행 입력의 `enabled`입니다. 이 소스를 실행하기 전에 어떤
분기까지 알 수 있는지 `inspect.gooo`가 읽습니다.

```sh
gooo body-origins --activity Adjust \
  --inputs-for source/control:Inspect --inputs-shape source-control-columns \
  examples/source-control-reader/source.gooo > /tmp/gooo-control-inputs.json
gooo package execute --inputs /tmp/gooo-control-inputs.json \
  examples/source-control-reader/gooo.workspace.json
```

출력은 하나의 JSON 값입니다. 이 예제의 분기 노드 4개 중 1개는 참,
1개는 거짓, 2개는 입력에 따라 달라집니다. 원래 소스의 해시, 활동 ID,
보관된 그래프의 해시도 함께 반환합니다.

| 값 | 읽는 방법 |
| --- | --- |
| `guard_count` | 이번 그래프에 보관된 분기 노드 수 |
| `known_true`, `known_false`, `unknown` | 이 분기 노드의 정적 판독 결과 |
| `join_count` | 보관된 제어 범위 합류 노드 수 |
| `out_of_scope_nodes` | 분기 판독 대상에서 제외한 값·호출 등의 노드 수 |
| `true_scope_bits`, `false_scope_bits` | 참·거짓으로 읽은 원래 분기/합류 노드의 위치 |
| `reason` | 입력을 읽을 수 없을 때의 이유 |

Go 어댑터는 IR의 노드 ID, 부모, 조건, 분기 방향, 알려진 상수 진리값을
그대로 옮깁니다. Gooo의 `Validate`는 열과 참조를 확인하고,
`Inspect`는 바깥 조건과 현재 조건을 함께 읽어 결과를 구성합니다.
외부에서 직접 입력한 해시는 전달받은 식별자로 보관합니다. 파일과의
연결을 확보하려면 위의 소스 조회 명령으로 입력을 만듭니다.

`unknown`은 실행 입력이 필요하다는 뜻입니다. 예를 들어 알려지지 않은
조건의 참·거짓 분기를 합친 경우, 현재 판독기는 기호적 분할 증명을
추가하지 않고 알 수 없음으로 유지합니다. 결과를 코드 제거에 사용하려면
해당 변환이 지켜야 할 별도의 의미 보존 조건이 필요합니다.

정적 판독 범위의 분모는 `guard_count`입니다. 위 예제에서는 2/4이며,
입력 의존 분기를 정상적으로 보존합니다. 분기 0개는 `NO_RETAINED_GUARDS`로
표시합니다. 그래프는 60노드, 단일 내보내기는 32KiB까지 받고 초과하면
거절합니다. 입력을 잘라서 완전한 그래프처럼 보고하지 않습니다.

새로 작성한 유한 사례 15개는 다음 명령으로 확인할 수 있습니다.

```sh
gooo package execute --cases examples/source-control-reader/cases.json \
  examples/source-control-reader/gooo.workspace.json
```

첫 로컬 실행에서 15/15가 맞았습니다. 그중 1개는 위 소스의 내보내기이고,
14개는 상수·중첩·합류·누락 입력을 다루는 통제 사례입니다.
이 수치의 범위는 위 15개 사례입니다.

이 판독은 소스 조건에 관한 관찰입니다. 실행 중 값의 생성과 사용은
[기록된 기원 읽기](../recorded-use-origins/README.md)에서 별도로 다룹니다.
PROV-O의 [사용](https://www.w3.org/TR/prov-o/#used)과
[파생](https://www.w3.org/TR/prov-o/#wasDerivedFrom)을 구분하면서,
보관한 사실이 표현하는 관계만 Gooo에서 조립하는 방향입니다.

## 같은 타입의 입력을 선언 위치로 구분하기

`Difference(Integer, Integer)`에서는 두 입력의 타입 ID가 같습니다.
원래 파일의 해시, 활동 ID, 입력 순서와 바이트 범위를 함께 보관하면
각각 어느 자리에 선언됐는지 다시 찾을 수 있습니다.

```sh
gooo body-origins --activity Difference --format text \
  examples/source-control-reader/input-positions.gooo.fixture
gooo body-origins --activity Difference \
  --inputs-for source/control:InspectInputs --inputs-shape input-declaration-columns \
  examples/source-control-reader/input-positions.gooo.fixture > /tmp/gooo-signature-inputs.json
gooo package execute --inputs /tmp/gooo-signature-inputs.json \
  examples/source-control-reader/inputs.workspace.json
```

첫 명령은 `input0`과 `input1`을 각각 원본 6행과 7행의 `Integer`에 연결합니다.
마지막 명령의 결과는 `SHARED_TYPE_DISTINCT_POSITIONS`, 입력 2개,
같은 타입 쌍 1개입니다. Go는 선언 사실을 옮기고, Gooo의 `InspectInputs`가
열의 길이·순서·위치를 확인한 뒤 타입 ID를 비교합니다. 표시 이름이 달라도
타입 ID가 같으면 같은 타입으로 읽습니다.

레코드의 여러 필드가 한 입력에서 왔다면 선언 행은 하나입니다.
본문의 필드별 기원은 계속 각 필드 ID로 구분합니다. 선언 좌표와
본문 표현 좌표는 각각 보관하므로, 본문 위치가 없는 입력 노드에도
원래 시그니처 위치를 표시할 수 있습니다.

현재 범위는 입력 1~16개, 내보내기 32KiB입니다. 위치는 이번 소스 안의
순서를 뜻합니다. 작성자가 붙이는 입력 이름과 안정적인 역할 ID는
후속 언어 설계에서 다룹니다. 입력 칸을 바꾸면 그 위치의 의미도 바뀝니다.
직접 만든 열의 해시는 받은 식별자로 보관하므로, 파일과 연결하려면
위의 소스 조회 명령으로 입력을 만듭니다.

첫 로컬 관찰에서 새 통제 사례 16개가 모두 기대 결과와 일치했습니다.
16개 입력의 같은 타입 쌍 120개, 표시 이름 변경, 서로 다른 타입,
겹치는 위치, 중복 포트, 누락 열과 범위 초과를 포함합니다.
별도의 실제 소스 내보내기에서도 두 입력의 선언 위치를 유지했습니다.
모델 호출은 0회입니다. [실행 기록](../../docs/observations/input-declarations/observations.json)에
실패한 초안과 수정 후 첫 실행의 원본 기록을 함께 보관합니다.

현재 판독기는 `for left in input.indices` 안에서 같은 목록의 `right`를
순회합니다. 두 목록의 위치를 비교하는 의도를 그대로 읽을 수 있습니다.
이 배열 중첩 순회는 기존 문법이며, 횟수를 지정하는 중첩 반복문에는
별도의 지원 제약이 있습니다. [현재 소스의 관찰](../../docs/observations/input-declarations/collection-observations.json)에서도
같은 16개 사례가 모두 통과했고, 보관한 이전 수정본과 출력이 일치했습니다.
두 수정본에서 확인한 사례의 분모는 각각 16개입니다.
