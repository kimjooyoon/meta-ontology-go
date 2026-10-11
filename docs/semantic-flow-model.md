# 같은 기능의 표현을 같은 결정 입력으로 읽기

개발 소스는 SDK0.2.35를 사용하며 `source_intent_condition_output_semantic_flow_v5` 모델을
직접 읽습니다. 모델 파일의 입력 버전에 맞춰 실제 조립과 `body-context` 사전 조회가
같은 배열을 만듭니다. 현재 설치된 공개 컴파일러0.6.25에는 아직 포함되지 않습니다.

후속 [관계 입력과 음수 신호 모델](relational-flow-model.md)도 명시적인 모델 형식으로
연결했습니다. 아래 v5 학습·CLI 측정은 당시 가중치와 생산자 기록을 그대로 유지합니다.

하나의 분기에서 비교 피연산자와 반환값을 추적할 수 있으면 직접 반환·별칭·대입·복사
표현을 정리합니다. `semantic_flow_context`에는 적용 여부와 이유, 정확한 값 흐름이
들어갑니다. 중첩 분기나 확정하지 못한 값은 원래384칸 입력을 유지합니다.
`source_features_sha256`는 이 버전에서 실제 사용한 소스 필드의 해시이며,
`input_sha256`는 전체 판단 배열의 해시입니다. 원래 Gooo 소스의 해시는 별도로 유지합니다.

```sh
gooo body-context --activity Choose --model model-v5.json \
  --feature-version source_intent_condition_output_semantic_flow_v5 \
  examples/body-codegen/source-semantic-flow-copy.gooo.fixture

gooo body-codegen --json --activity Choose --path-model model-v5.json \
  --path-step-attempts 1 \
  examples/body-codegen/source-semantic-flow-copy.gooo.fixture
```

모델은 소스가 허용한 후보의 순서를 정합니다. Gooo는 선택한 본문을 바로 조립해
선언된 출력·중간 조건을 검사합니다. 모델이 없으면 고정 순서로 이어가며, 저장한
프로그램은 모델 파일 없이도 재실행합니다. `--path-feedback-rounds`를 주면 실제 실패를
다음 판단에 전달합니다. 초기 입력에는 아직 실행하지 않은 후보의 실패가 들어가지 않습니다.

[공개 SDK34](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.34-experimental)의
v5 가중치는16개 직접 반환 소스로 학습했습니다. 같은 과제의 다른 표현에서 첫 선택40/64,
학습 소스10/16으로, v4 대비 개선과 후퇴가 함께 있었습니다. 후보 분포는 정답률 보장이
아니므로 출력 사례의 완전성과 첫 선택 성공률을 구분해야 합니다.
[학습 원본과 한계](https://github.com/kimjooyoon/gooo-decision-runtime/tree/ed84d95c43d1a68bf88bd9b8543139ed23ea86df/studies/semantic-flow-learning-20261010).

[고정한 CLI 계획](research/semantic-flow-cli-20261010/protocol.txt)은 이미 알려진 실패
과제의 두 표현을 실제 컴파일러에서 비교합니다. 이 문서의 API 연결 검사와 별도로
실제 학습 가중치의 성공·실패·실행 비용을 기록합니다.
[실제 결과](research/semantic-flow-cli-20261010/README.md)에서는 두 표현이 같은 배열과
후보 순서를 사용했지만 기본1회 대비 모델2회로 시도가 늘었습니다. 소스 사례와 모델을
지운 뒤의 저장 재실행은 모두 통과했습니다. 첫 선택 품질과 최종 완전성을 함께 기록합니다.
