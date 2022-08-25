package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

type GameObject struct {
	row, col, height, width, rowVelocity, columnVelocity int
	symbol                                               rune
}

var Screen tcell.Screen
var player1, player2, ball *GameObject

var gameObjects []*GameObject

const PADDLE_HEIGHT = 4
const PADDLE_WIDTH = 1
const PADDLE_SYMBOL = 0x2588
const BALL_SYMBOL = 0x25CF
const BALL_HEIGHT = 1
const BALL_WIDTH = 1
const INITIAL_BALL_ROW_VELOCITY = 1
const INITIAL_BALL_COLUMN_VELOCITY = 2

func main() {

	initScreen()
	initGame()
	userInput := listenUserInput()
	for !isGameOver() {
		time.Sleep(25 * time.Millisecond)
		key := readInput(userInput)
		handleUserInput(key)
		updateGameState()
		printGameState()
	}

	winner := getWinner()
	printGameEndInfo(winner)
	time.Sleep(5000 * time.Millisecond)
}

func updateGameState() {
	for _, obj := range gameObjects {
		obj.row += obj.rowVelocity
		obj.col += obj.columnVelocity
	}
	handleBallCollision()
}

func handleBallCollision() {
	screenWidth, screenHeight := Screen.Size()
	var doesBallHitPaddle bool

	// upper and lower collision
	if ball.row <= 0 || ball.row >= screenHeight {
		ball.rowVelocity *= -1
	}

	// collision with paddles
	if ball.col+ball.columnVelocity <= 0 {
		doesBallHitPaddle = ball.row >= player1.row && ball.row <= (player1.row+player1.height)
	} else if ball.col+ball.columnVelocity >= screenWidth {
		doesBallHitPaddle = ball.row >= player2.row && ball.row <= (player2.row+player2.height)
	}

	if doesBallHitPaddle {
		ball.columnVelocity *= -1
	}

}

func isGameOver() bool {
	screenWidth, _ := Screen.Size()
	return ball.col < 0 || ball.col > screenWidth
}

func getWinner() string {
	screenWidth, _ := Screen.Size()
	var winner string
	if ball.col < 0 {
		winner = "Player2"
	} else if ball.col > screenWidth {
		winner = "Player1"
	}
	return winner
}

func printGameEndInfo(winner string) {
	screenWidth, screenHeight := Screen.Size()
	winnerInfo := fmt.Sprintf("%s wins", winner)
	gameEndInfo := fmt.Sprint("Game Over !!")

	startPointX := screenWidth/2 - len(winnerInfo)/2
	startPointY := screenHeight / 2
	for i := 0; i < len(winnerInfo); i++ {
		Print(startPointX+i, startPointY, 1, 1, rune(winnerInfo[i]))
	}

	startPointX = screenWidth/2 - len(gameEndInfo)/2
	startPointY = screenHeight/2 - 1
	for i := 0; i < len(gameEndInfo); i++ {
		Print(startPointX+i, startPointY, 1, 1, rune(gameEndInfo[i]))
	}
	Screen.Show()
}

func Print(x, y, h, w int, ch rune) {
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			Screen.SetContent(x+j, y+i, ch, nil, tcell.StyleDefault)
		}
	}
}

func printGameState() {
	Screen.Clear()
	for _, obj := range gameObjects {
		Print(obj.col, obj.row, obj.height, obj.width, obj.symbol)
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
}

func initGame() {
	w, h := Screen.Size()
	PADDLE_START := h/2 - PADDLE_HEIGHT/2
	player1 = &GameObject{
		row:            PADDLE_START,
		col:            0,
		height:         PADDLE_HEIGHT,
		width:          PADDLE_WIDTH,
		symbol:         PADDLE_SYMBOL,
		rowVelocity:    0,
		columnVelocity: 0,
	}
	player2 = &GameObject{
		row:            PADDLE_START,
		col:            w - 1,
		height:         PADDLE_HEIGHT,
		width:          PADDLE_WIDTH,
		symbol:         PADDLE_SYMBOL,
		rowVelocity:    0,
		columnVelocity: 0,
	}
	ball = &GameObject{
		row:            h / 2,
		col:            w / 2,
		height:         BALL_HEIGHT,
		width:          BALL_WIDTH,
		symbol:         BALL_SYMBOL,
		rowVelocity:    INITIAL_BALL_COLUMN_VELOCITY,
		columnVelocity: INITIAL_BALL_COLUMN_VELOCITY,
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
	_, screenHeight := Screen.Size()
	if key == "Rune[q]" {
		Screen.Fini()
		os.Exit(0)
	} else if key == "Up" && player2.row > 0 {
		player2.row--
	} else if key == "Down" && (player2.row+PADDLE_HEIGHT) < screenHeight {
		player2.row++
	} else if key == "Rune[w]" && player1.row > 0 {
		player1.row--
	} else if key == "Rune[s]" && (player1.row+PADDLE_HEIGHT) < screenHeight {
		player1.row++
	}
}
