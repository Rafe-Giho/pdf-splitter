# 프로젝트 작업 문서

## 프로젝트명

PDF Splitter

## 수행 목표

- 실제 페이지를 보면서 안전하게 PDF를 분할한다.
- Windows와 macOS에서 같은 코드베이스로 이어서 개발한다.
- GUI 중심으로 빠르게 사용할 수 있게 한다.
- 기능 확장과 보안 하드닝이 쉬운 구조를 유지한다.

## 요구사항

- GUI 제공
- Windows/macOS 실행 파일 생성 가능 구조
- 페이지 미리보기 기반 분할
- 가벼운 구조와 유지보수성
- 보안 취약점 완화

## 현재 결정 사항

- 언어: Go
- GUI: Gio
- PDF 분석/분할: pdfcpu
- PDF 썸네일 렌더링: Apache PDFBox
- 구조: `cmd`, `internal/ui`, `internal/pdf`, `internal/platform`
- 출력 기본 경로: 원본 PDF와 같은 폴더
- 테마: 시스템 / 라이트 / 다크

## 브랜치 전략

- `master`: Windows 기준 안정 브랜치
- `codex/macos-bringup`: macOS 실행/검증/초기 빌드 안정화 브랜치

macOS 작업은 당분간 `codex/macos-bringup`에서 진행하고, 확인된 내용만 `master`로 반영합니다.

## 현재 작업 브랜치 목표

브랜치: `codex/macos-bringup`

목표:

- macOS에서 앱이 실제로 실행되는지 검증
- 파일 선택, PDF 분석, 썸네일, 분할 저장의 macOS 동작 확인
- `.app` 번들링 전 단계까지 필요한 수정 정리
- 이후 서명/노타리제이션 준비를 위한 기반 확보

## 작업 목록

### 완료

- Git 저장소 초기화
- Go 모듈 초기화
- Windows/macOS 빌드 스크립트 구성
- 커스텀 창 및 아이콘 적용
- 시스템/라이트/다크 테마 적용
- PDF 선택 및 경로 직접 입력
- Windows 드래그앤드롭 지원
- 페이지 수 분석
- 실제 PDF 썸네일 렌더링
- 리스트형 / 격자형 보기
- 페이지 확대 보기
- 페이지 사이 구분선 클릭 기반 분할
- 빠른 분할 단위 토글/정수 입력/추천 드롭다운
- 원본 폴더 차번 파일명 저장
- 분할 완료 후 저장 위치 열기
- 입력 경로/파일/헤더/크기/페이지 수 검증
- 암호화 PDF / 첨부파일 / 위험 액션 차단
- 분할 저장 원자성 보강
- 썸네일 캐시 정리 정책
- 보안 점검 스크립트 추가
- Windows 첫 릴리즈 패키징 스크립트 추가

### 진행 예정

- macOS 파일 선택 동작 검증 및 수정
- macOS PDFBox 경로/Java 실행 검증
- macOS 드래그앤드롭 지원
- 사용자 설정 저장
- 최근 작업 파일 목록
- macOS 실기기 검증
- 코드서명/배포 게시
- 취약점 스캔 자동화 환경 정리

## To-do List

- [x] 프로젝트 초기화
- [x] 빌드 가능한 앱 구조 만들기
- [x] PDF 분석 서비스 구현
- [x] PDF 분할 서비스 구현
- [x] 실제 PDF 썸네일 렌더링
- [x] GUI 입력/상태 표시 구현
- [x] 플랫폼별 파일 선택 처리
- [x] Windows 드래그앤드롭 처리
- [x] 페이지 확대 보기 구현
- [x] 리스트/격자 전환 구현
- [x] 구분선 클릭 기반 분할 처리
- [x] 빠른 분할 단위 UI 개선
- [x] 원본 폴더 차번 파일명 처리
- [x] 분할 완료 후 저장 위치 열기
- [x] 보안 하드닝 적용
- [x] 보안 점검 스크립트 추가
- [x] 첫 Windows 릴리즈 패키징 스크립트 추가
- [x] macOS 개발 인수인계 문서 추가
- [ ] macOS 실기기 빌드 검증
- [ ] macOS 파일 선택 동작 검증
- [ ] macOS 썸네일 렌더링 검증
- [ ] macOS 분할 저장 검증
- [x] GitHub 원격 저장소 연결
- [x] 첫 GitHub 릴리즈 게시
- [ ] govulncheck 환경 포함 자동 스캔
- [ ] macOS `.app` 번들링
- [ ] macOS 서명 및 노타리제이션

## 기술 스택

- Go 1.26+
- Gio
- pdfcpu
- Apache PDFBox 3.0.7
- PowerShell / osascript 보조 대화상자

## 개발 메모

- Windows 빌드는 `scripts/build.ps1` 기준 `windowsgui`로 만든다.
- 썸네일은 `tools/pdfbox/pdfbox-app-3.0.7.jar` 기준으로 렌더링한다.
- 위험한 PDF 기능은 분석 단계에서 먼저 차단한다.
- PDF 관련 기능은 `internal/pdf`에 집중시켜 이후 병합/회전/추출 기능을 같은 계층에 추가한다.
- macOS 작업 시작 전 `AGENTS.md`, `MACOS_SETUP.md`, `CONTRIBUTING.md`를 먼저 확인한다.
