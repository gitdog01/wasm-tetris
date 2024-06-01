# wasm-tetris

- wasm 로 만든 테트리스 게임입니다.
- wasm 코드와 sever 코드가 나누어져 있습니다.

- 빌드하기
- `GOOS=js GOARCH=wasm go build -o ./main.wasm ./main.go`
- 실행하기
- `go run ./server.go`

## docker 로 빌드하기

- go 1.22 버전에서 빌드를 합니다.
- 최종실행은 Scratch 이미지에서 합니다.
- `docker build -t tetris .`
- `docker run -p 8080:8080 tetris`

## chat server 추가
