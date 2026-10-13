# 소스의 동작 의도를 남기는 포맷

Gooo의 활동은 입력·결과 이름에 계산 본문, import, bind, 조립 선택지와 조건을 덧붙인다.
포맷 결과에서도 이 요소들이 남아야 다음 구성과 실행이 같은 의도를 읽을 수 있다.

기존 명목 포맷 경로를 새 소스 예제에 적용한 결과, 성공 상태로 출력됐지만 import,
계산 본문과 bind가 사라졌다. 이 관찰의 원본과 출력은
[예제의 이전 경로 기록](../examples/source-surface-format/baseline-observation.json)에 남겼다.
원본 파일을 직접 바꾸지는 않았다.

## 전체 소스 경로

이 변경을 포함한 개발 소스의 `gooo format`과 `gooo fix`는 import, bind, computes 또는
assembling 또는 entity fields가 있는 파일에 전체 syntax formatter를 사용한다.
명목 선언만 있는 파일은 기존 포맷 경로를 유지한다.
명목 formatter의 직접 AST 어댑터는 표현할 수 없는 확장 요소를 거절한다.

```sh
gooo format --json retained.gooo.fixture
```

출력은 변경 제안이다. 입력 파일은 그대로 두고 `direct_writes`는 0이다.
포맷 후 다시 파싱해 전체 소스 carrier를 비교한다. 이름, ID, 타입 참조, import 경로,
계산 본문 바이트, bind 포트와 조립 의도가 비교 대상이다.
소스 위치와 import 정렬 순서는 표시 정보로 제외한다.
조립은 기존 source-owned `Spec.Canonical` 계약으로 비교한다.
불일치나 표현하지 못하는 carrier는 변경 제안을 반환하지 않는다.

이 비교는 파싱한 소스의 보존 범위를 나타낸다. import 대상 해석, 실제 값의 계산,
모델 선택 결과와 실행의 동등성은 각각의 컴파일·관찰 과정에서 확인한다.
깊이 128을 넘는 carrier나 지원하지 않는 carrier 종류도 명시적인 오류가 된다.
레코드 소스는 패키지 실행과 같은 명시적 V4 scalar-field 프로필로 읽고 다시 쓴다.
필드의 ID, 이름, 타입 표기, required/optional, one/many와 선언 순서도 carrier 비교에 포함된다.
이 경로는 필드 타입을 import 대상에 해석하거나 레코드 값을 실행하지 않는다.
ordinary parser의 기본 프로필과 명목 formatter의 필드 거절 규약은 그대로다.

## 새 예제에서 관찰한 범위

[새 소스 예제](../examples/source-surface-format/README.md)를 한 번 포맷한 결과 다음을 남겼다.

| 요소 | 남은 개수 / 예제에서 선언한 개수 |
| --- | --- |
| import | 1 / 1 |
| 계산 본문 | 4 / 4 |
| bind | 1 / 1 |
| assembling | 1 / 1 |
| 한정 입력·결과 참조 | 4 / 4 |

조건의 `[9]` 입력과 `[13]` 별도 관찰 입력도 그대로 남았다.
이 입력들의 결과를 평가하지는 않았다. 원본 파일은 변경되지 않았고,
프로그램 실행·모델 예측·학습은 모두 0이다.
이전 예제를 재실행하지 않고, 더 풍부한 별도 소스의 변경 제안을 기록했다.

[import 타입 가이드](qualified-activity-types.ko.md)와 함께 보면,
참조의 표현과 그 소스를 다시 편집하는 흐름을 확인할 수 있다.

## 레코드 필드도 같은 소스에 남기기

[레코드 소스 예제](../examples/record-source-format/README.md)는 required 필드 두 개와
optional 필드 한 개, import, 두 계산 본문과 bind를 함께 둔다.
패키지 실행과 포맷 경로 사이에서 필드 지원 범위가 달라 작업이 끊기는 문제를 보완했다.
새 소스를 한 번 포맷해 필드 3/3개와 각 필드의 이름·ID·타입 표기·필수 여부·개수·순서를
보존했다. import 1/1개, 계산 본문 2/2개, bind 1/1개도 남았다.
원본 변경과 레코드 실행·모델 예측·학습은 모두 0이다.
이 관찰은 실제 개발 소스와 새 바이너리에 연결돼 있으며 이전 예제의 제작 기록은 유지한다.
