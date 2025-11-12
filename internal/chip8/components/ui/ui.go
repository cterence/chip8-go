package ui

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go-v2/internal/lib"
)

type UI struct {
	scale int

	window   *sdl.Window
	renderer *sdl.Renderer

	screen [WIDTH][HEIGHT]byte
}

type Option func(*UI)

const (
	WIDTH  = 64
	HEIGHT = 32
)

func New(options ...Option) *UI {
	ui := &UI{}

	for _, o := range options {
		o(ui)
	}

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

	ui.window, ui.renderer, err = sdl.CreateWindowAndRenderer("chip8", WIDTH*ui.scale, HEIGHT*ui.scale, 0)
	if err != nil {
		return fmt.Errorf("failed to create window and renderer: %w", err)
	}

	return nil
}

func (ui *UI) Update() error {
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

			// slog.Debug("rect", "x", x, "y", y, "rect", rc)

			ui.renderer.RenderFillRect(rc)
		}
	}

	if err := ui.renderer.Present(); err != nil {
		return fmt.Errorf("failed to present UI: %w", err)
	}

	return ui.handleEvents()
}

func (ui *UI) DrawSprite(x, y byte, sprite []byte) {
	for i := range sprite {
		for offset := range lib.BYTE_SIZE {
			xDraw := (int(x) + int(offset)) % WIDTH
			yDraw := (int(y) + i) % HEIGHT
			pixel := ui.screen[xDraw][yDraw]
			ui.screen[xDraw][yDraw] = pixel ^ lib.Bit(sprite[i], 7-offset)
			// slog.Debug("draw sprite", "x", x, "y", y, "xDraw", xDraw, "yDraw", yDraw, "pixel", pixel)
		}
	}
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

func (ui *UI) handleEvents() error {
	var event sdl.Event

	if sdl.PollEvent(&event) {
		if event.Type == sdl.EVENT_QUIT || event.Type == sdl.EVENT_WINDOW_DESTROYED {
			return sdl.EndLoop
		}
	}

	return nil
}
