# Changelog

모든 주목할 만한 변경사항은 이 파일에 기록됩니다.

[Semantic Versioning](https://semver.org/) 형식을 따릅니다.

## [Unreleased]

## [1.0.0] - 2025-11-15

### Added

#### 핵심 기능
- CSV/TSV 파일의 특정 컬럼 암호화(protect) 및 복호화(reveal)
- CRDP API를 통한 데이터 변환
- 대량 데이터 배치 처리 (기본: 100개/배치)
- 병렬 처리 지원 (configurable workers)
- 자동 헤더 감지 및 건너뛰기 옵션
- 중복 파일명 자동 처리 (번호 증가: e01_, e02_ 등)

#### CLI 옵션
- `--encode` / `-e`: 데이터 암호화
- `--decode` / `-d`: 데이터 복호화
- `--column` / `-c`: 변환 대상 컬럼 인덱스 (0-based)
- `--delimiter`: 컬럼 구분자 (기본: `,`)
- `--skip-header` / `-s`: 헤더 라인 건너뛰기
- `--output`: 출력 파일 경로
- `--batch-size`: API 배치 크기
- `--parallel` / `-p`: 병렬 처리 워커 수
- `--host`: CRDP 서버 호스트
- `--port`: CRDP 서버 포트
- `--policy`: 데이터 보호 정책
- `--timeout`: API 요청 타임아웃

#### 빌드 및 배포
- Makefile을 통한 자동화된 빌드
- 크로스플랫폼 빌드 지원 (Linux, macOS, Windows)
- 자동 테스트 및 커버리지 보고

#### 문서
- 상세한 README.md
- 아키텍처 설명 (docs/ARCHITECTURE.md)
- 기여 가이드 (CONTRIBUTING.md)
- 변경 이력 (CHANGELOG.md)

#### 개발 및 테스트
- 유닛 테스트 (pkg/crdp, pkg/converter)
- 테스트 커버리지 리포트
- CI/CD 파이프라인 (GitHub Actions)

### Changed

### Deprecated

### Removed

### Fixed

### Security

---

## 버전 형식

### Major (X.0.0)
- API 호환성이 깨지는 변경
- 주요 기능 추가

### Minor (1.X.0)
- 하위 호환 가능한 기능 추가
- 성능 개선

### Patch (1.0.X)
- 버그 수정
- 문서 개선

---

## 미래 계획

### v1.1.0 계획
- [ ] 직렬화된 데이터 형식 지원 (JSON, XML)
- [ ] 추가 구분자 자동 감지
- [ ] 진행 상황 실시간 모니터링
- [ ] 부분 실패 복구 기능

### v2.0.0 계획
- [ ] 웹 API 서버 추가
- [ ] GUI 클라이언트
- [ ] 클라우드 스토리지 통합 (S3, GCS)
- [ ] 실시간 스트리밍 처리

---

## 기여자

- [@sjrhee](https://github.com/sjrhee) - 프로젝트 시작자

---

## 라이선스

이 프로젝트는 [MIT License](LICENSE) 하에 배포됩니다.
