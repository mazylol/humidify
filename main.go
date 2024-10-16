package main

import (
	"flag"
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

const Reset = "\033[0m"

var (
	character = flag.String("character", "|", "Set the raindrop character")
	speed     = flag.String("speed", "normal", "Set the raindrop initial speed. [slow,normal,fast]")
	color     = flag.String("color", "blue", "Set the raindrop color. [blue,red,green,yellow,white]")
	density   = flag.Int("density", 32, "Set the raindrop density. Lower is more dense")

	width    = flag.Int("width", 0, "Set the width of the terminal")
	height   = flag.Int("height", 0, "Set the height of the terminal")
	gravity  = flag.Int("gravity", 10, "Set the gravity, or acceleration of the raindrops. Higher is faster.")
	noSplash = flag.Bool("no-splash", false, "Disable splash effect")
)

var colors = map[string]string{
	"blue":   "\033[34m",
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
	"white":  "\033[37m",
}

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

	grid[0][x] = colors[*color] + *character + Reset
	mu.Unlock()

	var duration time.Duration

	switch *speed {
	case "slow":
		duration = time.Duration(randRange(500, 1000)) * time.Millisecond
	case "normal":
		duration = time.Duration(randRange(300, 700)) * time.Millisecond
	case "fast":
		duration = time.Duration(randRange(100, 300)) * time.Millisecond
	}

	minDuration := 100 * time.Millisecond // Minimum sleep duration

	for i := 1; i < len(grid); i++ {
		mu.Lock()
		grid[i-1][x] = ""
		grid[i][x] = colors[*color] + *character + Reset
		mu.Unlock()

		time.Sleep(duration)

		if duration > minDuration {
			duration -= time.Duration(*gravity) * time.Millisecond
		}
	}

	mu.Lock()
	grid[len(grid)-1][x] = ""
	mu.Unlock()

	if !*noSplash {
		splash(x, cols, grid, mu)
	}
}

func main() {
	flag.Parse()

	if *width == 0 || *height == 0 {
		*width, *height = getTermWH()
	} else {
		*height--
	}

	grid := make([][]string, *height)
	for i := range grid {
		grid[i] = make([]string, *width)
	}

	cursor.Hide()
	tm.Clear()

	var mu sync.Mutex

	sema := make(chan struct{}, 50)

	go func() {
		for {
			for i := 0; i < *width; i++ {
				n := rand.Intn(*density)

				if n == 1 {
					sema <- struct{}{}
					go func(i int) {
						defer func() { <-sema }()
						handleDrop(i, *width, grid, &mu)
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
