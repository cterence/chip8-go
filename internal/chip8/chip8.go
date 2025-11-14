package chip8

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cterence/chip8-go/internal/chip8/components/cpu"
	"github.com/cterence/chip8-go/internal/chip8/components/debugger"
	"github.com/cterence/chip8-go/internal/chip8/components/memory"
	"github.com/cterence/chip8-go/internal/chip8/components/timer"
	"github.com/cterence/chip8-go/internal/chip8/components/ui"
	"github.com/cterence/chip8-go/internal/lib"
)

type Chip8 struct {
	cpu      *cpu.CPU
	mem      *memory.Memory
	ui       *ui.UI
	timer    *timer.Timer
	debugger *debugger.Debugger

	uiOptions []ui.Option

	ticks         int
	paused        bool
	lastFrame     time.Time
	lastTimerTick time.Time
	lastCPUTick   time.Time

	// Options
	debug              bool
	romBytes           []byte
	romFileName        string
	headless           bool
	tickLimit          int
	exitAfterTickLimit bool
	screenshot         bool
	testFlag           byte
}

const (
	CTPS              = 500
	CPU_TICK_PERIOD   = time.Second / CTPS
	TTPS              = 60
	TIMER_TICK_PERIOD = time.Second / TTPS
	FPS               = 60
	FRAME_PERIOD      = time.Second / FPS
)

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
	debugger := debugger.New(cpu, mem)

	c8.mem = mem
	c8.cpu = cpu
	c8.ui = ui
	c8.timer = t
	c8.debugger = debugger

	c8.ui.ResetChip8 = c8.init
	c8.ui.IsChip8Paused = func() bool { return c8.paused }
	c8.ui.TogglePauseChip8 = c8.togglePause
	c8.ui.TickChip8 = c8.tick

	return c8
}

func WithDebug(debug bool) Option {
	return func(c *Chip8) {
		c.debug = debug
	}
}

func WithPauseAfter(tickLimit int) Option {
	return func(c *Chip8) {
		if c.tickLimit == 0 {
			c.tickLimit = tickLimit
		}
	}
}

func WithExitAfter(tickLimit int) Option {
	return func(c *Chip8) {
		if c.tickLimit == 0 {
			c.tickLimit = tickLimit
		}

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

func (c8 *Chip8) Run(ctx context.Context) error {
	rCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	trapSigInt(cancel)

	if err := c8.init(); err != nil {
		return fmt.Errorf("failed to init chip8: %w", err)
	}
	defer c8.ui.Destroy()

	if c8.screenshot {
		defer c8.ui.Screenshot(c8.romFileName)
	}

	for {
		select {
		case <-rCtx.Done():
			return nil
		default:
			if !c8.paused {
				c8.handleTickLimitReached(cancel)

				if err := c8.tick(); err != nil {
					return err
				}

				c8.ticks++
			}

			if !c8.headless {
				if time.Since(c8.lastFrame) >= FRAME_PERIOD && c8.ui.Draw {
					c8.lastFrame = time.Now()

					err := c8.ui.Update()
					if err != nil {
						return fmt.Errorf("failed to update UI: %w", err)
					}
				}

				if err := c8.ui.HandleEvents(); err != nil {
					return err
				}
			}
		}
	}
}

func (c8 *Chip8) init() error {
	c8.paused = false
	c8.ticks = 0
	c8.lastTimerTick = time.Now()
	c8.lastCPUTick = time.Now()
	c8.lastFrame = time.Now()

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

func (c8 *Chip8) tick() error {
	if time.Since(c8.lastCPUTick) >= CPU_TICK_PERIOD {
		c8.lastCPUTick = time.Now()
		c8.cpu.Tick()
	}

	if time.Since(c8.lastTimerTick) >= TIMER_TICK_PERIOD {
		c8.lastTimerTick = time.Now()
		c8.timer.Tick()
	}

	if c8.debug {
		log.Println(c8.debugger.DebugLog())
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

	log.Printf("rom loaded: %d bytes\n", l)
}

func (c8 *Chip8) handleTickLimitReached(cancel context.CancelFunc) {
	if c8.tickLimit > 0 && c8.ticks == c8.tickLimit {
		log.Printf("tick limit reached: %d", c8.tickLimit)

		if c8.exitAfterTickLimit {
			cancel()
		}

		c8.paused = true
	}
}

func (c8 *Chip8) togglePause() {
	c8.paused = !c8.paused
}

func trapSigInt(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cancel()
	}()
}
