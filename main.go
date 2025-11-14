package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go/internal/chip8"
	"github.com/urfave/cli/v3"
)

func main() {
	var (
		debug      bool
		rom        string
		pauseAfter int
		exitAfter  int
		scale      int
		headless   bool
		screenshot bool
		testFlag   byte
	)

	cmd := &cli.Command{
		Name:  "chip8-go",
		Usage: "chip8 emulator",
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						&cli.IntFlag{
							Name:        "pause-after",
							Aliases:     []string{"p"},
							Usage:       "pause execution after t ticks",
							Destination: &pauseAfter,
						},
					},
					{
						&cli.IntFlag{
							Name:        "exit-after",
							Aliases:     []string{"e"},
							Usage:       "exit after t ticks",
							Destination: &exitAfter,
						},
					},
				},
			},
			{
				Flags: [][]cli.Flag{
					{
						&cli.BoolFlag{
							Name:        "headless",
							Usage:       "disable ui",
							Destination: &headless,
						},
					},
					{
						&cli.BoolFlag{
							Name:        "screenshot",
							Usage:       "save screenshot on exit",
							Destination: &screenshot,
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "print debug logs",
				Destination: &debug,
			},
			&cli.IntFlag{
				Name:        "scale",
				Aliases:     []string{"s"},
				Usage:       "pixel and window scale factor",
				Value:       8,
				Destination: &scale,
			},
			&cli.Uint8Flag{
				Name:        "test-flag",
				Aliases:     []string{"t"},
				Usage:       "populate 0x1FF address before run (used by timendus tests)",
				Destination: &testFlag,
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "rom",
				UsageText:   "rom path",
				Destination: &rom,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if rom == "" {
				return cli.ShowSubcommandHelp(c)
			}

			romBytes, err := os.ReadFile(rom)
			if err != nil {
				return fmt.Errorf("failed to read rom file: %w", err)
			}

			c8 := chip8.New(
				romBytes,
				chip8.WithDebug(debug),
				chip8.WithPauseAfter(pauseAfter),
				chip8.WithExitAfter(exitAfter),
				chip8.WithScreenshot(screenshot, rom),
				chip8.WithScale(scale),
				chip8.WithHeadless(headless),
				chip8.WithTestFlag(testFlag),
			)

			return c8.Run(ctx)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		if !errors.Is(err, sdl.EndLoop) {
			log.Fatalf("runtime error: %v", err)
		}
	}
}
