# 아키텍처 가이드

CRDP File Converter의 기술 설계 및 구현 세부사항을 설명합니다.

## 목차

1. [시스템 아키텍처](#시스템-아키텍처)
2. [모듈 설계](#모듈-설계)
3. [데이터 흐름](#데이터-흐름)
4. [병렬 처리](#병렬-처리)
5. [API 클라이언트](#api-클라이언트)
6. [성능 고려사항](#성능-고려사항)

## 시스템 아키텍처

```
┌─────────────────────────────────────────────────────────┐
│                    CLI 인터페이스                          │
│              (cmd/main.go - Cobra)                       │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
   ┌────▼─────────┐      ┌──────▼──────────┐
   │ 단일 처리      │      │  병렬 처리       │
   │(Sequential)  │      │(Parallel)       │
   └────┬─────────┘      └──────┬──────────┘
        │                       │
        └───────────┬───────────┘
                    │
         ┌──────────▼───────────┐
         │  Converter Module     │
         │  (pkg/converter/)    │
         │                      │
         │ - readAndCollectData │
         │ - performBulkConv.   │
         │ - mergeOutputFiles   │
         └──────────┬───────────┘
                    │
         ┌──────────▼───────────┐
         │  CRDP API Client     │
         │  (pkg/crdp/)         │
         │                      │
         │ - ProtectBulk        │
         │ - RevealBulk         │
         │ - Extract*Response   │
         └──────────┬───────────┘
                    │
         ┌──────────▼───────────┐
         │  CRDP Server         │
         │  (HTTP API)          │
         └──────────────────────┘
```

## 모듈 설계

### 1. CLI Module (cmd/main.go)

**책임:**
- 명령줄 인자 파싱 (Cobra framework)
- 사용자 입력 검증
- 작업 흐름 조율

**주요 구성:**
```go
var (
    delimiter   string
    column      int
    encode      bool
    decode      bool
    output      string
    parallel    int
    // ... 기타 옵션
)
```

**함수:**
- `runConversion()`: 메인 처리 함수
- `validateOperationFlags()`: 플래그 검증
- `generateOutputPath()`: 출력 파일명 생성
- `promptSkipHeader()`: 대화형 입력

### 2. Converter Module (pkg/converter/converter.go)

**책임:**
- 파일 읽기/쓰기
- 데이터 변환 조율
- 배치 처리
- 병렬 처리 관리

**핵심 타입:**
```go
type DumpConverter struct {
    client *crdp.Client
    host   string
    port   int
    policy string
}
```

**주요 메서드:**

#### ProcessFile (단일 처리)
```
읽기 → 수집 → 변환 → 쓰기
```

#### ProcessFileParallel (병렬 처리)
```
분할 → [워커1, 워커2, ...] → 병합 → 정리
```

#### SplitInputFile
- 파일을 동등한 부분으로 분할
- 헤더 보존
- 부분 파일 생성

#### performBulkConversion
- 배치 단위로 API 호출
- 진행 상황 표시 (프로그래스 바)
- 결과 수집

### 3. CRDP Client Module (pkg/crdp/client.go)

**책임:**
- CRDP API HTTP 통신
- 요청/응답 변환
- 에러 처리

**핵심 타입:**
```go
type Client struct {
    baseURL  string
    policy   string
    timeout  time.Duration
    client   *http.Client
}

type APIResponse struct {
    StatusCode int
    Body       interface{}
    RequestURL string
    Error      error
}
```

**주요 메서드:**
- `ProtectBulk(dataList)`: 대량 데이터 암호화
- `RevealBulk(dataList)`: 대량 데이터 복호화
- `ExtractProtectedListFromProtectResponse()`: 응답 파싱
- `ExtractRestoredListFromRevealResponse()`: 응답 파싱

## 데이터 흐름

### 단일 처리 흐름

```
입력 파일 (CSV)
    ↓
readAndCollectData()
├─ 행 읽기
├─ 헤더 감지
└─ 변환 대상 데이터 수집
    ↓
dataToConvert: ["value1", "value2", ...]
    ↓
performBulkConversion()
├─ 배치 생성 (크기: 100)
├─ API 호출 (ProtectBulk/RevealBulk)
└─ 결과 수집
    ↓
convertedList: ["encrypted1", "encrypted2", ...]
    ↓
writeConvertedOutput()
├─ 행 타입 확인 (header, convert, skip, empty)
├─ 데이터 교체
└─ 파일 쓰기
    ↓
출력 파일 (CSV)
```

### 병렬 처리 흐름

```
입력 파일
    ↓
SplitInputFile() → [part1, part2, part3, ...]
    ↓
┌─ 워커1 ─┐
│ part1   │ → writeConvertedOutput() → out.part1
├─────────┤
├─ 워커2 ─┤
│ part2   │ → writeConvertedOutput() → out.part2
├─────────┤
├─ 워커3 ─┤
│ part3   │ → writeConvertedOutput() → out.part3
└─────────┘
    ↓
mergeOutputFiles([out.part1, out.part2, out.part3])
├─ 헤더 처리 (첫 파일만)
└─ 데이터 병합 (헤더 스킵)
    ↓
최종 출력 파일
```

## 병렬 처리

### 구현 방식

```go
// 고루틴 기반 병렬 처리
var wg sync.WaitGroup
errChan := make(chan error, len(splits))

for i, split := range splits {
    wg.Add(1)
    go func(idx int, splitFile SplitFileResult) {
        defer wg.Done()
        // 개별 처리
        err := dc.ProcessFile(...)
        if err != nil {
            errChan <- err
        }
    }(i, split)
}

wg.Wait()
```

### 동기화 메커니즘

- **WaitGroup**: 모든 고루틴 완료 대기
- **Channel**: 에러 수집 및 통신
- **Mutex** (필요시): 공유 리소스 보호

### 성능 특성

```
파일 크기별 성능:
- < 10MB: 병렬 처리 오버헤드 > 이득 (단일 처리 추천)
- 10-100MB: 2-4 워커 추천
- > 100MB: 4-8 워커 추천
```

## API 클라이언트

### HTTP 요청 구조

#### Protect (암호화)
```
POST /v1/protectbulk HTTP/1.1
Content-Type: application/json

{
    "protection_policy_name": "P03",
    "data_array": ["value1", "value2", ...]
}
```

#### Reveal (복호화)
```
POST /v1/revealbulk HTTP/1.1
Content-Type: application/json

{
    "protection_policy_name": "P03",
    "protected_data_array": [
        {"protected_data": "encrypted1"},
        {"protected_data": "encrypted2"}
    ]
}
```

### 응답 처리

```go
// 응답 예시
{
    "protected_data": ["token1", "token2"]
    // 또는
    "protected_data_array": [
        {"protected_data": "token1"},
        {"protected_data": "token2"}
    ]
    // 또는
    "data": ["decrypted1", "decrypted2"]
}
```

응답 파싱은 여러 가능한 키를 시도합니다:
- `protected_data` / `protected_data_array`
- `data` / `restored` / `plain`
- 등등

## 성능 고려사항

### 메모리 사용

```
단일 처리:
- 전체 파일 메모리 로드
- 배치당 메모리: batchSize × sizeof(string)

병렬 처리:
- 파일 분할 시 메모리 로드
- 각 워커: 독립적인 메모리 사용
- 병합: 순차 읽기 (메모리 효율적)
```

### 시간 복잡도

```
단일 처리:
- 읽기: O(n)
- API 호출: O(n/batchSize)
- 쓰기: O(n)
- 총합: O(n)

병렬 처리:
- 분할: O(n)
- 병렬 처리: O(n/workers)
- 병합: O(n)
- 총합: O(n)
```

### 네트워크 최적화

```
배치 크기 최적화:
- 너무 작음 (1-10): 오버헤드 증가
- 적절함 (50-200): 버퍼 사이즈와 균형
- 너무 큼 (>500): 메모리 사용 증가

권장값: 100 (기본)
```

## 에러 처리

### 에러 전파

```
CLI 검증 오류
    ↓ (즉시 종료)
파일 I/O 오류
    ↓ (즉시 종료)
API 호출 오류
    ↓ (즉시 종료, 배치 단위)
결과 불일치
    ↓ (즉시 종료)
```

### 복구 전략

```
- 부분 성공: 불가 (원자성 보장)
- 재시도: 구현 계획 중
- 로깅: 모든 에러 기록
```

## 확장 가능성

### 향후 개선 방안

1. **스트리밍 처리**
   - 대용량 파일용 메모리 효율화
   
2. **데이터 포맷 지원**
   - JSON, Parquet, Protocol Buffers 등

3. **비동기 작업**
   - 채널 기반 파이프라인

4. **캐싱**
   - 변환 결과 캐시

5. **메트릭 수집**
   - 처리 시간, 처리량 등

---

**최종 수정**: 2025-11-15
