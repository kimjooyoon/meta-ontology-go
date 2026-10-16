# 이름을 바꿔도 이어지는 필드

코드에서 읽기 좋은 이름은 시간이 지나며 달라집니다. Gooo는 필드의 안정 ID로
의미를 구분합니다. 레지스트리를 Go 구조체로 옮기는 연결기도 이 ID를 읽도록
정리했습니다.

예를 들어 다음 필드의 이름은 `Subject`입니다.

```gooo
field Subject id "gooo://metric-registry/CohortTerminalDeclaration/Subject" type string required one
```

작성자가 역할을 더 잘 드러내는 이름으로 고치면 다음처럼 쓸 수 있습니다.

```gooo
field ExpectedSubject id "gooo://metric-registry/CohortTerminalDeclaration/Subject" type string required one
```

Go 쪽 필드는 `gooo` 태그로 같은 ID를 가리킵니다. Go 필드 이름도 독립적으로
바꿀 수 있습니다. 다음은 연결을 설명하는 구조체 조각입니다.

```go
type CohortTerminalSpec struct {
    ChosenSubject string `gooo:"gooo://metric-registry/CohortTerminalDeclaration/Subject"`
}
```

연결기는 원본 레코드 ID를 고르고, 각 필드 ID와 타입을 대응시켜 대입 코드를
생성합니다. 표시 이름이 같아도 ID가 다르면 연결을 거부합니다. 중복된 ID나 Go
필드 이름, 태그 누락, 타입 변경도 거부합니다. 태그는 선언 전체를 읽으며, 같은
메타데이터 키의 반복과 뒤쪽의 잘못된 문법도 거부합니다. 선택적 값과 배열은 이 단일 필드
연결 범위에서 명시적으로 거부합니다.

## 실제 적용 범위

`scripts/metric-registry-codegen`의 세 레코드 연결에 적용했습니다. 생산용 Gooo
원본과 생성된 세 파일의 바이트는 유지합니다. Go 구조체의 기존 JSON 이름도
유지합니다. 연결기의 기록 선택과 필드 대응은 안정 ID를 사용합니다.

`At`와 `Count` 활동을 찾는 설정은 기존 이름을 사용합니다. 중첩 레코드와 배열의
전송, 일반 외부 Go 구조체를 위한 공통 연결 문법은 후속 설계 범위입니다.

## 확인한 것

초기 연결기의 새 이름·ID·타입 경계 사례 23개와 생성된 연결 코드의 문자열·불리언·정수
타입 검사 3개가 로컬에서 통과했습니다. 실제 Gooo 레코드와 필드를 바꾸고 Go 필드도
다른 이름으로 바꾼 입력이 생성 도구를 통과하는 사례를 포함합니다. 원본 ID는
유지하며, 바뀐 입력의 파일 내용과 표시 이름은 해당 입력에 남습니다.

이후 한 필드의 중복 ID와 무시되는 태그 뒷부분 사례 3개를 재현하고 고쳤습니다.
해당 회귀 사례 3개와 새로운 태그 문법 사례 12개가 통과했습니다. 먼저 끝난 26개
사례는 로컬에서 재실행하지 않았습니다. [PR #1617의 원래 CI](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38863565314)에서
최종 소스의 전체 41개 사례와 생산용 코드 재생성 일치를 확인했습니다. 이 관찰의
소스는 `c3f78d98366e6a53900b4a23cfa1969158876d01`입니다. 같은 PR의 실제 생산·소비
경로에서 반례 16개도 통과했습니다. 이름 변경 사례는 생성 코드의 타입 검사이며,
이름을 바꾼 프로그램의 실제 실행까지 보장하는 범위로 넓히지는 않습니다.
기존 반례 실험이나 모델 실행은 로컬에서 다시 수행하지 않았습니다.
