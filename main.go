package main

import (
	"math/rand"
	"strconv"
	"syscall/js"
	"time"
)

const (
	width          = 10
	height         = 20
	size           = 32
	sidePanelWidth = 6 * size // 사이드 패널의 너비
)

var (
	canvas       js.Value
	ctx          js.Value
	board        [height][width]int
	currentPiece *Piece
	holdPiece    *Piece
	nextPieces   []Piece
	gameOver     bool
	score        int
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
			if !gameOver {
				movePieceDown()
				drawBoard()
				drawSidePanel()
				drawPiece(currentPiece)
			}
		}
	}()

	// JavaScript 이벤트 리스너 설정
	js.Global().Set("moveLeft", js.FuncOf(moveLeft))
	js.Global().Set("moveRight", js.FuncOf(moveRight))
	js.Global().Set("rotatePiece", js.FuncOf(rotatePiece))
	js.Global().Set("moveDown", js.FuncOf(moveDown))
	js.Global().Set("dropPiece", js.FuncOf(dropPiece))
	js.Global().Set("holdCurrentPiece", js.FuncOf(holdCurrentPiece)) // 홀드 기능 추가

	// JavaScript 이벤트 리스너
	done := make(chan struct{}, 0)
	<-done
}

func init() {
	nextPieces = make([]Piece, 5)
	for i := 0; i < 5; i++ {
		nextPieces[i] = pieces[rand.Intn(len(pieces))]
	}
}

func resetPiece() {
	currentPiece = &nextPieces[0]
	updateNextPieces()
	currentPiece.x = (width - currentPiece.width) / 2
	currentPiece.y = 0

	if checkCollision(currentPiece.x, currentPiece.y, currentPiece.shape) {
		gameOver = true
	}
}

func holdCurrentPiece(this js.Value, p []js.Value) interface{} {
	if holdPiece == nil {
		holdPiece = currentPiece
		resetPiece()
	} else {
		holdPiece, currentPiece = currentPiece, holdPiece
		currentPiece.x = (width - currentPiece.width) / 2
		currentPiece.y = 0
	}
	drawBoard()
	drawSidePanel()
	drawPiece(currentPiece)
	return nil
}

func updateNextPieces() {
	nextPieces = append(nextPieces[1:], pieces[rand.Intn(len(pieces))])
}

func movePieceDown() {
	if !checkCollision(currentPiece.x, currentPiece.y+1, currentPiece.shape) {
		currentPiece.y++
	} else {
		placePiece()
		resetPiece()
	}
}

func moveLeft(this js.Value, p []js.Value) interface{} {
	if !checkCollision(currentPiece.x-1, currentPiece.y, currentPiece.shape) {
		currentPiece.x--
	}
	drawBoard()
	drawSidePanel()
	drawPiece(currentPiece)
	return nil
}

func moveRight(this js.Value, p []js.Value) interface{} {
	if !checkCollision(currentPiece.x+1, currentPiece.y, currentPiece.shape) {
		currentPiece.x++
	}
	drawBoard()
	drawSidePanel()
	drawPiece(currentPiece)
	return nil
}

func moveDown(this js.Value, p []js.Value) interface{} {
	// 한 칸 아래로 이동
	if !checkCollision(currentPiece.x, currentPiece.y+1, currentPiece.shape) {
		currentPiece.y++
	} else {
		placePiece()
		resetPiece()
	}
	drawBoard()
	drawSidePanel()
	drawPiece(currentPiece)
	return nil
}

func rotatePiece(this js.Value, p []js.Value) interface{} {
	rotated := rotate(currentPiece.shape)
	if !checkCollision(currentPiece.x, currentPiece.y, rotated) {
		currentPiece.shape = rotated
	}
	drawBoard()
	drawSidePanel()
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
	// 현재 피스를 게임 보드에 배치
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if currentPiece.shape[i][j] == 1 && currentPiece.y+i >= 0 {
				board[currentPiece.y+i][currentPiece.x+j] = 1
			}
		}
	}
	// 가득 찬 줄을 제거
	clearFullRows()
}

func clearFullRows() {
	for y := 0; y < height; y++ {
		full := true
		for x := 0; x < width; x++ {
			if board[y][x] == 0 {
				full = false
				break
			}
		}
		if full {
			score += 100 // 줄 하나를 지울 때마다 100점 추가
			// 해당 줄을 제거하고, 모든 윗 줄을 아래로 이동
			for removeY := y; removeY > 0; removeY-- {
				for x := 0; x < width; x++ {
					board[removeY][x] = board[removeY-1][x]
				}
			}
			// 가장 위의 줄을 비워줍니다.
			for x := 0; x < width; x++ {
				board[0][x] = 0
			}
			// 같은 줄이 또 다시 가득 찼을 수 있으므로, 검사를 재실행
			y--
		}
	}
}

func drawBoard() {
	ctx.Call("clearRect", 0, 0, canvas.Get("width").Int()-sidePanelWidth, canvas.Get("height").Int())
	ctx.Set("fillStyle", "#f0f0f0")
	ctx.Call("fillRect", 0, 0, canvas.Get("width").Int()-sidePanelWidth, canvas.Get("height").Int())

	ctx.Set("strokeStyle", "#d3d3d3")
	ctx.Set("lineWidth", 1)
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			ctx.Call("strokeRect", j*size, i*size, size, size)
			if board[i][j] == 1 {
				ctx.Set("fillStyle", "black")
				ctx.Call("fillRect", j*size+1, i*size+1, size-2, size-2)
			}
		}
	}

	if gameOver {
		ctx.Set("font", "48px serif")
		ctx.Set("fillStyle", "red")
		ctx.Call("fillText", "Game Over", (canvas.Get("width").Int()-sidePanelWidth)/2-100, canvas.Get("height").Int()/2)
	}
}

func drawSidePanel() {
	ctx.Call("clearRect", canvas.Get("width").Int()-sidePanelWidth, 0, sidePanelWidth, canvas.Get("height").Int())
	ctx.Set("fillStyle", "#ffffff")
	ctx.Call("fillRect", canvas.Get("width").Int()-sidePanelWidth, 0, sidePanelWidth, canvas.Get("height").Int())

	ctx.Set("font", "24px serif")
	ctx.Set("fillStyle", "black")
	ctx.Call("fillText", "Score: "+strconv.Itoa(score), canvas.Get("width").Int()-sidePanelWidth+10, 30)

	drawNextPieces()
	drawHoldPiece()
}

func drawNextPieces() {
	for i, piece := range nextPieces {
		xOffset := canvas.Get("width").Int() - sidePanelWidth + 10
		yOffset := 50 + i*4*size
		if i == 0 {
			yOffset = 50 // 첫 번째 블록은 크게 표시
			drawMiniPiece(piece, xOffset, yOffset, size)
		} else {
			yOffset += size // 두 번째 블록부터는 크기를 줄임
			drawMiniPiece(piece, xOffset, yOffset, size/2)
		}
	}
}

func drawMiniPiece(piece Piece, xOffset, yOffset, blockSize int) {
	ctx.Set("fillStyle", "gray")
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if piece.shape[i][j] == 1 {
				ctx.Call("fillRect", xOffset+j*blockSize, yOffset+i*blockSize, blockSize, blockSize)
			}
		}
	}
}

func drawHoldPiece() {
	if holdPiece != nil {
		xOffset := canvas.Get("width").Int() - sidePanelWidth + 10
		yOffset := 250
		drawMiniPiece(*holdPiece, xOffset, yOffset, size)
	}
}

func drawPiece(piece *Piece) {
	ctx.Set("fillStyle", "red")
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if piece.shape[i][j] == 1 {
				ctx.Call("fillRect", (piece.x+j)*size+1, (piece.y+i)*size+1, size-2, size-2)
			}
		}
	}
}

func dropPiece(this js.Value, p []js.Value) interface{} {
	// 블록이 이동할 수 없을 때까지 반복적으로 내립니다.
	for checkCollision(currentPiece.x, currentPiece.y+1, currentPiece.shape) == false {
		currentPiece.y++
	}
	placePiece()
	resetPiece()
	drawBoard()
	drawSidePanel()
	drawPiece(currentPiece)
	return nil
}
