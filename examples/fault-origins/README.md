# 실패한 계산의 출처를 Gooo로 읽기

계산이 멈춰도 어디에서 어떤 값을 읽었는지는 남길 수 있습니다. 이 도구는
저장된 실행에서 실패한 연산을 받아, Gooo 코드로 그 형태와 출처를 설명합니다.

`describe.gooo`의 `Describe`가 다음을 결정합니다.

- 일반 식, 지역 변수 갱신, 레코드 필드 갱신을 구분합니다.
- 연결된 Gooo 오른쪽 식이 있으면 해당 식의 위치를 선택합니다.
- 나머지는 보관된 Go 표현식과 그 위치를 보여줍니다.

원래 연산은 `operation`에 계속 남습니다. `source_view`는 출력 위치가
`SELECTED_GOOO_RHS`인지 `GO_PROJECTION`인지 알려줍니다. 두 위치의 바이트
범위는 서로 다른 소스를 기준으로 합니다. 선택된 Gooo 소스는 패키지를
조립한 결과일 수 있으므로 원래 작업 파일 경로와도 구별합니다.

## 실행

저장된 파일은 앞선 지역 변수 나눗셈 관찰의 실행 기록입니다.

```sh
gzip -dc examples/fault-origins/recorded-runtime.json.gz > /tmp/gooo-fault-runtime.json
gooo package execute --origin-receipt /tmp/gooo-fault-runtime.json \
  --origin-kind faults \
  examples/fault-origins/gooo.workspace.json
```

두 결과가 나옵니다. 하나는 `amount /= divisor / 보정`의 바깥 갱신이 실패한
경우이고, 다른 하나는 안쪽 `divisor / 보정`이 먼저 실패한 경우입니다.
생산자 프로그램을 실행하지 않고 저장된 기록만 읽습니다. `Describe` 자체는
새로 실행됩니다. `--json`을 더하면 판독기의 실행 기록도 보관할 수 있습니다.

`--origin-root-activity-id`와 `--origin-activity-id`로 호출의 루트와 실제
연산이 선언된 activity를 각각 선택할 수 있습니다. `--origin-offset`과
`--origin-limit`은 선택된 행에 적용되고, 원래 입력 번호는 유지됩니다.

## 관찰 범위

호출·중첩 연산 사례 두 건과 필드·지역 변수 혼합 사례 두 건에서 Gooo 출력이
외부 비교값과 일치했습니다. 생산자 실행과 모델 호출은 0회입니다. 판독기에
정답 입력을 넣지 않았으므로 실행 기록의 유한 정답률 분모도 0으로 유지합니다.

현재 Gooo 위치 연결은 지역 변수 `/=`의 오른쪽 식을 다룹니다. 일반 나눗셈과
필드 갱신은 연산의 Go 위치를 유지합니다. 원본의 값·출처를 다시 실행해
검증하는 과정은 없으며, 입력 기록의 관계와 식별자가 일관된지를 확인합니다.
잘못된 연결·조회 조건 11종, 페이지 선택, 빈 선택 결과와 저장 파일의 동일성을
별도로 확인했습니다.

[저장 파일의 기원](recorded-origin.json)과
[판독·등록 기록](../../docs/observations/fault-origins/observations.json)에 범위를 남깁니다.
