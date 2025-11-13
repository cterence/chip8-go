package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cterence/chip8-go-v2/internal/chip8"
	"github.com/cterence/chip8-go-v2/internal/lib"
	"github.com/urfave/cli/v3"
)

const (
	ROM_EXT = ".ch8"
)

func main() {
	var (
		logLevel  string
		rom       string
		tickLimit int
		scale     int
		headless  bool
	)

	cmd := &cli.Command{
		Name:  "c8g",
		Usage: "chip8 emulator",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "log level (debug, info, error)",
				Value:       "info",
				Destination: &logLevel,
			},
			&cli.IntFlag{
				Name:        "tick-limit",
				Aliases:     []string{"t"},
				Usage:       "pause execution after t ticks",
				Destination: &tickLimit,
			},
			&cli.BoolFlag{
				Name:        "headless",
				Usage:       "disable ui",
				Destination: &headless,
			},
			&cli.IntFlag{
				Name:        "scale",
				Aliases:     []string{"s"},
				Usage:       "pixel and window scale factor",
				Value:       8,
				Destination: &scale,
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
				chip8.WithTickLimit(tickLimit),
				chip8.WithScale(scale),
				chip8.WithHeadless(headless),
			)

			return c8.Run()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("runtime error", "error", err)
	}
}
