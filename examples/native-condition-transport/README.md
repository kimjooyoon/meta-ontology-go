# 실행 결과와 설명 기록을 따로 보내기

조건 경로의 첫 구현은 반복문에서 기록이 커져 기존 결과 통신의 64KiB 한도를 넘었다.
[원래 PR #1527](https://github.com/kimjooyoon/meta-ontology-go/pull/1527)의 CI가 이 문제를 찾았다.
그 소스와 최초 실패를 유지하고, 별도 소스에서 수정한 기록 방식의 동작을 관찰했다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/native-condition-transport/cases.json \
  examples/native-condition-transport/workspace.json
```

이 예제는 패키지 경로를 포함한 입력명과 수량·처리 가능 수량·요청 번호를 선언한다.
조건과 승인 후보는 Gooo에 있다. 실행기는 순수 투영을 별도로 관찰해
조건의 실제 참·거짓, 활동 ID, 원래 투영의 위치를 남긴다.

| 새 관찰 | 결과 |
| --- | --- |
| quota1289 실제 CLI | 여섯 입력의 기대값 12/12, 조건 평가 10건, 기록 `COMPLETE` |
| 같은 CLI의 구성 | 후보 선택 1/1, 별도 평가 1/1, 모델·제공자 호출 각각 0회 |
| loops1213 별도 반복 소스 | 입력 80개의 기대값 80/80, 조건 평가 총 81,644건 |
| 반복 소스의 설명 | 기록 65,536건, 누락 16,108건, 기록 상태 `PARTIAL` |
| thresholds1247 별도 작은 소스 | 기대값 2/2, 조건 평가 2건, 기록 `COMPLETE` |

[observation.json](observation.json)과 [trace.json](trace.json)은 새 CLI의 원래 관찰이다.
[loops1213.json](loops1213.json), [thresholds1247.json](thresholds1247.json)은 서로 다른 API 소스 관찰이다.
[development-observations.json](development-observations.json)은 각 생산자와 최초 실패의 관계를 담는다.
새 생산자는 공개된 실패 소스 `4569d49`와 임시 Go 변경 여섯 개이며, 새 공개 커밋으로 다시 이름 붙이지 않았다.
이전 성공·실패 소스나 기존 CI 실험을 로컬에서 다시 실행한 결과는 포함하지 않는다.

## 기록을 읽는 방법

기존 결과 통신의 내용은 유지한다. 조건 기록은 실행 작업공간의 별도 파일에 쓰고,
실행기가 제한된 크기로 읽는다. 두 실행의 결과와 조건 파일을 각각 대조하고 해시를 남긴다.
작업공간을 정리할 때 관찰용 파일도 함께 제거한다.

`condition_observation`은 기록의 상태·보존 수·누락 수·파일 해시를 담는다.
각 입력의 `conditions_dropped`는 그 입력에서 누락한 조건 수다.
기록 한도에 도달해도 조건은 원래 값을 반환하고, 프로그램 결과는 계속 얻는다.
따라서 결과 80/80과 설명 `PARTIAL`을 동시에 보여 줄 수 있다.

기록은 여전히 전체 `if` 조건의 참·거짓이다. 개별 피연산자와 지역 값,
원래 Gooo 본문의 위치 대응, 여러 사실의 유도 그래프는 다음 범위다.
다른 언어 엔진을 같은 과제로 실행한 자료도 아직 없다.

[언어 비교](../../docs/ontology-language-comparison.ko.md) ·
[조건 기록 설계](../../docs/language/native-condition-evidence.ko.md)
