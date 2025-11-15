# CRDP File Converter (Go)

CRDP (CipherTrust RESTful Data Protection) 서비스를 사용하여 텍스트 파일의 특정 컬럼 데이터를 암호화(protect) 또는 복호화(reveal)하는 도구입니다.

> 참고: 모든 프로젝트에서 동일한 CRDP 서버 설정을 사용합니다.

## 기능

- CSV, TSV 등 구분자 기반 텍스트 파일 지원
- 특정 컬럼 선택적 변환
- CRDP API를 통한 데이터 보호/복호화
- 대량 데이터 처리를 위한 배치 처리 지원
- **병렬 처리**: 입력 파일을 여러 부분으로 나누어 동시 처리 가능
- 변환 결과 자동 저장 (실행 폴더에 `e{nn}_` 또는 `d{nn}_` 접두사)
- 중복 파일명 자동 처리 (번호 증가)
- 헤더 라인 건너뛰기 지원
- 오류 처리 및 진행 상황 표시

## 설치

### 사전 요구사항

- Go 1.21 이상

### 의존성 설치

```bash
go mod download
```

### 빌드

```bash
# 현재 OS용 빌드
go build -o crdp-file-converter ./cmd
```

또는 Makefile 사용:

```bash
# 현재 OS용 빌드
make build

# 크로스플랫폼 빌드 (Linux, Darwin/macOS, Windows)
make build-cross
# 생성 위치: bin/ 디렉토리
#   - bin/crdp-file-converter-linux-amd64
#   - bin/crdp-file-converter-linux-arm64
#   - bin/crdp-file-converter-darwin-amd64
#   - bin/crdp-file-converter-darwin-arm64
#   - bin/crdp-file-converter-windows-amd64.exe
```

또는 개발 모드에서 직접 실행:

```bash
go run ./cmd main.go <arguments>
```

## 사용법

### 기본 사용법

```bash
# CSV 파일의 2번째 컬럼(인덱스 1)을 암호화
./crdp-file-converter data.csv --column 1 --encode

# TSV 파일의 3번째 컬럼을 복호화 (헤더 건너뛰기)
./crdp-file-converter data.tsv --delimiter '\t' --column 2 --decode --skip-header

# 4개 워커를 사용한 병렬 처리
./crdp-file-converter large_data.csv --column 1 --encode --parallel 4

# 특정 파일명으로 출력 (자동 파일명 생성 무시)
./crdp-file-converter data.csv --column 1 --encode --output result.csv
```

### 명령줄 옵션

기본값을 포함한 모든 옵션을 확인하려면:

```bash
./crdp-file-converter --help
```

다음과 같이 표시됩니다:

```
CRDP Dump File Converter

Converts CSV/TSV files by encoding/decoding specific columns using CRDP API.

Example:
  crdp-file-converter data.csv --column 1 --encode
  crdp-file-converter data.tsv --delimiter '\t' --column 2 --decode --skip-header

Usage:
  crdp-file-converter <input_file> [flags]

Flags:
  -e, --encode             Encode (protect) data
  -d, --decode             Decode (reveal) data
  -c, --column int         Column index to convert (0-based, required) (default -1)
      --delimiter string   Column delimiter (default ",")
  -s, --skip-header        Skip header line
      --output string      Output file path (default: {e/d}{nn}_{filename}.{ext})
      --batch-size int     Bulk API batch size (default 100)
      --host string        CRDP host (default "192.168.0.231")
      --port int           CRDP port (default 32082)
      --policy string      Protection policy (default "P03")
      --timeout int        Request timeout in seconds (default 5)
  -h, --help               help for crdp-file-converter
```

**주요 옵션:**

- `input_file`: 변환할 입력 파일 경로 (필수)
- `--encode` / `-e`: 데이터 암호화 (protect)
- `--decode` / `-d`: 데이터 복호화 (reveal)
  > ℹ️ `--encode`과 `--decode` 중 **반드시 하나는 지정**해야 합니다.
- `--column` / `-c`: 변환할 컬럼 인덱스 (0부터 시작, 필수)
  > ℹ️ 기본값은 `-1` (미지정 상태를 나타냄. 반드시 명시적으로 지정해야 함)
- `--delimiter`: 컬럼 구분자 (기본값: `,`)
- `--skip-header` / `-s`: 헤더 라인 건너뛰기 플래그
- `--output`: 출력 파일 경로
  - 미지정 시: 자동 생성 형식 `e{nn}_파일명` 또는 `d{nn}_파일명` (중복 시 번호 증가)
  - 지정 시: 해당 이름의 파일을 생성 (자동 생성 로직 무시)
- `--batch-size`: CRDP API 배치 크기 (기본값: 100)
- `--parallel` / `-p`: 병렬 처리 워커 수 (기본값: 1 = 순차 처리)
- `--host`: CRDP 서버 호스트 (기본값: `192.168.0.231`)
- `--port`: CRDP 서버 포트 (기본값: `32082`)
- `--policy`: 데이터 보호 정책 (기본값: `P03`)
- `--timeout`: API 요청 타임아웃 (기본값: 5초)

### 특정 CRDP 서버와 정책 사용

```bash
./crdp-file-converter data.csv --column 1 --encode \
    --host 192.168.0.231 --port 32082 --policy P03
```

### 짧은 플래그 사용 예

```bash
# 짧은 플래그로 암호화 (-e는 encode, -c는 column, -s는 skip-header)
./crdp-file-converter data.csv -c 1 -e -s

# 짧은 플래그로 복호화 (-d는 decode)
./crdp-file-converter data.csv -c 1 -d -s
```

## 오류 처리

프로그램은 실행 중 오류 발생 시 **즉시 진행을 멈추고** 오류 메시지를 표시한 후 **종료 코드 1**로 종료합니다.

### 처리되는 오류 유형

#### 1. 입력 파일 오류
```bash
$ ./crdp-file-converter nonexistent.csv --column 1 --operation protect
2025/11/15 01:18:38 CRDP Server: 192.168.0.231:32082
2025/11/15 01:18:38 Policy: P03
2025/11/15 01:18:38 ❌ Error: input file not found: nonexistent.csv
```

#### 2. 잘못된 작업 유형
```bash
$ ./crdp-file-converter data.csv --column 1 --operation invalid
2025/11/15 01:18:38 ❌ Error: operation must be 'protect' or 'reveal'
```

#### 3. CRDP API 호출 실패
```bash
# CRDP 서버에 연결 불가능한 경우
2025/11/15 01:18:38 [/v1/protectbulk] Status: 0, Response: <nil>
2025/11/15 01:18:38 ❌ Error: batch 1 API call failed: 0 - <nil>
```

#### 4. 출력 디렉토리 생성 실패
```bash
# 출력 디렉토리 권한 문제
2025/11/15 01:18:38 ❌ Error: failed to create output directory: permission denied
```

### 경고 메시지 (진행 계속)

일부 상황에서는 **경고 메시지**만 표시하고 프로그램을 계속 진행합니다:

```bash
# 존재하지 않는 컬럼 인덱스
2025/11/15 01:18:50 Warning: line 1 does not have column 10 (total columns: 3)
2025/11/15 01:18:50 No data to convert.
2025/11/15 01:18:50 ✅ Conversion completed: out/sample_data_converted.csv

# API 응답에서 결과 개수 불일치 (부분적으로 처리)
2025/11/15 01:13:28 Warning: batch 1 result count mismatch: requested 10, got 10 (continuing with what we have)
```

### 종료 코드

- **0**: 성공적으로 완료
- **1**: 오류 발생으로 진행 중단

## 병렬 처리

대용량 파일의 경우 `--parallel` 옵션으로 처리 속도를 높일 수 있습니다:

```bash
# 4개 워커로 병렬 처리 (파일을 4등분하여 동시 처리)
./crdp-file-converter large_data.csv --column 1 --encode --parallel 4

# 8개 워커로 병렬 처리
./crdp-file-converter huge_data.csv --column 1 --decode --parallel 8
```

**동작 방식:**
1. 입력 파일을 지정된 개수로 나눔 (헤더 포함 유지)
2. 각 부분을 독립적인 워커에서 동시 처리
3. 처리된 부분들을 순서대로 병합하여 최종 결과 생성
4. 임시 파일 정리

**병렬 처리 팁:**
- 초기 권장값: CPU 코어 수 또는 가용 메모리 고려
- 작은 파일(<10MB)은 병렬 처리 오버헤드가 더 클 수 있음
- 대용량 파일(>100MB)에서 성능 향상 효과 높음

## 예시

### 1. CSV 파일 암호화

**입력 파일** (`data.csv`):
```
이름,주민번호,주소
홍길동,1234567890123,서울시
김철수,9876543210987,부산시
```

**명령**:
```bash
./crdp-file-converter data.csv --column 1 --encode --skip-header
```

**출력 파일** (`e01_data.csv`):
```
이름,주민번호,주소
홍길동,8555545382975,서울시
김철수,1234567890123,부산시
```

### 2. TSV 파일 복호화

```bash
./crdp-file-converter encrypted_data.tsv --delimiter '\t' --column 2 --decode
```

## 프로젝트 구조

```
crdp-file-converter/
├── cmd/
│   └── main.go              # CLI 엔트리 포인트 (Cobra 기반)
├── pkg/
│   ├── crdp/
│   │   ├── client.go        # CRDP API 클라이언트 (HTTP 통신)
│   │   └── client_test.go   # 클라이언트 유닛 테스트
│   └── converter/
│       ├── converter.go     # 파일 변환 로직 (배치 처리)
│       └── converter_test.go # 변환기 유닛 테스트
├── tests/                   # 통합 테스트 디렉토리
├── out/                     # 출력 파일 디렉토리
├── go.mod                   # Go 모듈 정의
├── go.sum                   # 의존성 해시
├── Makefile                 # 빌드 자동화 스크립트
├── .gitignore               # Git 무시 파일
├── sample_data.csv          # 샘플 입력 데이터
├── sample_data_restored.csv # 샘플 복호화 데이터
├── test_restored.csv        # 테스트 복호화 파일
└── README.md               # 이 파일
```

## 개발

### Makefile을 사용한 빌드/테스트

```bash
# 도움말 확인
make help

# 프로젝트 빌드
make build

# 테스트 실행
make test

# 커버리지 포함 테스트
make test-cov

# 의존성 설치
make install

# 코드 포맷팅
make fmt

# 코드 린트 검사
make lint

# 클린업
make clean

# 샘플 데이터로 실행
make run
```

### 수동 테스트

```bash
# 모든 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./pkg/crdp
go test ./pkg/converter

# 커버리지 포함
go test -cover ./...
```

### 코드 포맷팅

```bash
# 코드 포맷팅 (gofmt 사용)
gofmt -w .

# 또는 goimports를 사용하여 import도 정렬
goimports -w .
```

### Lint 검사

```bash
# golangci-lint 설치 (처음 한 번만)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 실행
golangci-lint run ./...
```

## 아키텍처

### CRDP Client (`pkg/crdp/client.go`)

- **역할**: CRDP API와의 통신 담당
- **주요 메서드**:
  - `ProtectBulk(dataList)`: 대량 데이터 보호
  - `RevealBulk(protectedDataList)`: 대량 데이터 복호화
  - `Extract*FromResponse()`: 응답에서 결과 추출

### Converter (`pkg/converter/converter.go`)

- **역할**: 파일 처리 및 변환 로직 담당
- **주요 메서드**:
  - `ProcessFile()`: 파일 전체 처리 (읽기 → 변환 → 저장)
  - `ProcessFileParallel()`: 병렬 처리로 파일 변환
  - `SplitInputFile()`: 입력 파일을 여러 부분으로 분할
  - `readAndCollectData()`: CSV/TSV 파일 읽기 및 데이터 수집
  - `performBulkConversion()`: 배치 단위 대량 변환 처리
  - `mergeOutputFiles()`: 분할 처리된 파일 병합
  - `writeConvertedOutput()`: 결과를 파일로 저장

## 오류 처리 및 복원력

### 오류 핸들링 전략

1. **치명적 오류**: 즉시 진행 중단 (종료 코드 1)
   - 입력 파일 없음
   - 잘못된 작업 유형
   - 출력 디렉토리 생성 실패
   - CRDP API 호출 실패 (HTTP 상태 코드 >= 300)
   - 파일 읽기/쓰기 오류

2. **경고 메시지**: 진행 계속
   - 존재하지 않는 컬럼 인덱스 (해당 라인 건너뜀)
   - 빈 데이터 (원본 유지)
   - 일부 결과 부족 (부분 처리)

3. **로깅**: 모든 주요 단계와 오류 상황을 상세히 기록
   - 파일 처리 시작/완료
   - 배치 진행 상황
   - API 응답 상태
   - 처리 결과 통계

### 스크립트에서의 오류 처리

종료 코드를 이용한 자동화:

```bash
#!/bin/bash

# 단일 파일 처리
./crdp-file-converter data.csv --column 1 --operation protect
if [ $? -ne 0 ]; then
    echo "Conversion failed!"
    exit 1
fi

# 배치 처리
for file in *.csv; do
    ./crdp-file-converter "$file" --column 1 --operation protect
    if [ $? -ne 0 ]; then
        echo "Failed to process $file"
        continue  # 다음 파일로 계속
    fi
done
```

## 라이선스

이 프로젝트는 MIT 라이선스를 따릅니다.

## 샘플 데이터

프로젝트에 포함된 샘플 파일:

- **sample_data.csv**: 작은 테스트 파일 (10개 데이터 행 + 헤더)
- **sample_data_large.csv**: 대규모 테스트 파일 (30,000개 데이터 행 + 헤더)

### 샘플 파일로 테스트

```bash
# 작은 파일 테스트
./crdp-file-converter sample_data.csv -c 1 -e

# 대규모 파일 테스트 (성능 측정용)
./crdp-file-converter sample_data_large.csv -c 1 -e
```
