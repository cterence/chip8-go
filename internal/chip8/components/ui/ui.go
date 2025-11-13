package ui

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go-v2/internal/lib"
)

type UI struct {
	scale  int
	screen [WIDTH][HEIGHT]byte

	window   *sdl.Window
	renderer *sdl.Renderer

	keyPressed byte
	sdlKeyIDs  map[sdl.Keycode]byte
	keyState   map[byte]bool

	lastTickTime time.Time

	ResetChip8 func() error
	resetTime  time.Time
}

type Option func(*UI)

const (
	WIDTH                 = 64
	HEIGHT                = 32
	FPS                   = 500
	TARGET_FRAME_DURATION = time.Second / FPS
)

func New(options ...Option) *UI {
	ui := &UI{}

	for _, o := range options {
		o(ui)
	}

	ui.sdlKeyIDs = map[sdl.Keycode]byte{
		sdl.K_1: 0x1,
		sdl.K_2: 0x2,
		sdl.K_3: 0x3,
		sdl.K_4: 0xC,
		sdl.K_Q: 0x4,
		sdl.K_W: 0x5,
		sdl.K_E: 0x6,
		sdl.K_R: 0xD,
		sdl.K_A: 0x7,
		sdl.K_S: 0x8,
		sdl.K_D: 0x9,
		sdl.K_F: 0xE,
		sdl.K_Z: 0xA,
		sdl.K_X: 0x0,
		sdl.K_C: 0xB,
		sdl.K_V: 0xF,
	}

	ui.keyState = make(map[byte]bool)

	return ui
}

func WithScale(scale int) Option {
	return func(u *UI) {
		u.scale = scale
	}
}

func (ui *UI) Init() error {
	if err := sdl.LoadLibrary(sdl.Path()); err != nil {
		return fmt.Errorf("failed to load sdl: %w", err)
	}

	err := sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		return fmt.Errorf("failed to init sdl: %w", err)
	}

	if ui.window == nil && ui.renderer == nil {
		ui.window, ui.renderer, err = sdl.CreateWindowAndRenderer("chip8", WIDTH*ui.scale, HEIGHT*ui.scale, 0)
		if err != nil {
			return fmt.Errorf("failed to create window and renderer: %w", err)
		}
	}

	for _, v := range ui.sdlKeyIDs {
		ui.keyState[v] = false
	}

	ui.lastTickTime = time.Now()
	ui.keyPressed = 0xFF
	ui.Reset()

	return nil
}

func (ui *UI) Update(tickTime time.Time) error {
	tickDuration := tickTime.Sub(ui.lastTickTime)

	if tickDuration < TARGET_FRAME_DURATION {
		delay := TARGET_FRAME_DURATION - tickDuration
		slog.Debug("delaying", "delay", delay)
		sdl.Delay(uint32(delay.Milliseconds()))
	}

	ui.lastTickTime = time.Now()

	for x := range ui.screen {
		for y := range ui.screen[x] {
			rc := &sdl.FRect{
				X: float32(x * ui.scale),
				Y: float32(y * ui.scale),
				W: float32(ui.scale),
				H: float32(ui.scale),
			}

			if err := ui.SetColor(ui.screen[x][y]); err != nil {
				return fmt.Errorf("failed to set color for pixel: %w", err)
			}

			ui.renderer.RenderFillRect(rc)
		}
	}

	if err := ui.renderer.Present(); err != nil {
		return fmt.Errorf("failed to present UI: %w", err)
	}

	return ui.handleEvents()
}

func (ui *UI) DrawSprite(x, y byte, sprite []byte) bool {
	erased := false
	startYDraw := y % HEIGHT

	for row := range sprite {
		yDraw := byte(int(y)+row) % HEIGHT
		prevXDraw := x % WIDTH

		if yDraw < startYDraw {
			continue
		}

		for offset := range lib.BYTE_SIZE {
			xDraw := byte(int(x)+int(offset)) % WIDTH
			pixel := ui.screen[xDraw][yDraw]
			newPixel := pixel ^ lib.Bit(sprite[row], 7-offset)

			if xDraw < prevXDraw {
				continue
			}

			if pixel == 1 && newPixel == 0 {
				erased = true
			}

			prevXDraw = xDraw
			ui.screen[xDraw][yDraw] = newPixel
		}
	}

	return erased
}

func (ui *UI) Reset() {
	for x := range ui.screen {
		for y := range ui.screen[x] {
			ui.screen[x][y] = 0
		}
	}
}

func (ui *UI) Destroy() {
	sdl.Quit()
	ui.renderer.Destroy()
	ui.window.Destroy()
}

func (ui *UI) SetColor(pixel byte) error {
	lib.Assert(pixel == 0 || pixel == 1, fmt.Errorf("pixel must be 0 or 1, actual %d", pixel))

	return ui.renderer.SetDrawColor(pixel*255, pixel*255, pixel*255, pixel*255)
}

func (ui *UI) IsKeyPressed(key byte) bool {
	return ui.keyState[key]
}

func (ui *UI) GetPressedKey() byte {
	for id, pressed := range ui.keyState {
		if pressed {
			return id
		}
	}

	return 0xFF
}

func (ui *UI) handleEvents() error {
	var event sdl.Event

	for sdl.PollEvent(&event) {
		switch event.Type {
		case sdl.EVENT_QUIT, sdl.EVENT_WINDOW_DESTROYED:
			return sdl.EndLoop
		case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
			keyId := ui.sdlKeyIDs[event.KeyboardEvent().Key]

			ui.keyState[keyId] = event.Type == sdl.EVENT_KEY_DOWN
			if event.KeyboardEvent().Key == sdl.K_SPACE && time.Since(ui.resetTime) > 200*time.Millisecond {
				if err := ui.ResetChip8(); err != nil {
					return fmt.Errorf("failed to reset chip8: %w", err)
				}

				ui.resetTime = time.Now()
			}

			if event.KeyboardEvent().Key == sdl.K_M {
				return sdl.EndLoop
			}
		}
	}

	return nil
}
