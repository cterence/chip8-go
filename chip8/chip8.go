package chip8

import (
	"fmt"
	"log"
	"math"
	"os"
)

type Chip8 struct {
	DrawFlag  bool
	PlaySound bool
	Stop      bool
	// Program counter
	pc uint16
	// Index register
	i      uint16
	stack  []uint16
	sp     uint16
	memory []uint8
	// Register list
	v   []uint8
	Gfx []uint8
	// Current state of Key
	Key        []uint8
	DelayTimer uint8
	SoundTimer uint8
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
		i:          0,
		stack:      make([]uint16, 16),
		sp:         0,
		memory:     make([]uint8, 4096),
		v:          make([]uint8, 16),
		Gfx:        make([]uint8, 64*32),
		Key:        make([]uint8, 16),
		DelayTimer: 0,
		SoundTimer: 0,
	}

	// Clear display
	for i := range c.Gfx {
		c.Gfx[i] = 0
	}
	// Clear stack
	// Clear registers V0-VF
	// Clear memory

	// Load fontset
	copy(c.memory[:], chip8Fontset[:])

	// Initialize registers and memory once
	return &c
}

func (c *Chip8) LoadRom(filePath string) {
	// Load game into memory
	buf, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("ERROR: cannot read file", err)
	}

	copy(c.memory[0x200:], buf[:])
}

func (c *Chip8) ExecuteOP() {
	// Fetch Opcode
	op := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])

	// Mask the opcode to only know its prefix, which tells what to do
	switch op & 0xF000 {

	case 0x0000:
		switch op & 0x000F {

		case 0x0000:
			fmt.Println("[0x00E0] Clear the display")
			for i := range c.Gfx {
				c.Gfx[i] = 0
			}
			c.pc += 2

		case 0x000E:
			c.pc = c.stack[c.sp-1]
			c.sp -= 1
			fmt.Printf("[0x00EE] Return to instruction on top of stack : 0x%X\n", c.pc)
			c.pc += 2

		default:
			log.Fatalf("ERROR: Unknown opcode: 0x%X", op)
			c.Stop = true
		}

	case 0x1000:
		c.pc = op & 0x0FFF
		fmt.Printf("[0x1NNN] Jump to 0x%X\n", c.pc)

	case 0x2000:
		if c.sp == 0xF {
			log.Fatal("ERROR: Trying to append to a full stack")
		}
		c.stack[c.sp] = c.pc
		c.sp++
		fmt.Printf("[0x2NNN] Save PC (0x%X) to stack (size: %X) and jump to 0x%X\n", c.pc, c.sp, op&0x0FFF)
		c.pc = op & 0x0FFF

	case 0x3000:
		x := op & 0x0F00 >> 8
		n := uint8(op & 0x00FF)
		c.pc += 2
		fmt.Printf("[0x3XNN] Skip next instruction if v%X(0x%X) eq NN(0x%X)\n", c.v[x], x, n)
		if c.v[x] == n {
			c.pc += 2
		}

	case 0x4000:
		x := op & 0x0F00 >> 8
		n := uint8(op & 0x00FF)
		fmt.Printf("[0x4XNN] Skip instruction if v%X=%X != %X\n", x, c.v[x], n)
		if c.v[x] != n {
			c.pc += 2
		}
		c.pc += 2

	case 0x5000:
		x := op & 0x0F00 >> 8
		y := op & 0x00F0 >> 4
		vx := c.v[x]
		vy := c.v[y]
		fmt.Printf("[0x5XY0] Skip next instruction if v%X == v%X : %X = %X\n", x, y, vx, vy)

		if vx == vy {
			c.pc += 2
		}
		c.pc += 2

	case 0x6000:
		x := op & 0x0F00 >> 8
		val := uint8(op & 0x00FF)
		fmt.Printf("[0x6XNN] Set v%X to 0x%X\n", x, val)
		c.v[x] = val
		c.pc += 2

	case 0x7000:
		x := op & 0x0F00 >> 8
		val := uint8(op & 0x00FF)
		fmt.Printf("[0x7XNN] Add 0x%X to v%X\n", val, x)
		c.v[x] += val
		c.pc += 2

	case 0x8000:
		switch op & 0x000F {

		case 0x0000:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			fmt.Printf("[0x8XY0] v%X = %X\n", x, c.v[x])
			c.v[x] = c.v[y]
			c.pc += 2

		case 0x0001:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			val := c.v[x] | c.v[y]
			fmt.Printf("[0x8XY1] v%X = 0x%X | 0x%X = 0x%X\n", x, c.v[x], c.v[y], val)
			c.v[x] = val
			c.pc += 2

		case 0x0002:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			val := c.v[x] & c.v[y]
			fmt.Printf("[0x8XY1] v%X = 0x%X & 0x%X = 0x%X\n", x, c.v[x], c.v[y], val)
			c.v[x] = val
			c.pc += 2

		case 0x0003:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			val := c.v[x] ^ c.v[y]
			fmt.Printf("[0x8XY1] v%X = 0x%X ^ 0x%X = 0x%X\n", x, c.v[x], c.v[y], val)
			c.v[x] = val
			c.pc += 2

		case 0x0004:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			res := c.v[x] + c.v[y]
			fmt.Printf("[0x8XY4] v%X = v%X + v%X = %X + %X = %X (vF = 1 if > 0xFF)\n", x, x, y, c.v[x], c.v[y], res)
			if res > 0xFF {
				c.v[0xF] = 1
				res = res % 0xFF
			}
			c.v[x] = res
			c.pc += 2

		case 0x0005:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			res := uint8(math.Abs(float64(c.v[x] - c.v[y])))
			if c.v[x] > c.v[y] {
				c.v[0xF] = 1
			}
			fmt.Printf("[0x8XY7] v%X = v%X - v%X = %X - %X = %X (vF = 1 if > 0xFF)\n", x, x, y, c.v[x], c.v[y], res)
			c.v[x] = res
			c.pc += 2

		case 0x0006:
			x := op & 0x0F00 >> 8
			lsb := c.v[x] & 0x0F
			if lsb == 1 {
				c.v[0xF] = 1
			}
			c.v[x] /= 2
			c.pc += 2

		case 0x0007:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4
			res := uint8(math.Abs(float64(c.v[y] - c.v[x])))
			if c.v[y] > c.v[x] {
				c.v[0xF] = 1
			}
			fmt.Printf("[0x8XY7] v%X = v%X - v%X = %X - %X = %X (vF = 1 if > 0xFF)\n", x, y, x, c.v[y], c.v[x], res)
			c.v[x] = res
			c.pc += 2

		case 0x000E:
			x := op & 0x0F00 >> 8
			msb := c.v[x] >> 4
			fmt.Printf("[0x000E] If MSB of v%X = 1, vF = 1 & v%X*2 = %X\n", x, x, c.v[x]*2)
			if msb == 1 {
				c.v[0xF] = 1
			}
			c.v[x] *= 2
			c.pc += 2

		default:
			log.Fatalf("ERROR: Unknown opcode: 0x%X", op)
			c.Stop = true
		}

	case 0x9000:
		x := op & 0x0F00 >> 8
		y := op & 0x00F0 >> 4
		vx := c.v[x]
		vy := c.v[y]
		fmt.Printf("[0x9XY0] Skip next instruction if v%X != v%X : %X = %X\n", x, y, vx, vy)

		if vx != vy {
			c.pc += 2
		}
		c.pc += 2

	case 0xA000:
		c.i = op & 0x0FFF
		fmt.Printf("[0xANNN] Set I to 0x%X\n", c.i)
		c.pc += 2

	case 0xD000:
		x := op & 0x0F00 >> 8
		y := op & 0x00F0 >> 4
		vx := c.v[x]
		vy := c.v[y]
		height := op & 0x000F
		fmt.Printf("[0xDXYN] Draw sprite at v%X=%X, v%X=%X, height=%X\n", x, vx, y, vy, height)
		var pixel uint8

		c.v[0xF] = 0

		for yline := uint16(0); yline < height; yline++ {
			pixel = c.memory[c.i+yline]
			for xline := uint16(0); xline < 8; xline++ {
				if (pixel & (0x80 >> xline)) != 0 {
					if c.Gfx[(uint16(vx)+xline)+((uint16(vy)+yline)*64)] == 1 {
						c.v[0xF] = 1
					}
					c.Gfx[(uint16(vx)+xline)+((uint16(vy)+yline)*64)] ^= 1
				}
			}
		}

		c.DrawFlag = true
		c.pc += 2

	case 0xE000:
		switch op & 0x00FF {

		case 0x00A1:
			x := op & 0x0F00 >> 8
			fmt.Printf("[0xEXA1] Skip next instruction if key 0x%X is 0(0x%X)\n", x, c.Key[x])
			if c.Key[x] == 0 {
				c.pc += 2
			}
			c.pc += 2

		case 0x009E:
			x := op & 0x0F00 >> 8
			if c.Key[c.v[x]] == 1 {
				c.pc += 2
			}
			c.pc += 2

		default:
			log.Fatalf("ERROR: Unknown opcode: 0x%X", op)
			c.Stop = true
		}

	case 0xF000:
		switch op & 0x00FF {

		case 0x0007:
			x := op & 0x0F00 >> 8
			fmt.Printf("[0xFX07] Set v%X to delay timer = %X\n", x, c.DelayTimer)
			c.v[x] = c.DelayTimer
			c.pc += 2

		case 0x000A:
			// Wait for key press, store the value of the key in Vx
			x := op & 0x0F00 >> 8
			fmt.Printf("[0xFX0A] Wait for key press and store it in v%X\n", x)
			var keyPress bool
			for i := range c.Key {
				if c.Key[i] != 0 {
					c.v[x] = uint8(i)
					keyPress = true
				}
			}
			if !keyPress {
				return
			}
			c.pc += 2

		case 0x0015:
			x := op & 0x0F00 >> 8
			c.DelayTimer = c.v[x]
			fmt.Printf("[0xFX15] Set delay timer to 0x%X", c.v[x])
			c.pc += 2

		case 0x0018:
			x := op & 0x0F00 >> 8
			c.SoundTimer = c.v[x]
			fmt.Printf("[0xFX18] Set sound timer to 0x%X", c.v[x])
			c.pc += 2

		case 0x001E:
			x := op & 0x0F00 >> 8
			c.i += uint16(c.v[x])
			fmt.Printf("[0xFX1E] Add v%X=0x%X to I=0x%X\n", x, c.v[x], c.i)
			c.pc += 2

		case 0x0033:
			x := op & 0x0F00 >> 8
			fmt.Printf("[0xFX33] I = %X, I+1 = %X, I+2 = %X\n", (c.v[x]/100)%10, (c.v[x]/10)%10, (c.v[x]/1)%10)
			c.memory[c.i] = (c.v[x] / 100) % 10
			c.memory[c.i+1] = (c.v[x] / 10) % 10
			c.memory[c.i+2] = (c.v[x] / 1) % 10
			c.pc += 2

		case 0x0055:
			x := op & 0x0F00 >> 8
			fmt.Printf("[0xFX65] Load memory starting from I (0x%X) into v[0x0] to v%X\n", c.i, x)
			var idx uint16
			for idx = 0; idx <= x; idx++ {
				c.memory[c.i+idx] = c.v[idx]
			}
			c.pc += 2

		case 0x0065:
			x := op & 0x0F00 >> 8
			fmt.Printf("[0xFX65] Load memory starting from I (0x%X) into v[0x0] to v%X\n", c.i, x)
			var idx uint16
			for idx = 0; idx <= x; idx++ {
				c.v[idx] = c.memory[c.i+idx]
			}
			c.pc += 2

		default:
			log.Fatalf("ERROR: Unknown opcode: 0x%X", op)
			c.Stop = true
		}

	default:
		log.Fatalf("ERROR: Unknown opcode: 0x%X", op)
		c.Stop = true
	}

	// Update timers
}
