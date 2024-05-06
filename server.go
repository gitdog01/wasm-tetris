package main

import (
	"log"
	"net/http"
)

func main() {
	// 프로젝트 디렉토리에서 모든 파일을 서비스하도록 설정
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
