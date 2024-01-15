package chip8

import (
	"fmt"
	"log"
	"os"
)

type Chip8 struct {
	DrawFlag  bool
	PlaySound bool
	Stop      bool
	// Program counter
	pc     uint16
	opcode uint16
	// Index register
	i      uint16
	stack  [16]uint16
	sp     uint16
	memory [4096]uint8
	// Register list
	v   [16]uint8
	gfx [64 * 32]uint8
	// Current state of key
	key        [16]uint8
	delayTimer uint8
	soundTimer uint8
}

var chip8Fontset = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func Init() *Chip8 {
	c := Chip8{
		DrawFlag:   false,
		PlaySound:  false,
		Stop:       false,
		pc:         0x200,
		opcode:     0,
		i:          0,
		stack:      [16]uint16{},
		sp:         0,
		memory:     [4096]uint8{},
		v:          [16]uint8{},
		gfx:        [64 * 32]uint8{},
		key:        [16]uint8{},
		delayTimer: 0,
		soundTimer: 0,
	}

	// Clear display
	// Clear stack
	// Clear registers V0-VF
	// Clear memory

	// Load fontset
	copy(c.memory[:], chip8Fontset[:])

	// Initialize registers and memory once
	return &c
}

func (c *Chip8) LoadRom() {
	// Load game into memory
	buf, err := os.ReadFile("roms/space-invaders.ch8")
	if err != nil {
		log.Fatal("Error reading file", err)
	}

	copy(c.memory[512:], buf[:])
}

func (c *Chip8) ExecuteOP() {
	// Fetch Opcode
	c.opcode = uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])

	// Mask the opcode to only know its prefix, which tells what to do
	switch c.opcode & 0xF000 {

	case 0x0000:
		switch c.opcode & 0x000F {

		default:
			log.Fatalf("Unknown opcode: 0x%X", c.opcode)
			c.Stop = true
		}

	case 0x1000:
		c.pc = c.opcode & 0x0FFF
		fmt.Printf("[0x1NNN] Jump to 0x%X\n", c.pc)

	// case 0x4000:
	// 	registryIndex := c.opcode & 0x0F00 >> 8
	// 	fmt.Println("0x", registryIndex)

	case 0x6000:
		c.v[c.opcode&0x0F00>>8] = uint8(c.opcode & 0x00FF)
		fmt.Printf("[0x6XNN] Set v[%X] to 0x%X\n", c.opcode&0x0F00>>8, uint8(c.opcode&0x00FF))
		c.pc += 2

	case 0xA000:
		c.i = c.opcode & 0x0FFF
		fmt.Printf("[0xANNN] Set i to 0x%X\n", c.i)
		c.pc += 2

	case 0xD000:
		x := c.opcode % 0x0F00 >> 8
		y := c.opcode % 0x00F0 >> 4
		height := c.opcode % 0x000F
		var pixel uint8

		c.v[0xF] = 0

		for yline := uint16(0); yline < height; yline++ {
			pixel = c.memory[c.i+yline]
			for xline := uint16(0); xline < 16; xline++ {
				if pixel&(0x80>>xline) != 0 {
					if c.gfx[(x+xline+(y+yline)*64)] == 1 {
						c.v[0xF] = 1
					}
					c.gfx[(x + xline + (y+yline)*64)] ^= 1
				}
			}

		}

		c.DrawFlag = true
		fmt.Printf("[0xDNNN] Draw sprite at (0x%X, 0x%X) with height 0x%X starting from memory 0x%X\n", x, y, height, c.i)
		c.pc += 2

	default:
		log.Fatalf("Unknown opcode: 0x%X", c.opcode)
		c.Stop = true
	}

	// Update timers
}
