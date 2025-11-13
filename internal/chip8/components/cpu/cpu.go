package cpu

import (
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/cterence/chip8-go-v2/internal/chip8/components/memory"
	"github.com/cterence/chip8-go-v2/internal/chip8/components/timer"
	"github.com/cterence/chip8-go-v2/internal/chip8/components/ui"
	"github.com/cterence/chip8-go-v2/internal/lib"
)

type register struct {
	value byte
}

type CPU struct {
	Paused     bool
	pressedKey byte

	mem   *memory.Memory
	ui    *ui.UI
	timer *timer.Timer

	reg [REGISTER_COUNT]register
	pc  uint16
	i   uint16

	stack [STACK_SIZE]uint16
	sp    uint8
}

const (
	REGISTER_COUNT byte   = 16
	STACK_SIZE     byte   = 16
	ADDR_MASK      uint16 = 0xFFF
	SP_INIT        uint8  = 0xFF
)

func New(mem *memory.Memory, ui *ui.UI, t *timer.Timer) *CPU {
	c := &CPU{
		mem:   mem,
		ui:    ui,
		timer: t,
	}

	return c
}

func (c *CPU) Init() {
	for i := range c.reg {
		c.writeReg(byte(i), 0)
	}

	for i := range c.stack {
		c.stack[i] = 0
	}

	c.Paused = false
	c.i = 0
	c.pc = memory.PROGRAM_RAM_START
	c.sp = SP_INIT
	c.pressedKey = 0xFF
}

func (c *CPU) Tick() {
	if !c.Paused {
		inst := c.decodeInstruction()
		c.execute(inst)
	}
}

func (c *CPU) decodeInstruction() uint16 {
	lib.Assert(c.pc < memory.PROGRAM_RAM_END, fmt.Errorf("illegal program counter position: 0x%03X", c.pc))

	hi := uint16(c.mem.Read(c.pc)) << lib.BYTE_SIZE
	c.pc++
	lo := uint16(c.mem.Read(c.pc))
	c.pc++

	return hi | lo
}

func (c *CPU) readReg(reg byte) byte {
	lib.Assert(reg < REGISTER_COUNT, fmt.Errorf("illegal read to register V%x", reg))
	v := c.reg[reg].value
	slog.Debug("writeReg", "reg", lib.FormatHex(reg, 1), "v", lib.FormatHex(v, 2))

	return v
}

func (c *CPU) writeReg(reg byte, v byte) {
	lib.Assert(reg < REGISTER_COUNT, fmt.Errorf("illegal write to register V%x", reg))
	slog.Debug("writeReg", "reg", lib.FormatHex(reg, 1), "v", lib.FormatHex(v, 2))

	c.reg[reg].value = v
}

func (c *CPU) pushStack(v uint16) {
	c.sp++
	c.stack[c.sp] = v
	slog.Debug("pushStack", "v", lib.FormatHex(v, 2), "sp", lib.FormatHex(c.sp, 4))
}

func (c *CPU) popStack() uint16 {
	v := c.stack[c.sp]
	c.sp--
	slog.Debug("popStack", "v", lib.FormatHex(v, 2), "sp", lib.FormatHex(c.sp, 4))

	return v
}

func (c *CPU) execute(inst uint16) {
	var (
		name string
	)

	implemented := true

	hi := byte(inst >> uint16(lib.BYTE_SIZE))
	lo := byte(inst)

	lo0, lo1, hi0, hi1 := lo&0xF, (lo>>4)&0xF, hi&0xF, (hi>>4)&0xF

	switch hi1 {
	case 0x0:
		switch lo0 {
		case 0x0:
			name = "CLS"
			debugLog(name, inst, c.pc)
			c.ui.Reset()
		case 0xE:
			name = "RET"
			debugLog(name, inst, c.pc)
			c.pc = c.popStack()
		default:
			implemented = false
		}
	case 0x1:
		v := inst & ADDR_MASK
		name = "JP " + lib.FormatHex(v, 3)
		debugLog(name, inst, c.pc)

		c.pc = v
	case 0x2:
		v := inst & ADDR_MASK
		name = "CALL " + lib.FormatHex(v, 3)
		debugLog(name, inst, c.pc)
		c.pushStack(c.pc)
		c.pc = v
	case 0x3:
		name = "SE V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)
		debugLog(name, inst, c.pc)

		if c.readReg(hi0) == lo {
			c.pc += 2
		}
	case 0x4:
		name = "SNE V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)
		debugLog(name, inst, c.pc)

		if c.readReg(hi0) != lo {
			c.pc += 2
		}
	case 0x5:
		name = "SE V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
		debugLog(name, inst, c.pc)

		if c.readReg(hi0) == c.readReg(lo1) {
			c.pc += 2
		}
	case 0x6:
		name = "LD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)
		debugLog(name, inst, c.pc)
		c.writeReg(hi0, lo)
	case 0x7:
		name = "ADD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)
		debugLog(name, inst, c.pc)
		c.writeReg(hi0, c.readReg(hi0)+lo)
	case 0x8:
		switch lo0 {
		case 0x0:
			name = "LD V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			debugLog(name, inst, c.pc)
			c.writeReg(hi0, c.readReg(lo1))
		case 0x1:
			name = "OR V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			debugLog(name, inst, c.pc)
			c.writeReg(hi0, c.readReg(hi0)|c.readReg(lo1))
			c.writeReg(0xF, 0)
		case 0x2:
			name = "AND V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			debugLog(name, inst, c.pc)
			c.writeReg(hi0, c.readReg(hi0)&c.readReg(lo1))
			c.writeReg(0xF, 0)
		case 0x3:
			name = "XOR V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			debugLog(name, inst, c.pc)
			c.writeReg(hi0, c.readReg(hi0)^c.readReg(lo1))
			c.writeReg(0xF, 0)
		case 0x4:
			name = "ADD V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			debugLog(name, inst, c.pc)
			a, b := c.readReg(hi0), c.readReg(lo1)
			v := uint16(a) + uint16(b)

			c.writeReg(hi0, byte(v))

			if v > 0xFF {
				c.writeReg(0xF, 1)
			} else {
				c.writeReg(0xF, 0)
			}
		case 0x5:
			name = "SUB V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""
			debugLog(name, inst, c.pc)
			a, b := c.readReg(hi0), c.readReg(lo1)
			v := a - b

			c.writeReg(hi0, v)

			if a >= b {
				c.writeReg(0xF, 1)
			} else {
				c.writeReg(0xF, 0)
			}
		case 0x6:
			name = "SHR V" + lib.FormatHex(hi0, 1) + " {, V" + lib.FormatHex(lo1, 1) + "}"
			debugLog(name, inst, c.pc)
			v := c.readReg(lo1)

			c.writeReg(hi0, v>>1)
			c.writeReg(0xF, lib.Bit(v, 0))
		case 0x7:
			name = "SUBN V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""
			debugLog(name, inst, c.pc)
			v := c.readReg(lo1) - c.readReg(hi0)
			c.writeReg(hi0, v)

			if c.readReg(hi0) > c.readReg(lo1) {
				c.writeReg(0xF, 0)
			} else {
				c.writeReg(0xF, 1)
			}
		case 0xE:
			name = "SHL V" + lib.FormatHex(hi0, 1) + " {, V" + lib.FormatHex(lo1, 1) + "}"
			debugLog(name, inst, c.pc)
			v := c.readReg(lo1)

			c.writeReg(hi0, v<<1)
			c.writeReg(0xF, lib.Bit(v, 7))
		default:
			implemented = false
		}
	case 0x9:
		name = "SNE V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""
		debugLog(name, inst, c.pc)

		if c.readReg(hi0) != c.readReg(lo1) {
			c.pc += 2
		}
	case 0xA:
		v := inst & ADDR_MASK
		name = "LD I, " + lib.FormatHex(v, 3)
		debugLog(name, inst, c.pc)
		slog.Debug("writeReg", "reg", "i", "v", lib.FormatHex(v, 3))
		c.i = v
	case 0xB:
		v := inst & ADDR_MASK
		name = "JP V0, " + lib.FormatHex(v, 3)
		debugLog(name, inst, c.pc)
		c.pc = v + uint16(c.readReg(0))
	case 0xC:
		name = "RND V" + lib.FormatHex(hi0, 1) + ", byte"
		debugLog(name, inst, c.pc)

		r := byte(rand.Intn(0xFF))
		v := r & lo
		c.writeReg(hi0, v)
	case 0xD:
		name = "DRW V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ", " + lib.FormatHex(lo0, 1)
		debugLog(name, inst, c.pc)
		x := c.readReg(hi0)
		y := c.readReg(lo1)

		sprite := make([]byte, lo0)

		for i := range sprite {
			sprite[i] = c.mem.Read(c.i + uint16(i))
		}

		if c.ui.DrawSprite(x, y, sprite) {
			c.writeReg(0xF, 1)
		} else {
			c.writeReg(0xF, 0)
		}
	case 0xE:
		switch lo {
		case 0x9E:
			name = "SKP V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)

			if c.ui.IsKeyPressed(c.readReg(hi0)) {
				c.pc += 2
			}
		case 0xA1:
			name = "SKNP V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)

			if !c.ui.IsKeyPressed(c.readReg(hi0)) {
				c.pc += 2
			}
		default:
			implemented = false
		}
	case 0xF:
		switch lo {
		case 0x07:
			name = "LD V" + lib.FormatHex(hi0, 1) + ", DT"
			debugLog(name, inst, c.pc)
			c.writeReg(hi0, c.timer.GetDelay())
		case 0x0A:
			name = "LD V" + lib.FormatHex(hi0, 1) + ", K"
			debugLog(name, inst, c.pc)

			if c.pressedKey == 0xFF {
				c.pressedKey = c.ui.GetPressedKey()
				c.pc -= 2

				return
			}

			if c.ui.IsKeyPressed(c.pressedKey) {
				c.pc -= 2

				return
			}

			c.writeReg(hi0, c.pressedKey)
			c.pressedKey = 0xFF
		case 0x15:
			name = "LD DT, V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)
			c.timer.SetDelay(c.readReg(hi0))
		case 0x18:
			name = "LD ST, V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)
			c.timer.SetSound(c.readReg(hi0))
		case 0x1E:
			name = "ADD I, V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)
			c.i = c.i + uint16(c.readReg(hi0))
		case 0x29:
			name = "LF F, V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)
			digit := c.readReg(hi0)
			c.i = uint16(digit * 5)
		case 0x33:
			name = "LD B, V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)
			v := fmt.Sprintf("%03d", c.readReg(hi0))
			c.mem.Write(c.i, v[0]-'0')
			c.mem.Write(c.i+1, v[1]-'0')
			c.mem.Write(c.i+2, v[2]-'0')
		case 0x55:
			name = "LD " + lib.FormatHex(c.i, 2) + ", V" + lib.FormatHex(hi0, 1)
			debugLog(name, inst, c.pc)

			for x := range hi0 + 1 {
				c.mem.Write(c.i, c.readReg(x))
				c.i++
			}
		case 0x65:
			name = "LD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(c.i, 2)
			debugLog(name, inst, c.pc)

			for x := range hi0 + 1 {
				c.writeReg(x, c.mem.Read(c.i))
				c.i++
			}
		default:
			implemented = false
		}
	default:
		implemented = false
	}

	lib.Assert(implemented, fmt.Errorf("unimplemented instruction: %04X", inst))
}

func debugLog(name string, inst, pc uint16) {
	slog.Debug("execute", "name", name, "inst", lib.FormatHex(inst, 4), "pc", lib.FormatHex(pc, 4))
}
