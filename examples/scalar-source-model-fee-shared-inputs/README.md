# 고정 요금 조건과 분기 선택

업무 조건은 다음과 같다.

- 음수 입력의 요금은 0이다.
- 나머지는 입력의 두 배로 계산하되 상한은 90이다.
- 요금은 음수가 될 수 없다.

[Gooo 소스](main.gooo.fixture)는 이 조건을 `RequiredFee`와 `Nonnegative` 활동으로
고정하고, `Fee`의 두 분기 배치를 허용된 선택으로 선언한다.
원래 본문은 상한 분기의 반환값이 뒤바뀐 상태다.
모델은 선언한 두 선택의 네 조합을 정렬하며 조건식은 후보 결과를 평가한다.

## 조건 입력은 함께 선언한다

`criteria`의 입력은 블록 안의 모든 조건식이 공유한다.
아래 두 조건식과 입력 네 개는 **조건식·입력 쌍 8개**를 평가한다.

```gooo
criteria {
 predicate "nonnegative" activity "Nonnegative"
 predicate "required" activity "RequiredFee"
 input "[-3]" input "[0]" input "[21]" input "[99]"
 holdout_input "[-7]" holdout_input "[30]" holdout_input "[45]" holdout_input "[46]"
}
```

조건식마다 같은 입력을 반복하면 중복 튜플 오류가 난다.
앞서 작성한 별도 소스는 이 오류로 파싱 단계에서 멈췄고 모델 호출은 0회였다.
위 예제는 다른 ID·상수·입력으로 새 관찰을 수행했다. 원래 실패를 재실행하지 않았다.

## 관찰 결과

[관찰 기록](observation.json)은 공개 컴파일러 소스와 새 테스트 파일 하나에 연결된다.
합성 모델이 첫 번째로 제안한 mask 3은 선택 조건 **4/8**을 만족했다.
예산 2 안에서 다음 mask 1을 선택해 조건 **8/8**을 만족했다.
별도 입력 네 개의 두 조건식도 **8/8**을 만족했다. 여기에는 0 입력과
상한 분기 경계 45·46이 포함돼 있다.

모델 호출은 **1회**, 시도는 **2/2**, 상태는 `COMPLETE_FINITE`였다.
관찰한 입력 밖의 모든 정수에 대한 증명, 학습 가중치의 정답률, 전체 메모리와
CPU 사용률은 추가 평가가 필요하다. 학습은 0회였고 생성된 Gooo 네이티브 프로그램은
실행하지 않았다. 실제 조건은 기존 유한 소스 평가 흐름에서 확인했다.

## 결정론적 구성

이 변경을 포함한 개발 소스에서 모델 없이 구성한다.

```sh
go run ./cmd/gooo body-codegen --json --activity Fee \
  examples/scalar-source-model-fee-shared-inputs/main.gooo.fixture
```

호환하는 수치 결정 아티팩트를 사용하려면 `--path-model /path/to/metadata.json`을
추가한다. 모델 파일은 이 예제에 포함하지 않는다.
입력 역할과 저장 기록은 [모델 연결 안내](../../docs/scalar-source-model.ko.md)를 참고한다.
