package main

import (
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"

	"atomicgo.dev/cursor"
	tm "github.com/buger/goterm"
	"github.com/mattn/go-tty"
	"golang.org/x/term"
)

const Blue = "\033[34m"
const Reset = "\033[0m"

// cols, rows
func getTermWH() (int, int) {
	if !term.IsTerminal(0) {
		panic("not in a term")
	}

	width, height, err := term.GetSize(0)
	if err != nil {
		panic("cannot get term size!")
	}

	return width, height - 1
}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func splash(origin int, cols int, grid [][]string) {
	if origin > 0 && origin < cols-1 {
		grid[len(grid)-1][origin-1] = "'"
		grid[len(grid)-1][origin+1] = "'"
		grid[len(grid)-2][origin] = "."

		time.Sleep(time.Millisecond * 150)

		grid[len(grid)-1][origin-1] = ""
		grid[len(grid)-1][origin+1] = ""
		grid[len(grid)-2][origin] = ""
	}
}

func handleDrop(x int, cols int, grid [][]string) {
	if grid[0][x] != "" {
		return
	}

	grid[0][x] = Blue + "@" + Reset

	duration := time.Duration(randRange(300, 700)) * time.Millisecond

	for i := 1; i < len(grid); i++ {
		grid[i-1][x] = ""
		grid[i][x] = Blue + "@" + Reset

		time.Sleep(duration)

		duration -= 10 * time.Millisecond
	}

	grid[len(grid)-1][x] = ""

	splash(x, cols, grid)
}

func main() {
	cols, rows := getTermWH()

	grid := make([][]string, rows)
	for i := range grid {
		grid[i] = make([]string, cols)
	}

	cursor.Hide()
	tm.Clear()

	go func() {
		for {
			for i := 0; i < cols; i++ {
				n := rand.IntN(32)

				if n == 1 {
					go handleDrop(i, cols, grid)
				}
			}

			time.Sleep(1000 * time.Millisecond)
		}
	}()

	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		_, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}

		signal.Stop(stop)
		close(stop)
	}()

	for {
		select {
		case <-stop:
			cursor.Show()
			return
		default:
			tm.MoveCursor(1, 1)

			for _, row := range grid {
				for _, item := range row {
					if item == "" {
						tm.Print(" ")
						continue
					}

					tm.Print(item)
				}

				tm.Println()
			}

			tm.Flush()
		}
	}
}
