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
			for col := range grid[0] {
				n := rand.IntN(16)

				if n == 1 {
					grid[0][col] = Blue + "@" + Reset
				}
			}

            time.Sleep(time.Second)

            for col := range grid[0] {
                grid[0][col] = " "
            }

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
