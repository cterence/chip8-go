package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/Zyko0/go-sdl3/bin/binimg"
	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go/internal/lib"
)

type UI struct {
	Draw bool

	scale       int
	frameBuffer [WIDTH][HEIGHT]byte

	window   *sdl.Window
	renderer *sdl.Renderer

	keyPressed byte
	sdlKeyIDs  map[sdl.Keycode]byte
	keyState   map[byte]bool

	eventCooldown time.Time

	ResetChip8       func() error
	IsChip8Paused    func() bool
	TogglePauseChip8 func()
	TickChip8        func() error
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
		ui.window, ui.renderer, err = sdl.CreateWindowAndRenderer("CHIP-8", WIDTH*ui.scale, HEIGHT*ui.scale, sdl.WINDOW_RESIZABLE)
		if err != nil {
			return fmt.Errorf("failed to create window and renderer: %w", err)
		}
	}

	for _, v := range ui.sdlKeyIDs {
		ui.keyState[v] = false
	}

	ui.Draw = false
	ui.keyPressed = 0xFF
	ui.Reset()

	return nil
}

func (ui *UI) Update() error {
	for x := range ui.frameBuffer {
		for y := range ui.frameBuffer[x] {
			rc := &sdl.FRect{
				X: float32(x * ui.scale),
				Y: float32(y * ui.scale),
				W: float32(ui.scale),
				H: float32(ui.scale),
			}

			if err := ui.SetColor(ui.frameBuffer[x][y]); err != nil {
				return fmt.Errorf("failed to set color for pixel: %w", err)
			}

			ui.renderer.RenderFillRect(rc)
		}
	}

	if err := ui.renderer.Present(); err != nil {
		return fmt.Errorf("failed to present UI: %w", err)
	}

	ui.Draw = false

	return nil
}

func (ui *UI) DrawSprite(x, y byte, sprite []byte) bool {
	collision := false
	startYDraw := y % HEIGHT

	for row := range sprite {
		yDraw := byte(int(y)+row) % HEIGHT
		prevXDraw := x % WIDTH

		if yDraw < startYDraw {
			continue
		}

		for offset := range lib.BYTE_SIZE {
			xDraw := byte(int(x)+int(offset)) % WIDTH
			spritePixel := lib.Bit(sprite[row], 7-offset)
			oldPixel := ui.frameBuffer[xDraw][yDraw]
			newPixel := spritePixel ^ oldPixel

			if xDraw < prevXDraw {
				continue
			}

			if spritePixel == 1 && oldPixel == 1 {
				collision = true
			}

			prevXDraw = xDraw
			ui.frameBuffer[xDraw][yDraw] = newPixel
		}
	}

	ui.Draw = true

	return collision
}

func (ui *UI) Reset() {
	for x := range ui.frameBuffer {
		for y := range ui.frameBuffer[x] {
			ui.frameBuffer[x][y] = 0
		}
	}

	ui.Draw = true
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

func (ui *UI) Screenshot(romFileName string) {
	defer binimg.Load().Unload()

	screenshotFile, _ := strings.CutSuffix(filepath.Base(romFileName), ".ch8")
	screenshotFile = fmt.Sprintf("%s-%s.jpg", screenshotFile, time.Now().Format("20060102150405"))

	log.Printf("saving screenshot: %s", screenshotFile)

	surface, err := ui.renderer.ReadPixels(nil)
	if err != nil {
		log.Fatalf("failed to get surface: %v", err)

		return
	}
	defer surface.Destroy()

	if err := img.SaveJPG(surface, screenshotFile, 90); err != nil {
		log.Fatalf("failed to save screenshot: %v", err)

		return
	}
}

func (ui *UI) HandleEvents() error {
	var event sdl.Event

	for sdl.PollEvent(&event) {
		switch event.Type {
		case sdl.EVENT_QUIT, sdl.EVENT_WINDOW_DESTROYED:
			return sdl.EndLoop
		case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
			key := event.KeyboardEvent().Key
			switch key {
			case sdl.K_1, sdl.K_2, sdl.K_3, sdl.K_4, sdl.K_Q, sdl.K_W, sdl.K_E, sdl.K_R, sdl.K_A, sdl.K_S, sdl.K_D, sdl.K_F, sdl.K_Z, sdl.K_X, sdl.K_C, sdl.K_V:
				keyId := ui.sdlKeyIDs[key]
				ui.keyState[keyId] = event.Type == sdl.EVENT_KEY_DOWN
			default:
				if time.Since(ui.eventCooldown) > 100*time.Millisecond && event.Type == sdl.EVENT_KEY_DOWN {
					switch key {
					case sdl.K_SPACE:
						log.Println("reset")

						if err := ui.ResetChip8(); err != nil {
							return fmt.Errorf("failed to reset chip8: %w", err)
						}
					case sdl.K_P:
						ui.TogglePauseChip8()
					case sdl.K_T:
						if ui.IsChip8Paused() {
							return ui.TickChip8()
						}
					case sdl.K_M:
						log.Println("exit")

						return sdl.EndLoop
					}

					ui.eventCooldown = time.Now()
				}
			}
		}
	}

	return nil
}
