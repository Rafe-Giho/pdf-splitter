# macOS 작업 준비

이 프로젝트는 현재 Windows 기준으로 공개 릴리즈가 나가 있고, macOS는 개발 이어가기용 준비 단계입니다.

## 1. 맥북에서 저장소 받기

```bash
git clone https://github.com/Rafe-Giho/pdf-splitter.git
cd pdf-splitter
```

이미 받아둔 저장소가 있으면:

```bash
git pull origin master
```

## 2. 필수 도구 설치

### Xcode Command Line Tools

```bash
xcode-select --install
```

### Homebrew

설치되어 있지 않으면:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Go

```bash
brew install go
```

확인:

```bash
go version
```

### Java Runtime

PDF 썸네일 렌더링에 필요합니다.

```bash
brew install openjdk
```

설치 후 안내되는 방식대로 `PATH` 또는 `JAVA_HOME`을 잡아 주세요.

확인:

```bash
java -version
```

## 3. 의존성 준비

```bash
go mod download
```

이 프로젝트는 `tools/pdfbox/pdfbox-app-3.0.7.jar`를 저장소에 포함하고 있으므로 별도 다운로드는 필요 없습니다.

## 4. 로컬 실행

```bash
go run ./cmd/pdfsplitter
```

우선 확인할 것:

- 앱이 실행되는지
- 파일 선택이 되는지
- PDF 분석이 되는지
- 썸네일이 생성되는지
- 분할 실행이 되는지

## 5. macOS 빌드

```bash
chmod +x ./scripts/build-macos.sh
./scripts/build-macos.sh
```

기본 산출물:

- `dist/macos/pdf-splitter`

Apple Silicon이 기본값이고, Intel이면 아키텍처를 명시합니다.

```bash
./scripts/build-macos.sh amd64
```

## 6. 현재 상태에서 알아둘 점

- macOS는 아직 실기기 검증이 완료되지 않았습니다.
- `.app` 번들, 서명, 노타리제이션은 아직 없습니다.
- 공개 배포용보다는 개발 이어가기와 기능 확인 단계로 보는 게 맞습니다.

## 7. 맥북에서 바로 이어서 하면 좋은 작업 순서

1. `go run ./cmd/pdfsplitter`로 실행 확인
2. 실제 PDF로 분석/썸네일/분할 동작 확인
3. `./scripts/build-macos.sh`로 바이너리 빌드
4. macOS 전용 문제 정리
5. 이후 `.app` 번들링과 서명 작업 진행

## 8. 권장 점검 명령

```bash
go test ./...
```

빌드만 빠르게 보려면:

```bash
go build ./cmd/pdfsplitter
```

## 9. 맥북에서 막히기 쉬운 지점

- `java`가 잡히지 않으면 썸네일 렌더링이 실패할 수 있습니다.
- `xcode-select --install`이 안 되어 있으면 일부 빌드가 막힐 수 있습니다.
- macOS 파일 선택/창 동작은 Windows와 다를 수 있으니 직접 확인이 필요합니다.
