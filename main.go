package main

import (
	"syscall/js"
	"time"
)

const (
	width  = 10
	height = 20
	size   = 32
)

var (
	canvas       js.Value
	ctx          js.Value
	board        [height][width]int
	currentPiece *Piece
	pieces       = []Piece{
		// I
		{0, [4][4]int{
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 4, 1, 0, 0},
		// O
		{1, [4][4]int{
			{1, 1, 0, 0},
			{1, 1, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 2, 2, 0, 0},
		// T
		{2, [4][4]int{
			{0, 1, 0, 0},
			{1, 1, 1, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 3, 2, 0, 0},
		// S
		{3, [4][4]int{
			{0, 1, 1, 0},
			{1, 1, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 3, 2, 0, 0},
		// Z
		{4, [4][4]int{
			{1, 1, 0, 0},
			{0, 1, 1, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 3, 2, 0, 0},
		// L
		{5, [4][4]int{
			{1, 0, 0, 0},
			{1, 1, 1, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 3, 2, 0, 0},
		// J
		{6, [4][4]int{
			{0, 0, 1, 0},
			{1, 1, 1, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		}, 3, 2, 0, 0},
	}
)

type Piece struct {
	index  int
	shape  [4][4]int
	width  int
	height int
	x, y   int
}

func main() {
	// JavaScript로부터 캔버스 컨텍스트 가져오기
	canvas = js.Global().Get("document").Call("getElementById", "gameCanvas")
	ctx = canvas.Call("getContext", "2d")

	// 초기 블록 설정
	resetPiece()

	// 화면 갱신 주기 설정
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range ticker.C {
			movePieceDown()
			drawBoard()
			drawPiece(currentPiece)
		}
	}()

	// JavaScript 이벤트 리스너 설정
	js.Global().Set("moveLeft", js.FuncOf(moveLeft))
	js.Global().Set("moveRight", js.FuncOf(moveRight))
	js.Global().Set("rotatePiece", js.FuncOf(rotatePiece))

	// JavaScript 이벤트 리스너
	done := make(chan struct{}, 0)
	<-done
}

func resetPiece() {
	// 무작위로 블록 선택
	currentPiece = &pieces[time.Now().UnixNano()%int64(len(pieces))]
	currentPiece.x = (width - currentPiece.width) / 2
	currentPiece.y = 0
}

func movePieceDown() {
	if !checkCollision(currentPiece.x, currentPiece.y+1, currentPiece.shape) {
		currentPiece.y++
	} else {
		placePiece()
		resetPiece()
	}
}

func moveLeft(js.Value, []js.Value) interface{} {
	if !checkCollision(currentPiece.x-1, currentPiece.y, currentPiece.shape) {
		currentPiece.x--
	}
	drawBoard()
	drawPiece(currentPiece)
	return nil
}

func moveRight(js.Value, []js.Value) interface{} {
	if !checkCollision(currentPiece.x+1, currentPiece.y, currentPiece.shape) {
		currentPiece.x++
	}
	drawBoard()
	drawPiece(currentPiece)
	return nil
}

func rotatePiece(js.Value, []js.Value) interface{} {
	rotated := rotate(currentPiece.shape)
	if !checkCollision(currentPiece.x, currentPiece.y, rotated) {
		currentPiece.shape = rotated
	}
	drawBoard()
	drawPiece(currentPiece)
	return nil
}

func checkCollision(x, y int, shape [4][4]int) bool {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if shape[i][j] == 1 {
				if x+j < 0 || x+j >= width || y+i >= height {
					return true
				}
				if y+i >= 0 && board[y+i][x+j] == 1 {
					return true
				}
			}
		}
	}
	return false
}

func rotate(shape [4][4]int) [4][4]int {
	var rotated [4][4]int
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			rotated[j][4-i-1] = shape[i][j]
		}
	}
	return rotated
}

func placePiece() {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if currentPiece.shape[i][j] == 1 && currentPiece.y+i >= 0 {
				board[currentPiece.y+i][currentPiece.x+j] = 1
			}
		}
	}
	clearFullRows()
}

func clearFullRows() {
	for i := 0; i < height; i++ {
		full := true
		for j := 0; j < width; j++ {
			if board[i][j] == 0 {
				full = false
				break
			}
		}
		if full {
			for k := i; k > 0; k-- {
				for j := 0; j < width; j++ {
					board[k][j] = board[k-1][j]
				}
			}
			for j := 0; j < width; j++ {
				board[0][j] = 0
			}
		}
	}
}

func drawBoard() {
	ctx.Call("clearRect", 0, 0, canvas.Get("width").Int(), canvas.Get("height").Int())
	ctx.Set("fillStyle", "black")
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if board[i][j] == 1 {
				ctx.Call("fillRect", j*size, i*size, size, size)
			}
		}
	}
}

func drawPiece(piece *Piece) {
	ctx.Set("fillStyle", "red")
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if piece.shape[i][j] == 1 {
				ctx.Call("fillRect", (piece.x+j)*size, (piece.y+i)*size, size, size)
			}
		}
	}
}
