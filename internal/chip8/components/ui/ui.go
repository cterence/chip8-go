package ui

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go-v2/internal/lib"
)

type UI struct {
	keymap keymap
	scale  int
	screen [WIDTH][HEIGHT]byte

	window   *sdl.Window
	renderer *sdl.Renderer
}

type keymap struct {
	key1 bool
	key2 bool
	key3 bool
	key4 bool
	keyQ bool
	keyW bool
	keyE bool
	keyR bool
	keyA bool
	keyS bool
	keyD bool
	keyF bool
	keyZ bool
	keyX bool
	keyC bool
	keyV bool
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

	for i := range sprite {
		for offset := range lib.BYTE_SIZE {
			xDraw := (int(x) + int(offset)) % WIDTH
			yDraw := (int(y) + i) % HEIGHT
			pixel := ui.screen[xDraw][yDraw]

			newPixel := pixel ^ lib.Bit(sprite[i], 7-offset)
			if pixel == 1 && newPixel == 0 {
				erased = true
			}

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
	pressed, keyExists := false, true

	switch key {
	case 0x1:
		pressed = ui.keymap.key1
	case 0x2:
		pressed = ui.keymap.key2
	case 0x3:
		pressed = ui.keymap.key3
	case 0xC:
		pressed = ui.keymap.key4
	case 0x4:
		pressed = ui.keymap.keyQ
	case 0x5:
		pressed = ui.keymap.keyW
	case 0x6:
		pressed = ui.keymap.keyE
	case 0xD:
		pressed = ui.keymap.keyR
	case 0x7:
		pressed = ui.keymap.keyA
	case 0x8:
		pressed = ui.keymap.keyS
	case 0x9:
		pressed = ui.keymap.keyD
	case 0xE:
		pressed = ui.keymap.keyF
	case 0xA:
		pressed = ui.keymap.keyZ
	case 0x0:
		pressed = ui.keymap.keyX
	case 0xB:
		pressed = ui.keymap.keyC
	case 0xF:
		pressed = ui.keymap.keyV
	default:
		keyExists = false
	}

	lib.Assert(keyExists, fmt.Errorf("key does not %X exist", key))

	return pressed
}

func (ui *UI) PollKeyPress() byte {
	pressed := byte(0xFF)

	if ui.keymap.key1 {
		return 0x1
	}

	if ui.keymap.key2 {
		return 0x2
	}

	if ui.keymap.key3 {
		return 0x3
	}

	if ui.keymap.key4 {
		return 0xC
	}

	if ui.keymap.keyQ {
		return 0x4
	}

	if ui.keymap.keyW {
		return 0x5
	}

	if ui.keymap.keyE {
		return 0x6
	}

	if ui.keymap.keyR {
		return 0xD
	}

	if ui.keymap.keyA {
		return 0x7
	}

	if ui.keymap.keyS {
		return 0x8
	}

	if ui.keymap.keyD {
		return 0x9
	}

	if ui.keymap.keyF {
		return 0xE
	}

	if ui.keymap.keyZ {
		return 0xA
	}

	if ui.keymap.keyX {
		return 0x0
	}

	if ui.keymap.keyC {
		return 0xB
	}

	if ui.keymap.keyV {
		return 0xF
	}

	return pressed
}

func (ui *UI) handleEvents() error {
	var event sdl.Event

	if sdl.PollEvent(&event) {
		switch event.Type {
		case sdl.EVENT_QUIT, sdl.EVENT_WINDOW_DESTROYED:
			return sdl.EndLoop
		case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
			pressed := event.Type == sdl.EVENT_KEY_DOWN
			switch event.KeyboardEvent().Key {
			case sdl.K_1:
				ui.keymap.key1 = pressed
			case sdl.K_2:
				ui.keymap.key2 = pressed
			case sdl.K_3:
				ui.keymap.key3 = pressed
			case sdl.K_4:
				ui.keymap.key4 = pressed
			case sdl.K_Q:
				ui.keymap.keyQ = pressed
			case sdl.K_W:
				ui.keymap.keyW = pressed
			case sdl.K_E:
				ui.keymap.keyE = pressed
			case sdl.K_R:
				ui.keymap.keyR = pressed
			case sdl.K_A:
				ui.keymap.keyA = pressed
			case sdl.K_S:
				ui.keymap.keyS = pressed
			case sdl.K_D:
				ui.keymap.keyD = pressed
			case sdl.K_F:
				ui.keymap.keyF = pressed
			case sdl.K_Z:
				ui.keymap.keyZ = pressed
			case sdl.K_X:
				ui.keymap.keyX = pressed
			case sdl.K_C:
				ui.keymap.keyC = pressed
			case sdl.K_V:
				ui.keymap.keyV = pressed
			}
		}
	}

	return nil
}
