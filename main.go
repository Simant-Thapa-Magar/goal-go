package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

type Paddle struct {
	row, col int
}

var Screen tcell.Screen
var player1, player2 *Paddle

const PADDLE_HEIGHT = 4
const PADDLE_WIDTH = 1
const PADDLE_SYMBOL = 0x2588

func main() {

	initScreen()
	initGame()
	userInput := listenUserInput()
	for {
		printGameState()
		time.Sleep(50 * time.Millisecond)

		key := readInput(userInput)
		handleUserInput(key)

	}
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
	Print(player1.col, player1.row, PADDLE_HEIGHT, PADDLE_WIDTH, PADDLE_SYMBOL)
	Print(player2.col, player2.row, PADDLE_HEIGHT, PADDLE_WIDTH, PADDLE_SYMBOL)
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
	player1 = &Paddle{
		row: PADDLE_START,
		col: 0,
	}
	player2 = &Paddle{
		row: PADDLE_START,
		col: w - 1,
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
	} else if key == "Up" && player1.row > 0 {
		player1.row--
	} else if key == "Down" && (player1.row+PADDLE_HEIGHT) < screenHeight {
		player1.row++
	} else if key == "Rune[w]" && player2.row > 0 {
		player2.row--
	} else if key == "Rune[s]" && (player2.row+PADDLE_HEIGHT) < screenHeight {
		player2.row++
	}
}
