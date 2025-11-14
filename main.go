package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cterence/chip8-go/internal/chip8"
	"github.com/cterence/chip8-go/internal/lib"
	"github.com/urfave/cli/v3"
)

const (
	ROM_EXT = ".ch8"
)

func main() {
	var (
		logLevel   string
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
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "log level (debug, info, error)",
				Value:       "info",
				Destination: &logLevel,
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
			lib.SetLogger(logLevel)

			if rom == "" {
				return cli.ShowSubcommandHelp(c)
			}

			ext := filepath.Ext(rom)
			if ext != ROM_EXT {
				return fmt.Errorf("rom file must have %s extension, actual %s", ROM_EXT, ext)
			}

			romBytes, err := os.ReadFile(rom)
			if err != nil {
				return fmt.Errorf("failed to read rom file: %w", err)
			}

			c8 := chip8.New(
				romBytes,
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
		slog.Error("runtime error", "error", err)
		os.Exit(1)
	}
}
