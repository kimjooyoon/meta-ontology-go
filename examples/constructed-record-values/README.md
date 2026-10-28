# 직접 만든 배열의 기원 읽기

`Read`는 `left`와 `right`로 두 항목을 만들고, `pick`으로 고른 항목의 `value`를
반환한다. 계산은 같은 값이어도 각 항목의 원문 위치와 원래 필드 ID를 유지한다.

```sh
go run ./cmd/gooo body-origins --activity Read examples/constructed-record-values/main.gooo
go run ./examples/record-origin-reader --activity Read examples/constructed-record-values/main.gooo
go run ./cmd/gooo body-compose --source examples/constructed-record-values/main.gooo \
  --entry Read --cases examples/constructed-record-values/cases.json
```

계산된 인덱스의 값 기원은 `left`, `right`, `pick`이다. 읽지 않은 `noise`는 제외한다.
인덱스를 상수 `0`으로 바꾸면 `left`만 따라간다. 배열 전체를 반환할 때는 모든
항목 필드가 값에 포함되므로 `noise`도 전체 배열의 기원에 포함된다.

같은 기존 Gooo 탐색기가 일반 부모 관계와 반복 전달 관계를 읽는다. 배열을
지역 이름으로 복사하거나 레코드 필드에 저장하고 조건 분기로 교체해도 항목의
필드별 기원을 보존한다. 조건 입력은 제어 기원을 요청했을 때 별도로 따른다.
중첩 레코드의 필드와 스칼라 배열도 다룬다. 원문에 작성된 순서와 상수 인덱스는
정적 출처이며, 실제 실행의 원소 읽기 기록과 구별한다.

형태 정보와 전체 값의 루트를 분리한다. 반복 여부와 원소 선택은 형태를 따르고,
필드 읽기는 해당 필드의 값만 따른다. 순수 스칼라 배열의 조회에는 비어 있는
레코드 사전을 허용하며, 존재하지 않는 값·버전 참조는 계속 거부한다. 이 조회의
단계 상한은 Gooo가 계산한다.

작성한 값 기대 4개는 서로 다른 위치, 같은 값을 가진 두 항목, `2^53`보다 큰
정수를 포함한다. 여러 실행 경로로 확인해도 이 분모는 4개다. 기존 512노드,
64바인딩, 16레코드 타입과 배열·입력 한계를 유지한다. 노드 한도를 넘은 출처는
부분 결과를 내보내지 않으며 기존 코드 생성 경로는 유지한다.

현재 정적 출처는 순서대로 작성한 배열 리터럴과 읽기 전용 순회를 다룬다.
빈 리터럴의 순회는 원문상 실행될 항목이 없어 입력 기원을 만들지 않는다.
빈 배열의 인덱스 읽기, 또는 반복문을 거치며 배열 자체를 교체하는 출처는 명시적
미해결 상태로 남긴다. 실행 시의 잘못된 인덱스 검사도 별도로 유지한다.
이 범위는 [#1849](https://github.com/kimjooyoon/meta-ontology-go/issues/1849)에서 추적한다.
