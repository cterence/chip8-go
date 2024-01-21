package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/cterence/chip8-go/chip8"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chip8 *chip8.Chip8
	keys  [16]ebiten.Key
}

func (g *Game) Update() error {
	g.chip8.ExecuteOP()

	for i, key := range g.keys {
		if ebiten.IsKeyPressed(key) {
			g.chip8.Key[i] = 1
		} else {
			g.chip8.Key[i] = 0
		}
	}

	if g.chip8.PlaySound {
		playBeep()
		g.chip8.SoundTimer = 0
	}

	if g.chip8.Stop {
		os.Exit(0)
	}

	if g.chip8.DelayTimer > 0 {
		time.Sleep(time.Duration(g.chip8.DelayTimer) * time.Millisecond)
		g.chip8.DelayTimer = 0
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i, pix := range g.chip8.Gfx {
		if pix == 1 {
			screen.Set(int(i%64), int(i/64), color.White)
		} else {
			screen.Set(int(i%64), int(i/64), color.Black)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 64, 32
}

func playBeep() {
	// Play beep sound
	fmt.Println("BEEP")
}

func main() {
	fmt.Println("Launching CHIP-8 Emulator...")

	ebiten.SetWindowSize(640, 320)
	ebiten.SetWindowTitle("CHIP-8 Emulator")
	game := &Game{}

	c := chip8.Init()
	c.LoadRom(os.Args[1])

	game.chip8 = c

	game.keys = [16]ebiten.Key{
		ebiten.Key1,
		ebiten.Key2,
		ebiten.Key3,
		ebiten.Key4,
		ebiten.KeyQ,
		ebiten.KeyW,
		ebiten.KeyE,
		ebiten.KeyR,
		ebiten.KeyA,
		ebiten.KeyS,
		ebiten.KeyD,
		ebiten.KeyF,
		ebiten.KeyZ,
		ebiten.KeyX,
		ebiten.KeyC,
		ebiten.KeyV,
	}

	ebiten.SetTPS(500)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
