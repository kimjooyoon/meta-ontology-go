# 기준을 바꾸면 어디를 다시 확인해야 할까

요금 계산 함수 `Fee`는 조립할 때 `Reference`에 답을 묻고, `Main`은 `Fee`를
호출합니다. 두 소스의 차이는 `Reference`의 계산식 한 곳입니다.

```sh
gooo impact examples/source-change-impact/before.gooo.fixture \
  examples/source-change-impact/after.gooo.fixture
```

이 명령이 포함된 개발 버전 CLI에서 실행합니다. 결과는 변경된 선언 1개와
재확인 대상 3개(`Reference`, `Fee`, `Main`)를 보여줍니다. 같은 `Integer`를
쓰는 `Unrelated`는 영향을 연결할 근거가 없어 대상에 포함되지 않습니다.

`Fee`에는 선택용 사례 1개, 관찰 입력 1개, 별도 평가 사례 1개가 있습니다.
영향 보고서는 각 수를 그대로 보여줍니다. 이번 명령에서 함수를 실행하거나
모델에 질문하지는 않으므로 새 통과율은 나오지 않습니다. 다음 실행에서 무엇을
확인할지 정하는 데 쓰는 지도입니다.

[보고서 형식과 범위](../../docs/source-change-impact.md)를 참고하세요.

## 실제 조립 결과와 연결하기

개발 중인 명령으로 다음 두 입력을 실행했습니다. 원래 소스는 `Main(11) = 15`,
`Main(-7) = 33`을 모두 맞혔습니다. 기준 함수만 바꾼 소스는 `Fee`의 기존 선택
사례와 기준 함수의 답이 충돌한다는 진단으로 조립을 중단했습니다.
선언에 적힌 기대값은 그대로 유지됐습니다.

```sh
gooo body-compose --source examples/source-change-impact/before.gooo.fixture \
  --entry Main --cases examples/source-change-impact/evaluation.json \
  --out /tmp/gooo-impact-before

gooo body-compose --source examples/source-change-impact/after.gooo.fixture \
  --entry Main --cases examples/source-change-impact/evaluation.json \
  --out /tmp/gooo-impact-after
```

출력 경로는 새 폴더를 지정합니다. 두 번째 명령의 예상 결과는 원래 사례와
기준 함수의 충돌 오류입니다. 이 관측은 이 예제의 두 실행 사례와 한 변경에
해당합니다. 변경 영향 보고서는 재확인할 대상을 안내하고, 실제 결과는 별도 실행
기록으로 확인합니다.
