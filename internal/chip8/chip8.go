package chip8

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go/internal/chip8/components/cpu"
	"github.com/cterence/chip8-go/internal/chip8/components/memory"
	"github.com/cterence/chip8-go/internal/chip8/components/timer"
	"github.com/cterence/chip8-go/internal/chip8/components/ui"
	"github.com/cterence/chip8-go/internal/lib"
)

type Chip8 struct {
	cpu   *cpu.CPU
	mem   *memory.Memory
	ui    *ui.UI
	timer *timer.Timer

	uiOptions []ui.Option

	romBytes           []byte
	romFileName        string
	headless           bool
	tickLimit          int
	exitAfterTickLimit bool
	screenshot         bool
	testFlag           byte
}

type Option func(*Chip8)

func New(romBytes []byte, options ...Option) *Chip8 {
	c8 := &Chip8{
		romBytes: romBytes,
	}

	for _, o := range options {
		o(c8)
	}

	mem := memory.New()
	ui := ui.New(c8.uiOptions...)
	t := timer.New()
	cpu := cpu.New(mem, ui, t)

	c8.mem = mem
	c8.cpu = cpu
	c8.ui = ui
	c8.timer = t

	c8.ui.ResetChip8 = c8.Init

	return c8
}

func WithPauseAfter(tickLimit int) Option {
	return func(c *Chip8) {
		c.tickLimit = tickLimit
	}
}

func WithExitAfter(tickLimit int) Option {
	return func(c *Chip8) {
		c.tickLimit = tickLimit
		c.exitAfterTickLimit = tickLimit > 0
	}
}

func WithScreenshot(screenshot bool, rom string) Option {
	return func(c *Chip8) {
		c.screenshot = screenshot
		c.romFileName = rom
	}
}

func WithScale(scale int) Option {
	return func(c *Chip8) {
		c.uiOptions = append(c.uiOptions, ui.WithScale(scale))
	}
}

func WithHeadless(headless bool) Option {
	return func(c *Chip8) {
		c.headless = headless
	}
}

func WithTestFlag(testFlag byte) Option {
	return func(c *Chip8) {
		c.testFlag = testFlag
	}
}

func (c8 *Chip8) Init() error {
	c8.mem.Init()
	c8.cpu.Init()
	c8.timer.Init()

	if !c8.headless {
		if err := c8.ui.Init(); err != nil {
			return fmt.Errorf("failed to init UI: %w", err)
		}
	}

	if c8.testFlag != 0 {
		c8.mem.Write(0x1FF, c8.testFlag)
	}

	c8.loadROM()

	return nil
}

func (c8 *Chip8) Run(ctx context.Context) error {
	rCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	trapSigInt(cancel)

	if err := c8.Init(); err != nil {
		return fmt.Errorf("failed to init chip8: %w", err)
	}
	defer c8.ui.Destroy()

	if c8.screenshot {
		defer c8.ui.Screenshot(c8.romFileName)
	}

	var (
		err   error
		ticks int
	)

	for err == nil {
		select {
		case <-rCtx.Done():
			return nil
		default:
			if c8.tickLimit > 0 && ticks == c8.tickLimit {
				slog.Info("tick limit reached", "ticks", ticks)

				if c8.exitAfterTickLimit {
					return nil
				}

				c8.cpu.Paused = true
			}

			tickTime := time.Now()

			c8.cpu.Tick()

			if !c8.headless {
				err = c8.ui.Update(tickTime)
			}

			c8.timer.Tick(tickTime)

			ticks++
		}
	}

	if err != nil && err != sdl.EndLoop {
		return err
	}

	return nil
}

func (c8 *Chip8) loadROM() {
	l := len(c8.romBytes)
	lib.Assert(l <= int(memory.PROGRAM_RAM_SIZE), fmt.Errorf("rom file size %d is bigger than chip8 program ram %d", l, memory.PROGRAM_RAM_SIZE))

	for i, b := range c8.romBytes {
		a := uint16(i) + memory.PROGRAM_RAM_START
		c8.mem.Write(a, b)
	}

	slog.Debug("rom loaded", "size", l)
}

func trapSigInt(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cancel()
	}()
}
