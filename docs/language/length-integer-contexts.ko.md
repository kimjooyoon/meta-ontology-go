# 길이를 Integer 문맥에 연결하기

배열의 행 수와 정수 카운터를 비교하거나 길이를 결과 필드에 넣을 때,
Gooo 소스에서 길이의 구현 자료형까지 적을 필요가 줄었습니다.

```gooo
activity ByteCount(Text)->Integer computes `return len(input)`
activity BelowLength(Text,Integer)->Boolean computes `return input1<len(input0)`
```

반환, 정수 연산·비교, 대입, 레코드의 Integer 필드, 순수 활동의 Integer 인자가
길이 값을 받을 수 있습니다. `len`은 Text의 UTF-8 바이트 수 또는 배열 원소 수를
셉니다. 한글 두 글자는 바이트 수 6이 될 수 있으며, 글자 수가 필요한 기능은
별도의 문자열 연산으로 정의해야 합니다.

## 기존 생성물과 새 문맥

이미 타입이 맞는 프로그램은 길이 변수의 생성 형태를 유지합니다. 기존 명시적
`int64(len(...))`도 계속 사용할 수 있습니다. 길이와 Integer가 만나는 자료형
불일치를 실제로 해결할 때, 컴파일러는 필요한 확장을 넣고 관련 추론 길이
변수를 64비트 Integer로 맞춥니다. 새로 허용되는 변수는 지역 변수 기원 기록에도
참여합니다.

```gooo
let count=len(input.labels)
let bytes=0
for let row=0;row<8;row++ {
 if row<count {bytes=bytes+len(input.labels[row])}
}
```

이 예제는 횟수가 정해진 반복문의 기존 상한을 유지합니다. 배열이 8개보다 길면
앞의 8개만 처리하며, [전체 예제](../../examples/length-aware-summary/README.md)는
그 경우를 `complete=false`로 돌려줍니다.

길이에서 기원한 값을 확장하는 규칙입니다. Boolean·Text 문맥과 다른 호스트
정수의 자료형 오류는 최종 타입 검사에서 거부합니다. 소스에 직접 선언한
`len` 활동은 자신의 활동 ID와 호출 관계를 유지합니다.

## 원본과 생성물의 위치

컴파일러가 추가한 확장은 생성물에 보입니다. 원본 식·선택 후보·기호 의존성
그래프에는 작성한 `len(...)`과 연산을 남깁니다. 확장이 원본에 있었다고
기록하지 않습니다. 추가 노드 수는 소스 의미 단위 계산에서 구분합니다.

새 입력의 실제 실행, 초기 실패와 적용 범위는
[관찰 기록](../observations/length-integer-contexts/README.md)에 보관합니다.
