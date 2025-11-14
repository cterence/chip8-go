package cpu

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cterence/chip8-go/internal/chip8/components/memory"
	"github.com/cterence/chip8-go/internal/chip8/components/timer"
	"github.com/cterence/chip8-go/internal/chip8/components/ui"
	"github.com/cterence/chip8-go/internal/lib"
)

type register struct {
	value byte
}

type CPU struct {
	pressedKey byte

	mem   *memory.Memory
	ui    *ui.UI
	timer *timer.Timer

	reg [REGISTER_COUNT]register
	pc  uint16
	i   uint16

	stack [STACK_SIZE]uint16
	sp    uint8

	debugInfo debugInfo
}

type debugInfo struct {
	inst string
}

const (
	REGISTER_COUNT     byte   = 16
	STACK_SIZE         byte   = 16
	ADDR_MASK          uint16 = 0xFFF
	TPS                       = 660
	TARGET_TICK_PERIOD        = time.Second / TPS
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

	c.i = 0
	c.pc = memory.PROGRAM_RAM_START
	c.sp = 0
	c.pressedKey = 0xFF
}

func (c *CPU) Tick() {
	inst := c.decodeInstruction()
	c.execute(inst)
}

func (c *CPU) DebugInfo() string {
	var debugInfo strings.Builder

	debugInfo.WriteString(fmt.Sprintf("OP: %-13s ", c.debugInfo.inst))
	debugInfo.WriteString("PC:" + lib.FormatHex(c.pc, 4) + " ")
	debugInfo.WriteString("SP:" + lib.FormatHex(c.sp, 2) + " ")
	debugInfo.WriteString("I:" + lib.FormatHex(c.i, 4) + " ")
	debugInfo.WriteString("ST:" + lib.FormatHex(c.stack[c.sp], 4) + " ")

	for r, v := range c.reg {
		rs, vs := lib.FormatHex(byte(r), 1), lib.FormatHex(v.value, 2)
		debugInfo.WriteString("V" + rs + ":" + vs)

		if r < len(c.reg)-1 {
			debugInfo.WriteString(" ")
		}
	}

	return debugInfo.String()
}

func (c *CPU) readReg(reg byte) byte {
	lib.Assert(reg < REGISTER_COUNT, fmt.Errorf("illegal read to register V%x", reg))
	v := c.reg[reg].value

	return v
}

func (c *CPU) writeReg(reg byte, v byte) {
	lib.Assert(reg < REGISTER_COUNT, fmt.Errorf("illegal write to register V%x", reg))

	c.reg[reg].value = v
}

func (c *CPU) pushStack(v uint16) {
	c.stack[c.sp] = v
	c.sp++
}

func (c *CPU) popStack() uint16 {
	c.sp--
	v := c.stack[c.sp]

	return v
}

func (c *CPU) decodeInstruction() uint16 {
	lib.Assert(c.pc < memory.PROGRAM_RAM_END, fmt.Errorf("illegal program counter position: 0x%03X", c.pc))

	hi := uint16(c.mem.Read(c.pc)) << lib.BYTE_SIZE
	lo := uint16(c.mem.Read(c.pc + 1))

	return hi | lo
}

func (c *CPU) execute(inst uint16) {
	implemented := true

	hi := byte(inst >> uint16(lib.BYTE_SIZE))
	lo := byte(inst)

	lo0, lo1, hi0, hi1 := lo&0xF, (lo>>4)&0xF, hi&0xF, (hi>>4)&0xF

	switch hi1 {
	case 0x0:
		switch lo0 {
		case 0x0:
			c.debugInfo.inst = "CLS"
			c.ui.Reset()
		case 0xE:
			c.debugInfo.inst = "RET"
			c.pc = c.popStack()
		default:
			implemented = false
		}
	case 0x1:
		v := inst & ADDR_MASK
		c.debugInfo.inst = "JP " + lib.FormatHex(v, 3)

		c.pc = v

		return
	case 0x2:
		v := inst & ADDR_MASK
		c.debugInfo.inst = "CALL " + lib.FormatHex(v, 3)
		c.pushStack(c.pc)
		c.pc = v

		return
	case 0x3:
		c.debugInfo.inst = "SE V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)

		if c.readReg(hi0) == lo {
			c.pc += 2
		}
	case 0x4:
		c.debugInfo.inst = "SNE V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)

		if c.readReg(hi0) != lo {
			c.pc += 2
		}
	case 0x5:
		c.debugInfo.inst = "SE V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)

		if c.readReg(hi0) == c.readReg(lo1) {
			c.pc += 2
		}
	case 0x6:
		c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)
		c.writeReg(hi0, lo)
	case 0x7:
		c.debugInfo.inst = "ADD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)
		c.writeReg(hi0, c.readReg(hi0)+lo)
	case 0x8:
		switch lo0 {
		case 0x0:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.writeReg(hi0, c.readReg(lo1))
		case 0x1:
			c.debugInfo.inst = "OR V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.writeReg(hi0, c.readReg(hi0)|c.readReg(lo1))
			c.writeReg(0xF, 0)
		case 0x2:
			c.debugInfo.inst = "AND V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.writeReg(hi0, c.readReg(hi0)&c.readReg(lo1))
			c.writeReg(0xF, 0)
		case 0x3:
			c.debugInfo.inst = "XOR V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.writeReg(hi0, c.readReg(hi0)^c.readReg(lo1))
			c.writeReg(0xF, 0)
		case 0x4:
			c.debugInfo.inst = "ADD V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			a, b := c.readReg(hi0), c.readReg(lo1)
			v := uint16(a) + uint16(b)

			c.writeReg(hi0, byte(v))

			if v > 0xFF {
				c.writeReg(0xF, 1)
			} else {
				c.writeReg(0xF, 0)
			}
		case 0x5:
			c.debugInfo.inst = "SUB V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""
			a, b := c.readReg(hi0), c.readReg(lo1)
			v := a - b

			c.writeReg(hi0, v)

			if a >= b {
				c.writeReg(0xF, 1)
			} else {
				c.writeReg(0xF, 0)
			}
		case 0x6:
			c.debugInfo.inst = "SHR V" + lib.FormatHex(hi0, 1) + " {, V" + lib.FormatHex(lo1, 1) + "}"
			v := c.readReg(lo1)

			c.writeReg(hi0, v>>1)
			c.writeReg(0xF, lib.Bit(v, 0))
		case 0x7:
			c.debugInfo.inst = "SUBN V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""
			v := c.readReg(lo1) - c.readReg(hi0)
			c.writeReg(hi0, v)

			if c.readReg(hi0) > c.readReg(lo1) {
				c.writeReg(0xF, 0)
			} else {
				c.writeReg(0xF, 1)
			}
		case 0xE:
			c.debugInfo.inst = "SHL V" + lib.FormatHex(hi0, 1) + " {, V" + lib.FormatHex(lo1, 1) + "}"
			v := c.readReg(lo1)

			c.writeReg(hi0, v<<1)
			c.writeReg(0xF, lib.Bit(v, 7))
		default:
			implemented = false
		}
	case 0x9:
		c.debugInfo.inst = "SNE V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""

		if c.readReg(hi0) != c.readReg(lo1) {
			c.pc += 2
		}
	case 0xA:
		v := inst & ADDR_MASK
		c.debugInfo.inst = "LD I, " + lib.FormatHex(v, 3)
		c.i = v
	case 0xB:
		v := inst & ADDR_MASK
		c.debugInfo.inst = "JP V0, " + lib.FormatHex(v, 3)
		c.pc = v + uint16(c.readReg(0))

		return
	case 0xC:
		c.debugInfo.inst = "RND V" + lib.FormatHex(hi0, 1) + ", byte"

		r := byte(rand.Intn(0xFF))
		v := r & lo
		c.writeReg(hi0, v)
	case 0xD:
		c.debugInfo.inst = "DRW V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ", " + lib.FormatHex(lo0, 1)
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
			c.debugInfo.inst = "SKP V" + lib.FormatHex(hi0, 1)

			if c.ui.IsKeyPressed(c.readReg(hi0)) {
				c.pc += 2
			}
		case 0xA1:
			c.debugInfo.inst = "SKNP V" + lib.FormatHex(hi0, 1)

			if !c.ui.IsKeyPressed(c.readReg(hi0)) {
				c.pc += 2
			}
		default:
			implemented = false
		}
	case 0xF:
		switch lo {
		case 0x07:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", DT"
			c.writeReg(hi0, c.timer.GetDelay())
		case 0x0A:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", K"

			if c.pressedKey == 0xFF {
				c.pressedKey = c.ui.GetPressedKey()

				return
			}

			if c.ui.IsKeyPressed(c.pressedKey) {
				return
			}

			c.writeReg(hi0, c.pressedKey)
			c.pressedKey = 0xFF
		case 0x15:
			c.debugInfo.inst = "LD DT, V" + lib.FormatHex(hi0, 1)
			c.timer.SetDelay(c.readReg(hi0))
		case 0x18:
			c.debugInfo.inst = "LD ST, V" + lib.FormatHex(hi0, 1)
			c.timer.SetSound(c.readReg(hi0))
		case 0x1E:
			c.debugInfo.inst = "ADD I, V" + lib.FormatHex(hi0, 1)
			c.i = c.i + uint16(c.readReg(hi0))
		case 0x29:
			c.debugInfo.inst = "LF F, V" + lib.FormatHex(hi0, 1)
			digit := c.readReg(hi0)
			c.i = uint16(digit * 5)
		case 0x33:
			c.debugInfo.inst = "LD B, V" + lib.FormatHex(hi0, 1)
			v := fmt.Sprintf("%03d", c.readReg(hi0))
			c.mem.Write(c.i, v[0]-'0')
			c.mem.Write(c.i+1, v[1]-'0')
			c.mem.Write(c.i+2, v[2]-'0')
		case 0x55:
			c.debugInfo.inst = "LD " + lib.FormatHex(c.i, 2) + ", V" + lib.FormatHex(hi0, 1)

			for x := range hi0 + 1 {
				c.mem.Write(c.i, c.readReg(x))
				c.i++
			}
		case 0x65:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(c.i, 2)

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

	c.pc += 2

	lib.Assert(implemented, fmt.Errorf("unimplemented instruction: %04X", inst))
}
