package cpu

import (
	"fmt"
	"log/slog"

	"github.com/cterence/chip8-go-v2/internal/chip8/components/memory"
	"github.com/cterence/chip8-go-v2/internal/chip8/components/ui"
	"github.com/cterence/chip8-go-v2/internal/lib"
)

type register struct {
	value byte
}

type CPU struct {
	Paused bool

	mem *memory.Memory
	ui  *ui.UI

	reg [REGISTER_COUNT]register
	pc  uint16
	i   uint16

	stack [STACK_SIZE]uint16
	sp    uint8
}

const (
	REGISTER_COUNT byte   = 16
	STACK_SIZE     byte   = 16
	INST_MASK      uint16 = 0xFFF
	SP_INIT        uint8  = 0xFF
)

func New(mem *memory.Memory, ui *ui.UI) *CPU {
	c := &CPU{
		mem: mem,
		ui:  ui,
	}

	return c
}

func (c *CPU) Init() {
	c.pc = memory.PROGRAM_RAM_START
	c.sp = SP_INIT
}

func (c *CPU) Tick() int {
	cycles := 0

	if !c.Paused {
		inst := c.decodeInstruction()
		cycles = c.execute(inst)
	}

	return cycles
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

func (c *CPU) execute(inst uint16) int {
	cycles, implemented, name := 0, true, ""

	hi := byte(inst >> uint16(lib.BYTE_SIZE))
	lo := byte(inst)

	lo0, lo1, hi0, hi1 := lo&0xF, (lo>>4)&0xF, hi&0xF, (hi>>4)&0xF

	switch hi1 {
	case 0x0:
		switch lo0 {
		case 0x0:
			name = "CLS"

			c.ui.Reset()
		case 0xE:
			name = "RET"
			c.pc = c.popStack()
		default:
			implemented = false
		}
	case 0x1:
		v := inst & INST_MASK
		name = "JP " + lib.FormatHex(v, 3)
		c.pc = v
	case 0x2:
		v := inst & INST_MASK
		name = "CALL " + lib.FormatHex(v, 3)

		c.pushStack(c.pc)
		c.pc = v
	case 0x3:
		name = fmt.Sprintf("SE V%s, %s", lib.FormatHex(hi0, 1), lib.FormatHex(lo, 2))
		if c.readReg(hi0) == lo {
			c.pc += 2
		}
	case 0x4:
		name = fmt.Sprintf("SNE V%s, %s", lib.FormatHex(hi0, 1), lib.FormatHex(lo, 2))
		if c.readReg(hi0) != lo {
			c.pc += 2
		}
	case 0x5:
		name = fmt.Sprintf("SE V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
		if c.readReg(hi0) == c.readReg(lo1) {
			c.pc += 2
		}
	case 0x6:
		name = fmt.Sprintf("LD V%s, %s", lib.FormatHex(hi0, 1), lib.FormatHex(lo, 2))
		c.writeReg(hi0, lo)
	case 0x7:
		name = fmt.Sprintf("ADD V%s, %s", lib.FormatHex(hi0, 1), lib.FormatHex(lo, 2))
		c.writeReg(hi0, c.readReg(hi0)+lo)
	case 0x8:
		switch lo0 {
		case 0x0:
			name = fmt.Sprintf("LD V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			c.writeReg(hi0, c.readReg(lo1))
		case 0x1:
			name = fmt.Sprintf("OR V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			c.writeReg(hi0, c.readReg(hi0)|c.readReg(lo1))
		case 0x2:
			name = fmt.Sprintf("AND V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			c.writeReg(hi0, c.readReg(hi0)&c.readReg(lo1))
		case 0x3:
			name = fmt.Sprintf("XOR V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			c.writeReg(hi0, c.readReg(hi0)^c.readReg(lo1))
		case 0x4:
			name = fmt.Sprintf("ADD V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			v := uint16(c.readReg(hi0)) + uint16(c.readReg(lo1))
			c.writeReg(hi0, byte(v))

			if v > 0xFF {
				c.writeReg(0xF, 1)
			} else {
				c.writeReg(0xF, 0)
			}
		case 0x5:
			name = fmt.Sprintf("SUB V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			v := c.readReg(hi0) - c.readReg(lo1)
			c.writeReg(hi0, v)

			if c.readReg(hi0) < c.readReg(lo1) {
				c.writeReg(0xF, 0)
			} else {
				c.writeReg(0xF, 1)
			}
		case 0x6:
			name = fmt.Sprintf("SHR V%s {, V%s}", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			v := c.readReg(hi0)

			c.writeReg(0xF, lib.Bit(v, 7))
			c.writeReg(hi0, v>>1)
		case 0x7:
			name = fmt.Sprintf("SUBN V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			v := c.readReg(lo1) - c.readReg(hi0)
			c.writeReg(hi0, v)

			if c.readReg(hi0) > c.readReg(lo1) {
				c.writeReg(0xF, 0)
			} else {
				c.writeReg(0xF, 1)
			}
		case 0xE:
			name = fmt.Sprintf("SHL V%s {, V%s}", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
			v := c.readReg(hi0)

			c.writeReg(0xF, lib.Bit(v, 0))
			c.writeReg(hi0, v<<1)
		default:
			implemented = false
		}
	case 0x9:
		name = fmt.Sprintf("SNE V%s, V%s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1))
		if c.readReg(hi0) != c.readReg(lo1) {
			c.pc += 2
		}
	case 0xA:
		v := inst & INST_MASK
		name = "LD I, " + lib.FormatHex(v, 3)
		slog.Debug("writeReg", "reg", "i", "v", lib.FormatHex(v, 3))
		c.i = v
	case 0xD:
		name = fmt.Sprintf("DRW V%s, V%s, %s", lib.FormatHex(hi0, 1), lib.FormatHex(lo1, 1), lib.FormatHex(lo0, 1)) //nolint:dupword
		x := c.readReg(hi0)
		y := c.readReg(lo1)

		sprite := make([]byte, lo0)

		for i := range sprite {
			sprite[i] = c.mem.Read(c.i + uint16(i))
		}

		c.ui.DrawSprite(x, y, sprite)
	case 0xF:
		switch lo {
		case 0x1E:
			name = "ADD I, V" + lib.FormatHex(hi0, 1)
			c.i = c.i + uint16(c.readReg(hi0))
		case 0x33:
			name = "LD B, V" + lib.FormatHex(hi0, 1)
			v := fmt.Sprintf("%03d", c.readReg(hi0))
			c.mem.Write(c.i, v[0]-'0')
			c.mem.Write(c.i+1, v[1]-'0')
			c.mem.Write(c.i+2, v[2]-'0')
		case 0x55:
			a := c.i

			name = fmt.Sprintf("LD %s, V%s", lib.FormatHex(a, 2), lib.FormatHex(hi0, 1))
			for x := range hi0 + 1 {
				c.mem.Write(a+uint16(x), c.readReg(x))
			}
		case 0x65:
			a := c.i

			name = fmt.Sprintf("LD V%s, %s", lib.FormatHex(hi0, 1), lib.FormatHex(a, 2))
			for x := range hi0 + 1 {
				c.writeReg(x, c.mem.Read(a+uint16(x)))
			}
		default:
			implemented = false
		}
	default:
		implemented = false
	}

	lib.Assert(implemented, fmt.Errorf("unimplemented instruction: %04X", inst))
	slog.Debug("executed", "inst", lib.FormatHex(inst, 4), "name", name, "pc", lib.FormatHex(c.pc, 4))

	return cycles
}
