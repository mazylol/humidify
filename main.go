package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"atomicgo.dev/cursor"
	tm "github.com/buger/goterm"
	"github.com/mattn/go-tty"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

const Reset = "\033[0m"

var character string
var speed string
var color string
var density int
var width int
var height int
var gravity int
var noSplash bool

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

	grid[0][x] = colors[color] + character + Reset
	mu.Unlock()

	var duration time.Duration

	switch speed {
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
		grid[i][x] = colors[color] + character + Reset
		mu.Unlock()

		time.Sleep(duration)

		if duration > minDuration {
			duration -= time.Duration(gravity) * time.Millisecond
		}
	}

	mu.Lock()
	grid[len(grid)-1][x] = ""
	mu.Unlock()

	if !noSplash {
		splash(x, cols, grid, mu)
	}
}

func main() {
	cmd := &cli.Command{
		Name:                  "humidify",
		Usage:                 "rain in the terminal",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "character",
				Aliases:     []string{"ch"},
				Value:       "|",
				Usage:       "Set the raindrop character",
				Destination: &character,
			},
			&cli.StringFlag{
				Name:        "speed",
				Aliases:     []string{"s"},
				Value:       "normal",
				Usage:       "Set the raindrop initial speed. [slow,normal,fast]",
				Destination: &speed,
			},
			&cli.StringFlag{
				Name:        "color",
				Aliases:     []string{"co"},
				Value:       "blue",
				Usage:       "Set the raindrop color. [blue,red,green,yellow,white]",
				Destination: &color,
			},
			&cli.IntFlag{
				Name:        "density",
				Aliases:     []string{"d"},
				Value:       32,
				Usage:       "Set the raindrop density",
				Destination: &density,
			},
			&cli.IntFlag{
				Name:        "width",
				Aliases:     []string{"w"},
				Value:       0,
				Usage:       "Set the width of the terminal",
				Destination: &width,
			},
			&cli.IntFlag{
				Name:        "height",
				Aliases:     []string{"h"},
				Value:       0,
				Usage:       "Set the height of the terminal",
				Destination: &height,
			},
			&cli.IntFlag{
				Name:        "gravity",
				Aliases:     []string{"g"},
				Value:       10,
				Usage:       "Set the gravity, or acceleration of the raindrops. Higher is faster.",
				Destination: &gravity,
			},
			&cli.BoolFlag{
				Name:        "no-splash",
				Aliases:     []string{"ns"},
				Value:       false,
				Usage:       "Disable splash effect",
				Destination: &noSplash,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if width == 0 || height == 0 {
				width, height = getTermWH()
			} else {
				height--
			}

			grid := make([][]string, height)
			for i := range grid {
				grid[i] = make([]string, width)
			}

			cursor.Hide()
			tm.Clear()

			var mu sync.Mutex

			sema := make(chan struct{}, 50)

			go func() {
				for {
					for i := 0; i < width; i++ {
						n := rand.Intn(density)

						if n == 1 {
							sema <- struct{}{}
							go func(i int) {
								defer func() { <-sema }()
								handleDrop(i, width, grid, &mu)
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
					return nil
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

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
