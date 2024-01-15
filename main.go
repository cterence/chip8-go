package main

import (
	"fmt"
	"os"

	"github.com/cterence/chip8-go/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

func initInput() {
	fmt.Println("Init input : Not implemented yet")
}

func initGraphics() *sdl.Renderer {
	// Initialize SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	// Create a window
	window, err := sdl.CreateWindow("CHIP-8 Emulator", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 320, sdl.WINDOW_UTILITY|sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	// Create a renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		panic(err)
	}

	// Set the draw color to black
	renderer.SetDrawColor(0, 0, 0, 255)

	// Clear the window with the draw color
	renderer.Clear()

	// Update the window
	renderer.Present()

	return renderer
}

func drawGraphics(c *chip8.Chip8, renderer *sdl.Renderer) {
	// for i, pix := range c.Gfx {
	// 	if i != 0 && i%64 == 0 {
	// 		fmt.Println()
	// 	}
	// 	if pix == 1 {
	// 		fmt.Print("X")
	// 	} else {
	// 		fmt.Print(" ")
	// 	}
	// }
	// fmt.Println()

	// Draw the pixels as white rectangles on the window
	for i, pix := range c.Gfx {
		renderer.SetDrawColor(0, 0, 0, 255)
		if pix == 1 {
			renderer.SetDrawColor(255, 255, 255, 255)
		}
		renderer.FillRect(&sdl.Rect{
			X: int32(i%64) * 10,
			Y: int32(i/64) * 10,
			W: 10,
			H: 10,
		})
	}

	// Update the window
	renderer.Present()
}

func playBeep() {
	fmt.Println("BEEP")
}

func setInput(c *chip8.Chip8) {
	// Set the key press state (Press and Release)
	sdlevents := sdl.PollEvent()
	switch t := sdlevents.(type) {
	case *sdl.KeyboardEvent:
		if t.Type == sdl.KEYDOWN {
			switch t.Keysym.Sym {
			case sdl.K_1:
				c.Key[0x1] = 1
			case sdl.K_2:
				c.Key[0x2] = 1
			case sdl.K_3:
				c.Key[0x3] = 1
			case sdl.K_4:
				c.Key[0xC] = 1
			case sdl.K_q:
				c.Key[0x4] = 1
			case sdl.K_w:
				c.Key[0x5] = 1
			case sdl.K_e:
				c.Key[0x6] = 1
			case sdl.K_r:
				c.Key[0xD] = 1
			case sdl.K_a:
				c.Key[0x7] = 1
			case sdl.K_s:
				c.Key[0x8] = 1
			case sdl.K_d:
				c.Key[0x9] = 1
			case sdl.K_f:
				c.Key[0xE] = 1
			case sdl.K_z:
				c.Key[0xA] = 1
			case sdl.K_x:
				c.Key[0x0] = 1
			case sdl.K_c:
				c.Key[0xB] = 1
			case sdl.K_v:
				c.Key[0xF] = 1
			}
		} else if t.Type == sdl.KEYUP {
			switch t.Keysym.Sym {
			case sdl.K_1:
				c.Key[0x1] = 0
			case sdl.K_2:
				c.Key[0x2] = 0
			case sdl.K_3:
				c.Key[0x3] = 0
			case sdl.K_4:
				c.Key[0xC] = 0
			case sdl.K_q:
				c.Key[0x4] = 0
			case sdl.K_w:
				c.Key[0x5] = 0
			case sdl.K_e:
				c.Key[0x6] = 0
			case sdl.K_r:
				c.Key[0xD] = 0
			case sdl.K_a:
				c.Key[0x7] = 0
			case sdl.K_s:
				c.Key[0x8] = 0
			case sdl.K_d:
				c.Key[0x9] = 0
			case sdl.K_f:
				c.Key[0xE] = 0
			case sdl.K_z:
				c.Key[0xA] = 0
			case sdl.K_x:
				c.Key[0x0] = 0
			case sdl.K_c:
				c.Key[0xB] = 0
			case sdl.K_v:
				c.Key[0xF] = 0
			}
		}
	}
}

func main() {
	fmt.Println("Launching CHIP-8 Emulator...")

	initInput()
	c := chip8.Init()
	c.LoadRom(os.Args[1])

	renderer := initGraphics()
	defer renderer.Destroy()
	defer sdl.Quit()

	// spew.Dump(c)
	for {
		c.ExecuteOP()

		if c.DrawFlag {
			drawGraphics(c, renderer)
			c.DrawFlag = false
		}

		if c.PlaySound {
			playBeep()
		}

		setInput(c)

		if c.Stop {
			break
		}
		// Run at 60Hz
		sdl.Delay(1000 / 60)
	}
}
