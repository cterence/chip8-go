package chip8

import (
	"fmt"
	"log/slog"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go-v2/internal/chip8/components/cpu"
	"github.com/cterence/chip8-go-v2/internal/chip8/components/memory"
	"github.com/cterence/chip8-go-v2/internal/chip8/components/ui"
	"github.com/cterence/chip8-go-v2/internal/lib"
)

type Chip8 struct {
	cpu *cpu.CPU
	mem *memory.Memory
	ui  *ui.UI

	uiOptions []ui.Option

	tickLimit int
	scale     int
}

type Option func(*Chip8)

func New(options ...Option) *Chip8 {
	c8 := &Chip8{}

	for _, o := range options {
		o(c8)
	}

	mem := memory.New()
	ui := ui.New(c8.uiOptions...)
	cpu := cpu.New(mem, ui)

	c8.mem = mem
	c8.cpu = cpu
	c8.ui = ui

	return c8
}

func WithTickLimit(tickLimit int) Option {
	return func(c *Chip8) {
		c.tickLimit = tickLimit
	}
}

func WithScale(scale int) Option {
	return func(c *Chip8) {
		c.uiOptions = append(c.uiOptions, ui.WithScale(scale))
	}
}

func (c8 *Chip8) LoadROM(romBytes []byte) {
	l := len(romBytes)
	lib.Assert(l <= int(memory.PROGRAM_RAM_SIZE), fmt.Errorf("rom file size %d is bigger than chip8 program ram %d", l, memory.PROGRAM_RAM_SIZE))

	for i, v := range romBytes {
		a := uint16(i) + memory.PROGRAM_RAM_START
		c8.mem.Write(a, v)
	}

	slog.Debug("rom loaded", "size", l)
}

func (c8 *Chip8) Run() error {
	c8.cpu.Init()

	if err := c8.ui.Init(); err != nil {
		return fmt.Errorf("failed to init UI: %w", err)
	}
	defer c8.ui.Destroy()

	totalCycles, cycles := 0, 0

	var (
		err   error
		ticks int
	)

	for err == nil {
		if c8.tickLimit > 0 && ticks == c8.tickLimit {
			slog.Info("tick limit reached", "ticks", ticks)

			c8.cpu.Paused = true
		}

		cycles = c8.cpu.Tick()
		totalCycles += cycles
		err = c8.ui.Update()

		ticks++
	}

	if err != nil && err != sdl.EndLoop {
		return err
	}

	return nil
}
