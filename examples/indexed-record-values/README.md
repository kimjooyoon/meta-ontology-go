# 인덱스 계산의 기원을 보존하기

`Read`는 항목 목록을 지역 변수에 받고, `offset + 1`로 고른 항목의 `value`를
반환한다. 출처 조회는 목록과 인덱스를 계산한 입력을 함께 보존한다.
`ignored`와 `unused`는 선택한 결과의 데이터 기원에 포함되지 않는다.

```sh
go run ./cmd/gooo body-origins --activity Read examples/indexed-record-values/main.gooo
go run ./examples/record-origin-reader --activity Read examples/indexed-record-values/main.gooo
go run ./cmd/gooo body-compose --source examples/indexed-record-values/main.gooo \
  --entry Read --cases examples/indexed-record-values/cases.json
```

정적 노드의 `source_index`는 원문에 작성된 인덱스식이다. 실제 입력에서 계산한
배열 위치는 실행 기록에서 별도로 확인한다. 한글 지역 이름 `커서`도 원본 바이트
위치에 연결된다. 기존 출처 탐색기가 원래 노드와 필드 ID를 읽는다.

작성한 네 값 기대에는 서로 다른 위치, 같은 값을 가진 두 항목, `2^53`보다 큰
정수가 포함된다. 정적 출처가 해석돼도 실행의 배열 범위 검사는 유지된다.
현재 경로는 필수 배열 필드를 다룬다. [직접 만든 배열](../constructed-record-values/README.md)은
별도 예제로 확인한다. 선택 필드의 정적 출처는 남은 범위이며, 512노드 한도를
넘으면 부분 성공 결과를 내보내지 않는다.
