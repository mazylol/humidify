package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"atomicgo.dev/cursor"
	tm "github.com/buger/goterm"
	"github.com/mattn/go-tty"
	"golang.org/x/term"
)

const Blue = "\033[34m"
const Reset = "\033[0m"

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
	return rand.Intn(max-min) + min
}

func splash(origin int, cols int, grid [][]string, mu *sync.Mutex) {
	if origin > 0 && origin < cols-1 {
		mu.Lock()
		grid[len(grid)-1][origin-1] = "'"
		grid[len(grid)-1][origin+1] = "'"
		grid[len(grid)-2][origin] = "."
		mu.Unlock()

		time.Sleep(time.Millisecond * 150)

		mu.Lock()
		grid[len(grid)-1][origin-1] = ""
		grid[len(grid)-1][origin+1] = ""
		grid[len(grid)-2][origin] = ""
		mu.Unlock()
	}
}

func handleDrop(x int, cols int, grid [][]string, mu *sync.Mutex) {
	mu.Lock()
	if grid[0][x] != "" {
		mu.Unlock()
		return
	}

	grid[0][x] = Blue + "@" + Reset
	mu.Unlock()

	duration := time.Duration(randRange(300, 700)) * time.Millisecond
	minDuration := 100 * time.Millisecond // Minimum sleep duration

	for i := 1; i < len(grid); i++ {
		mu.Lock()
		grid[i-1][x] = ""
		grid[i][x] = Blue + "@" + Reset
		mu.Unlock()

		time.Sleep(duration)

		if duration > minDuration {
			duration -= 10 * time.Millisecond
		}
	}

	mu.Lock()
	grid[len(grid)-1][x] = ""
	mu.Unlock()

	splash(x, cols, grid, mu)
}

func main() {
	cols, rows := getTermWH()

	grid := make([][]string, rows)
	for i := range grid {
		grid[i] = make([]string, cols)
	}

	cursor.Hide()
	tm.Clear()

	var mu sync.Mutex

	sema := make(chan struct{}, 50)

	go func() {
		for {
			for i := 0; i < cols; i++ {
				n := rand.Intn(32)

				if n == 1 {
					sema <- struct{}{}
					go func(i int) {
						defer func() { <-sema }()
						handleDrop(i, cols, grid, &mu)
					}(i)
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

			mu.Lock()
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
			mu.Unlock()

			tm.Flush()

			time.Sleep(50 * time.Millisecond)
		}
	}
}
