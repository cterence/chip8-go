package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go/internal/chip8"
	"github.com/cterence/chip8-go/internal/chip8/components/cpu"
	"github.com/urfave/cli/v3"
)

func main() {
	var (
		compatibilityMode cpu.CompatibilityMode
		debug             bool
		rom               string
		pauseAfter        int
		exitAfter         int
		scale             int
		headless          bool
		screenshot        bool
		testFlag          byte
		speed             float32
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
			&cli.Float32Flag{
				Name:        "speed",
				Aliases:     []string{"s"},
				Usage:       "emulator speed",
				Value:       1.0,
				Destination: &speed,
			},
			&cli.IntFlag{
				Name:        "scale",
				Usage:       "pixel and window scale factor",
				Value:       4,
				Destination: &scale,
			},
			&cli.Uint8Flag{
				Name:        "test-flag",
				Aliases:     []string{"t"},
				Usage:       "populate 0x1FF address before run (used by timendus tests)",
				Destination: &testFlag,
			},
			&cli.StringFlag{
				Name:    "compatibility-mode",
				Aliases: []string{"m"},
				Usage:   "force compatibility mode (chip8, super, xo)",
				Action: func(_ context.Context, _ *cli.Command, mode string) error {
					switch mode {
					case "chip8":
						compatibilityMode = cpu.CM_CHIP8
					case "super":
						compatibilityMode = cpu.CM_SUPERCHIP
					case "xo":
						compatibilityMode = cpu.CM_XOCHIP
					default:
						return fmt.Errorf("unknown compatibility mode: %s", mode)
					}

					return nil
				},
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
				chip8.WithCompatibilityMode(compatibilityMode),
				chip8.WithDebug(debug),
				chip8.WithPauseAfter(pauseAfter),
				chip8.WithExitAfter(exitAfter),
				chip8.WithScreenshot(screenshot, rom),
				chip8.WithScale(scale),
				chip8.WithSpeed(speed),
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
