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
	compatibilityMode lib.CompatibilityMode
	scale             int

	SelectedFrameBuffer SelectedFrameBuffer
	frameBuffer         [2][WIDTH][HEIGHT]byte
	colorPalette        [4]uint32

	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	surface  *sdl.Surface

	keyPressed *byte
	sdlKeyIDs  map[sdl.Keycode]byte
	keyState   map[byte]bool

	res             int
	eventCooldown   time.Time
	scrollDirection ScrollDirection
	scrollPixels    int32

	ResetChip8       func() error
	IsChip8Paused    func() bool
	TogglePauseChip8 func()
	TickChip8        func() error
}

type Option func(*UI)

const (
	WIDTH  = 128
	HEIGHT = 64
)

type ScrollDirection uint8

const (
	SD_NONE ScrollDirection = iota
	SD_LEFT
	SD_RIGHT
	SD_DOWN
	SD_UP
)

type SelectedFrameBuffer uint8

const (
	SF_NONE SelectedFrameBuffer = iota
	SF_0
	SF_1
	SF_BOTH
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
	ui.colorPalette = [4]uint32{0xFF0C0F1C, 0xFF87B6FF, 0xFFFFA7C8, 0xFFD0A7FF}
	// ui.colorPalette = [4]uint32{0xFF0C0F1C, 0xFF87B6FF, 0xFFFFA7C8, 0xFFFFA7C8}

	return ui
}

func WithScale(scale int) Option {
	return func(u *UI) {
		u.scale = scale
	}
}

func WithCompatibilityMode(mode lib.CompatibilityMode) Option {
	return func(ui *UI) {
		ui.compatibilityMode = mode
	}
}

func (ui *UI) Init() error {
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

	if ui.texture == nil {
		ui.texture, err = ui.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, WIDTH*ui.scale, HEIGHT*ui.scale)
		if err != nil {
			return fmt.Errorf("failed to create SDL texture: %w", err)
		}
	}

	if ui.surface == nil {
		ui.surface, err = sdl.CreateSurface(WIDTH*ui.scale, HEIGHT*ui.scale, sdl.PIXELFORMAT_ARGB8888)
		if err != nil {
			return fmt.Errorf("failed to create SDL surface: %w", err)
		}
	}

	for _, v := range ui.sdlKeyIDs {
		ui.keyState[v] = false
	}

	ui.scrollDirection = SD_NONE
	ui.scrollPixels = 0
	ui.res = 2
	ui.keyPressed = nil
	ui.SelectedFrameBuffer = 0
	ui.Reset()

	return nil
}

func (ui *UI) Update() error {
	for x := range WIDTH {
		for y := range HEIGHT {
			rc := &sdl.Rect{
				X: int32(x * ui.scale),
				Y: int32(y * ui.scale),
				W: int32(ui.scale * ui.res),
				H: int32(ui.scale * ui.res),
			}

			pixel0, pixel1 := ui.frameBuffer[0][x][y], ui.frameBuffer[1][x][y]

			color := ui.colorPalette[pixel1<<1|pixel0]

			if err := ui.surface.FillRect(rc, color); err != nil {
				return fmt.Errorf("failed to fill rect: %w", err)
			}
		}
	}

	ui.scrollDirection = SD_NONE
	ui.scrollPixels = 0

	if err := ui.texture.Update(nil, ui.surface.Pixels(), ui.surface.Pitch); err != nil {
		return fmt.Errorf("failed to update texture: %w", err)
	}

	if err := ui.renderer.Clear(); err != nil {
		return fmt.Errorf("failed to clear renderer: %w", err)
	}

	if err := ui.renderer.RenderTexture(ui.texture, nil, nil); err != nil {
		return fmt.Errorf("failed to render texture: %w", err)
	}

	if err := ui.renderer.Present(); err != nil {
		return fmt.Errorf("failed to present UI: %w", err)
	}

	return nil
}

func (ui *UI) ToggleHiRes(enable bool) {
	if enable {
		ui.res = 1
	} else {
		ui.res = 2
	}
}

func (ui *UI) DrawSprite(x, y byte, sprite []byte) bool {
	collision := false

	fbIDs := ui.getFrameBufferIDs()

	switch len(fbIDs) {
	case 1:
		collision = ui.drawSpriteOnFramebuffer(x, y, sprite, fbIDs[0])
	case 2:
		collision = ui.drawSpriteOnFramebuffer(x, y, sprite[:len(sprite)/2], fbIDs[0]) || ui.drawSpriteOnFramebuffer(x, y, sprite[len(sprite)/2:], fbIDs[1])
	}

	return collision
}

func (ui *UI) Reset() {
	for _, i := range ui.getFrameBufferIDs() {
		ui.resetFramebuffer(i)
	}
}

func (ui *UI) Destroy() {
	ui.renderer.Destroy()
	ui.window.Destroy()
	ui.surface.Destroy()
	ui.texture.Destroy()
}

func (ui *UI) IsKeyPressed(key byte) bool {
	return ui.keyState[key]
}

func (ui *UI) GetPressedKey() *byte {
	for id, pressed := range ui.keyState {
		if pressed {
			return &id
		}
	}

	return nil
}

func (ui *UI) SelectFrameBuffer(id byte) {
	switch id {
	case 0:
		ui.SelectedFrameBuffer = SF_NONE
	case 1:
		ui.SelectedFrameBuffer = SF_0
	case 2:
		ui.SelectedFrameBuffer = SF_1
	case 3:
		ui.SelectedFrameBuffer = SF_BOTH
	default:
		log.Fatalf("framebuffer id must be 0, 1, 2 or 3 actual %d", id)
	}
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

func (ui *UI) Scroll(sd ScrollDirection, pixels int) {
	for _, i := range ui.getFrameBufferIDs() {
		ui.scrollFrameBuffer(sd, pixels, i)
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

func (ui *UI) scrolledCoords(x, y int, sd ScrollDirection, pixels int) (int, int) {
	newX, newY := x, y

	switch sd {
	case SD_LEFT:
		newX -= pixels * ui.res
	case SD_RIGHT:
		newX += pixels * ui.res
	case SD_UP:
		newY -= pixels * ui.res
	case SD_DOWN:
		newY += pixels * ui.res
	}

	return newX, newY
}

func (ui *UI) getFrameBufferIDs() []byte {
	switch ui.SelectedFrameBuffer {
	case SF_NONE, SF_0:
		return []byte{0}
	case SF_1:
		return []byte{1}
	case SF_BOTH:
		return []byte{0, 1}
	default:
		log.Fatalf("unknown framebuffer ID: %d", ui.SelectedFrameBuffer)
	}

	return nil
}

func (ui *UI) drawSpriteOnFramebuffer(x, y byte, sprite []byte, frameBufferID byte) bool {
	collision := false
	startYDraw := (y * byte(ui.res)) % HEIGHT
	spriteWidth := byte(8)
	spriteHeight := byte(len(sprite))

	if len(sprite) == 32 {
		spriteWidth = 16
		spriteHeight = 16
	}

	for row := range spriteHeight {
		yDraw := (y + row) * byte(ui.res) % HEIGHT
		prevXDraw := (x * byte(ui.res)) % WIDTH

		if yDraw < startYDraw {
			continue
		}

		spriteRow := uint16(sprite[row])

		if spriteWidth == 16 {
			spriteRow = (uint16(sprite[row*2]) << uint16(lib.BYTE_SIZE)) | uint16(sprite[row*2+1])
		}

		for offset := range spriteWidth {
			xDraw := byte((int(x)+int(offset))*ui.res) % WIDTH
			spritePixel := lib.Bit(spriteRow, uint16(spriteWidth-1-offset))
			oldPixel := ui.frameBuffer[frameBufferID][xDraw][yDraw]
			newPixel := spritePixel ^ oldPixel

			if xDraw < prevXDraw {
				continue
			}

			if spritePixel == 1 && oldPixel == 1 {
				collision = true
			}

			prevXDraw = xDraw

			// if spritePixel == 1 {
			// 	fmt.Printf("Drawing pixel at (%d, %d)\n", xDraw, yDraw)
			// }

			ui.frameBuffer[frameBufferID][xDraw][yDraw] = newPixel
			if ui.res == 2 {
				ui.frameBuffer[frameBufferID][xDraw+1][yDraw] = newPixel
				ui.frameBuffer[frameBufferID][xDraw][yDraw+1] = newPixel
				ui.frameBuffer[frameBufferID][xDraw+1][yDraw+1] = newPixel
			}
		}
	}

	return collision
}

func (ui *UI) resetFramebuffer(frameBufferID byte) {
	for x := range WIDTH {
		for y := range HEIGHT {
			ui.frameBuffer[frameBufferID][x][y] = 0
		}
	}
}

func (ui *UI) scrollFrameBuffer(sd ScrollDirection, pixels int, frameBufferID byte) {
	var tmpBuf [WIDTH][HEIGHT]byte

	for x := range WIDTH {
		for y := range HEIGHT {
			newX, newY := ui.scrolledCoords(x, y, sd, pixels)
			if newX < 0 || newY < 0 || newX >= WIDTH || newY >= HEIGHT {
				continue
			}

			tmpBuf[newX][newY] = ui.frameBuffer[frameBufferID][x][y]
		}
	}

	copy(ui.frameBuffer[frameBufferID][:][:], tmpBuf[:][:])
}
