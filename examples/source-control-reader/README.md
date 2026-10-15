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
