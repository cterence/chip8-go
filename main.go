package main

import (
	"fmt"

	"github.com/cterence/chip8-go/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

func initInput() {
	fmt.Println("Init input : Not implemented yet")
}

func initGraphics() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	rect := sdl.Rect{0, 0, 200, 200}
	colour := sdl.Color{R: 255, G: 0, B: 255, A: 255} // purple
	pixel := sdl.MapRGBA(surface.Format, colour.R, colour.G, colour.B, colour.A)
	surface.FillRect(&rect, pixel)
	window.UpdateSurface()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

func drawGraphics() {
	fmt.Println("Draw Graphics : Not implemented yet")
}

func playBeep() {
	fmt.Println("Play Beep : Not implemented yet")
}

func setInput() {
	fmt.Println("Set Input : Not implemented yet")
}

func main() {
	fmt.Println("Launching CHIP-8 Emulator...")

	initInput()
	initGraphics()

	c := chip8.Init()
	c.LoadRom()

	// spew.Dump(c)
	for {
		c.ExecuteOP()

		if c.DrawFlag {
			drawGraphics()
			c.DrawFlag = false
		}

		if c.PlaySound {
			playBeep()
		}

		setInput()

		if c.Stop {
			break
		}
	}
}
