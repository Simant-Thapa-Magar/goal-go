package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

type GameObject struct {
	row, col, height, width, rowVelocity, columnVelocity int
	symbol                                               rune
}

type Coordinate struct {
	x, y int
}

var Screen tcell.Screen
var player1, player2, ball *GameObject
var gameReadyStatus, gamePauseStatus bool

var gameObjects []*GameObject
var coordinatesToClear []Coordinate
var borderToRegenerate *Coordinate
var screenWidth, screenHeight int
var player1Points, player2Points int
var roundWinner string

const FRAME_HEIGHT = 14
const FRAME_WIDTH = 80
const FRAME_BORDER_THICKNESS = 1
const FRAME_BORDER_VERTICAL = '║'
const FRAME_BORDER_HORIZONTAL = '═'
const FRAME_BORDER_TOP_LEFT = '╔'
const FRAME_BORDER_TOP_RIGHT = '╗'
const FRAME_BORDER_BOTTOM_RIGHT = '╝'
const FRAME_BORDER_BOTTOM_LEFT = '╚'
const PADDLE_HEIGHT = 4
const PADDLE_WIDTH = 1
const PADDLE_SYMBOL = 0x2588
const BALL_SYMBOL = 0x25CF
const BALL_HEIGHT = 1
const BALL_WIDTH = 1
const INITIAL_BALL_ROW_VELOCITY = 1
const INITIAL_BALL_COLUMN_VELOCITY = 1
const BEST_OF = 5

func main() {

	initScreen()
	initGame()
	displayFrame()
	userInput := listenUserInput()
	for !isGameOver() {
		time.Sleep(50 * time.Millisecond)
		key := readInput(userInput)
		handleUserInput(key)
		updateGameState()
		printGameState()
		if isRoundOver() {
			updateScore()
			printRoundInfo()
			time.Sleep(2500 * time.Millisecond)
			initGame()
			clearAllScreen()
			displayFrame()
		}
	}

	winner, winningPoint := getWinner()
	printGameEndInfo(winner, winningPoint)
	time.Sleep(5000 * time.Millisecond)
}

func updateGameState() {
	if gamePauseStatus || !gameReadyStatus {
		return
	}
	coordinatesToClear = append(coordinatesToClear, Coordinate{
		ball.col,
		ball.row,
	})
	for _, obj := range gameObjects {
		obj.row += obj.rowVelocity
		obj.col += obj.columnVelocity
	}
	handleBallCollision()
}

func handleBallCollision() {
	var doesBallHitPaddle bool
	frameOriginX, frameOriginY := getFrameOrigin()
	// upper and lower collision
	if ball.row < frameOriginY+1 || ball.row > frameOriginY+FRAME_HEIGHT-1 {
		ball.rowVelocity *= -1
		ball.col -= ball.columnVelocity
		borderToRegenerate = &Coordinate{
			ball.col,
			ball.row,
		}
	}

	// collision with paddles
	if ball.col+ball.columnVelocity <= frameOriginX+1 {
		doesBallHitPaddle = ball.row >= player1.row && ball.row <= (player1.row+player1.height)
	} else if ball.col+ball.columnVelocity >= frameOriginX+FRAME_WIDTH-2 {
		doesBallHitPaddle = ball.row >= player2.row && ball.row <= (player2.row+player2.height)
	}

	if doesBallHitPaddle {
		ball.columnVelocity *= -1
	}

}

func isGameOver() bool {
	return player1Points > BEST_OF/2 || player2Points > BEST_OF/2
}

func isRoundOver() bool {
	frameOriginX, _ := getFrameOrigin()
	return ball.col <= frameOriginX+1 || ball.col >= frameOriginX+FRAME_WIDTH-2
}

func getWinner() (string, int) {
	var winner string
	var winningPoint int
	if player2Points > player1Points {
		winner = "Player2"
		winningPoint = player2Points
	} else if player1Points > player2Points {
		winner = "Player1"
		winningPoint = player1Points
	}
	return winner, winningPoint
}

func printGameEndInfo(winner string, winningPoint int) {
	winnerInfo := fmt.Sprintf("%s wins with %d points", winner, winningPoint)
	gameEndInfo := fmt.Sprint("Game Over !!")

	printInCenter(screenHeight/2-1, gameEndInfo, false)
	printInCenter(screenHeight/2, winnerInfo, false)
}

func Print(x, y, h, w int, ch rune) {
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			Screen.SetContent(x+j, y+i, ch, nil, tcell.StyleDefault)
		}
	}
}

func printGameState() {
	if !gameReadyStatus {
		printInCenter(screenHeight/2-2, "Welcome to Goal !!", true)
		printInCenter(screenHeight/2-1, fmt.Sprintf("First player to get %d points wins", int(math.Floor(BEST_OF/2)+1)), true)
		printInCenter(screenHeight/2, "Press space to start the game!", true)
		return
	} else if gamePauseStatus {
		printInCenter(screenHeight/2, "Game Paused !", true)
		return
	}
	clearCoordinates()
	for _, obj := range gameObjects {
		Print(obj.col, obj.row, obj.height, obj.width, obj.symbol)
	}
	if borderToRegenerate != nil {
		Print(borderToRegenerate.x, borderToRegenerate.y, FRAME_BORDER_THICKNESS, FRAME_BORDER_THICKNESS, FRAME_BORDER_HORIZONTAL)
	}
	Screen.Show()
}

func initScreen() {
	var error error
	encoding.Register()
	Screen, error = tcell.NewScreen()
	if error != nil {
		fmt.Fprintf(os.Stderr, "%v\n", error)
		os.Exit(1)
	}
	if e := Screen.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	Screen.SetStyle(defStyle)
	screenWidth, screenHeight = Screen.Size()
}

func initGame() {
	frameOriginX, _ := getFrameOrigin()
	PADDLE_START := screenHeight/2 - PADDLE_HEIGHT/2
	player1 = &GameObject{
		row:            PADDLE_START,
		col:            frameOriginX + 1,
		height:         PADDLE_HEIGHT,
		width:          PADDLE_WIDTH,
		symbol:         PADDLE_SYMBOL,
		rowVelocity:    0,
		columnVelocity: 0,
	}
	player2 = &GameObject{
		row:            PADDLE_START,
		col:            frameOriginX + FRAME_WIDTH - 2,
		height:         PADDLE_HEIGHT,
		width:          PADDLE_WIDTH,
		symbol:         PADDLE_SYMBOL,
		rowVelocity:    0,
		columnVelocity: 0,
	}
	rowVelocityDirection := getDirectionModifier()
	columnVelocityDirection := getDirectionModifier()
	ball = &GameObject{
		row:            screenHeight / 2,
		col:            screenWidth / 2,
		height:         BALL_HEIGHT,
		width:          BALL_WIDTH,
		symbol:         BALL_SYMBOL,
		rowVelocity:    INITIAL_BALL_COLUMN_VELOCITY * rowVelocityDirection,
		columnVelocity: INITIAL_BALL_COLUMN_VELOCITY * columnVelocityDirection,
	}
	gameObjects = []*GameObject{
		player1, player2, ball,
	}
}

func listenUserInput() chan string {
	userInput := make(chan string)
	go func() {
		for {
			switch ev := Screen.PollEvent().(type) {
			case *tcell.EventKey:
				userInput <- ev.Name()
			}
		}
	}()
	return userInput
}

func readInput(userInput chan string) string {
	var key string
	select {
	case key = <-userInput:
	default:
		key = ""
	}
	return key
}

func handleUserInput(key string) {
	var x, y int
	_, frameOriginY := getFrameOrigin()
	if key == "Rune[q]" {
		Screen.Fini()
		os.Exit(0)
	} else if key == "Rune[ ]" {
		gameReadyStatus = true
	} else if key == "Rune[p]" {
		gamePauseStatus = !gamePauseStatus
	} else if !gamePauseStatus && gameReadyStatus {
		if key == "Up" && player2.row > frameOriginY+1 {
			x = player2.col
			y = player2.row + player2.height - 1
			player2.row--
		} else if key == "Down" && (player2.row+PADDLE_HEIGHT) < FRAME_HEIGHT+frameOriginY {
			x = player2.col
			y = player2.row
			player2.row++
		} else if key == "Rune[w]" && player1.row > frameOriginY+1 {
			x = player1.col
			y = player1.row + player1.height - 1
			player1.row--
		} else if key == "Rune[s]" && (player1.row+PADDLE_HEIGHT) < FRAME_HEIGHT+frameOriginY {
			x = player1.col
			y = player1.row
			player1.row++
		}
		coordinatesToClear = append(coordinatesToClear, Coordinate{
			x, y,
		})
	}
}

func clearCoordinates() {
	for _, coordinate := range coordinatesToClear {
		Print(coordinate.x, coordinate.y, 1, 1, ' ')
	}
	coordinatesToClear = nil
}

func updateScore() {
	frameOriginX, _ := getFrameOrigin()
	if ball.col <= frameOriginX+1 {
		roundWinner = "Player2"
		player2Points++
	} else if ball.col >= frameOriginX+FRAME_WIDTH-2 {
		roundWinner = "Player1"
		player1Points++
	}
}

func printRoundInfo() {
	printInCenter(screenHeight/2-1, fmt.Sprintf("%s wins the round", roundWinner), true)
	printInCenter(screenHeight/2, fmt.Sprintf("Score: Player1 %d - %d Player2", player1Points, player2Points), true)
}

func printInCenter(yAxis int, word string, trackClearCoordinates bool) {
	wordLength := len(word)
	startPointX := (screenWidth - wordLength) / 2
	for i := 0; i < wordLength; i++ {
		Print(startPointX+i, yAxis, 1, 1, rune(word[i]))
		if trackClearCoordinates {
			coordinatesToClear = append(coordinatesToClear, Coordinate{
				startPointX + i, yAxis,
			})
		}
	}
	Screen.Show()
}

func clearAllScreen() {
	Screen.Clear()
}

func getDirectionModifier() int {
	rand.Seed(time.Now().Unix())
	random := rand.Intn(11)
	return int(math.Pow(-1, float64(random)))
}

func getFrameOrigin() (int, int) {
	return (screenWidth-FRAME_WIDTH)/2 - 1, (screenHeight-FRAME_HEIGHT)/2 - 1
}

func displayFrame() {
	originX, originY := getFrameOrigin()
	var topSymbol, bottomSymbol rune
	for i := 0; i < FRAME_WIDTH; i++ {
		// display top and bottm border
		if i == 0 {
			topSymbol = FRAME_BORDER_TOP_LEFT
			bottomSymbol = FRAME_BORDER_BOTTOM_LEFT
		} else if i == FRAME_WIDTH-1 {
			topSymbol = FRAME_BORDER_TOP_RIGHT
			bottomSymbol = FRAME_BORDER_BOTTOM_RIGHT
		} else {
			topSymbol = FRAME_BORDER_HORIZONTAL
			bottomSymbol = FRAME_BORDER_HORIZONTAL
		}
		Print(originX+i, originY, FRAME_BORDER_THICKNESS, FRAME_BORDER_THICKNESS, topSymbol)
		Print(originX+i, originY+FRAME_HEIGHT, FRAME_BORDER_THICKNESS, FRAME_BORDER_THICKNESS, bottomSymbol)
	}

	for j := 1; j < FRAME_HEIGHT; j++ {
		// display side border
		Print(originX, originY+j, FRAME_BORDER_THICKNESS, FRAME_BORDER_THICKNESS, FRAME_BORDER_VERTICAL)
		Print(originX+FRAME_WIDTH-1, originY+j, FRAME_BORDER_THICKNESS, FRAME_BORDER_THICKNESS, FRAME_BORDER_VERTICAL)
	}
}
