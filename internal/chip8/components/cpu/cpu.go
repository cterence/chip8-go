package cpu

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cterence/chip8-go/internal/chip8/components/apu"
	"github.com/cterence/chip8-go/internal/chip8/components/memory"
	"github.com/cterence/chip8-go/internal/chip8/components/timer"
	"github.com/cterence/chip8-go/internal/chip8/components/ui"
	"github.com/cterence/chip8-go/internal/lib"
)

type register struct {
	value byte
}

type CPU struct {
	mem   *memory.Memory
	ui    *ui.UI
	timer *timer.Timer
	apu   *apu.APU

	reg   [REGISTER_COUNT]register
	pc    uint16
	i     uint16
	stack [STACK_SIZE]uint16
	sp    uint8

	romFileName             string
	pressedKey              *byte
	forcedCompatibilityMode bool
	compatibilityMode       lib.CompatibilityMode
	ticks                   int
	debugInfo               debugInfo

	SetCurrentTPS func(float32)
}

type regStorage struct {
	Registers [REGISTER_COUNT]register
}

type Option func(*CPU)

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

func New(mem *memory.Memory, ui *ui.UI, t *timer.Timer, apu *apu.APU, options ...Option) *CPU {
	c := &CPU{
		mem:   mem,
		ui:    ui,
		timer: t,
		apu:   apu,
	}

	for _, o := range options {
		o(c)
	}

	return c
}

func WithCompatibilityMode(mode lib.CompatibilityMode) Option {
	return func(c *CPU) {
		c.compatibilityMode = mode
		c.forcedCompatibilityMode = mode != lib.CM_NONE
	}
}

func WithRomFileName(romFileName string) Option {
	return func(c *CPU) {
		c.romFileName = romFileName
	}
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
	c.pressedKey = nil
	c.updateCompatibilityMode(c.compatibilityMode)
	c.ticks = 0
}

func (c *CPU) Tick() {
	inst := c.decodeInstruction()
	c.execute(inst)
	c.ticks++
}

func (c *CPU) DebugInfo() string {
	var debugInfo strings.Builder

	debugInfo.WriteString(fmt.Sprintf("OP: %-13s ", c.debugInfo.inst))
	debugInfo.WriteString(fmt.Sprintf("TK: %-5d ", c.ticks))
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
	lib.Assert(reg < REGISTER_COUNT, fmt.Errorf("illegal read to register V%s", lib.FormatHex(reg, 2)))
	v := c.reg[reg].value

	return v
}

func (c *CPU) writeReg(reg byte, v byte) {
	lib.Assert(reg < REGISTER_COUNT, fmt.Errorf("illegal write to register V%s", lib.FormatHex(reg, 2)))

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

func (c *CPU) updateCompatibilityMode(mode lib.CompatibilityMode) {
	if c.forcedCompatibilityMode && c.ticks > 0 {
		return
	}

	c.compatibilityMode = mode
	c.apu.CompatibilityMode = mode
	c.timer.CompatibilityMode = mode

	switch mode {
	case lib.CM_CHIP8, lib.CM_NONE:
		c.SetCurrentTPS(500)
	case lib.CM_SUPERCHIP:
		c.SetCurrentTPS(700)
	case lib.CM_XOCHIP:
		c.SetCurrentTPS(math.MaxFloat32) // Unlimited CPU ticks
	}
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
		switch hi0 {
		case 0x0:
			switch lo1 {
			case 0xE:
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
			case 0xC:
				c.updateCompatibilityMode(lib.CM_SUPERCHIP)
				c.debugInfo.inst = "SCD " + lib.FormatHex(lo0, 1)
				c.ui.Scroll(ui.SD_DOWN, int(lo0))
			case 0xD:
				c.compatibilityMode = lib.CM_XOCHIP
				c.debugInfo.inst = "SCU " + lib.FormatHex(lo0, 1)
				c.ui.Scroll(ui.SD_UP, int(lo0))
			case 0xF:
				switch lo0 {
				case 0xB:
					c.updateCompatibilityMode(lib.CM_SUPERCHIP)
					c.debugInfo.inst = "SCR 4"
					c.ui.Scroll(ui.SD_RIGHT, 4)
				case 0xC:
					c.updateCompatibilityMode(lib.CM_SUPERCHIP)
					c.debugInfo.inst = "SCL 4"
					c.ui.Scroll(ui.SD_LEFT, 4)
				case 0xD:
					c.debugInfo.inst = "EXIT"

					log.Println("exit called, pausing instead")
					c.ui.TogglePauseChip8()
				case 0xE:
					c.updateCompatibilityMode(lib.CM_SUPERCHIP)
					c.debugInfo.inst = "LORES"
					c.ui.ToggleHiRes(false)
				case 0xF:
					c.updateCompatibilityMode(lib.CM_SUPERCHIP)
					c.debugInfo.inst = "HIRES"
					c.ui.ToggleHiRes(true)
				default:
					implemented = false
				}
			default:
				implemented = false
			}
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
			if c.decodeInstruction() == 0xF000 {
				c.pc += 2
			}
		}
	case 0x4:
		c.debugInfo.inst = "SNE V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(lo, 2)

		if c.readReg(hi0) != lo {
			c.pc += 2
			if c.decodeInstruction() == 0xF000 {
				c.pc += 2
			}
		}
	case 0x5:
		switch lo0 {
		case 0x2:
			c.debugInfo.inst = "SFM V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.updateCompatibilityMode(lib.CM_XOCHIP)

			regCount := byte(math.Abs(float64(hi0)-float64(lo1))) + 1

			if hi0 < lo1 {
				for i := range regCount {
					c.mem.Write(c.i+uint16(i), c.readReg(hi0+i))
				}
			} else {
				for i := range regCount {
					c.mem.Write(c.i+uint16(i), c.readReg(hi0-i))
				}
			}
		case 0x3:
			c.debugInfo.inst = "LFM V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.updateCompatibilityMode(lib.CM_XOCHIP)

			regCount := byte(math.Abs(float64(hi0)-float64(lo1))) + 1

			if hi0 < lo1 {
				for i := range regCount {
					c.writeReg(hi0+i, c.mem.Read(c.i+uint16(i)))
				}
			} else {
				for i := range regCount {
					c.writeReg(hi0-i, c.mem.Read(c.i+uint16(i)))
				}
			}
		default:
			c.debugInfo.inst = "SE V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)

			if c.readReg(hi0) == c.readReg(lo1) {
				c.pc += 2
				if c.decodeInstruction() == 0xF000 {
					c.pc += 2
				}
			}
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

			if c.compatibilityMode == lib.CM_CHIP8 {
				c.writeReg(0xF, 0)
			}
		case 0x2:
			c.debugInfo.inst = "AND V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.writeReg(hi0, c.readReg(hi0)&c.readReg(lo1))

			if c.compatibilityMode == lib.CM_CHIP8 {
				c.writeReg(0xF, 0)
			}
		case 0x3:
			c.debugInfo.inst = "XOR V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1)
			c.writeReg(hi0, c.readReg(hi0)^c.readReg(lo1))

			if c.compatibilityMode == lib.CM_CHIP8 {
				c.writeReg(0xF, 0)
			}
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

			var v byte

			switch c.compatibilityMode {
			case lib.CM_CHIP8, lib.CM_XOCHIP:
				v = c.readReg(lo1)
			default:
				v = c.readReg(hi0)
			}

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

			var v byte

			switch c.compatibilityMode {
			case lib.CM_CHIP8, lib.CM_XOCHIP:
				v = c.readReg(lo1)
			default:
				v = c.readReg(hi0)
			}

			c.writeReg(hi0, v<<1)
			c.writeReg(0xF, lib.Bit(v, 7))
		default:
			implemented = false
		}
	case 0x9:
		c.debugInfo.inst = "SNE V" + lib.FormatHex(hi0, 1) + ", V" + lib.FormatHex(lo1, 1) + ""

		if c.readReg(hi0) != c.readReg(lo1) {
			c.pc += 2
			if c.decodeInstruction() == 0xF000 {
				c.pc += 2
			}
		}
	case 0xA:
		v := inst & ADDR_MASK
		c.debugInfo.inst = "LD I, " + lib.FormatHex(v, 3)
		c.i = v
	case 0xB:
		v := inst & ADDR_MASK

		switch c.compatibilityMode {
		case lib.CM_CHIP8, lib.CM_XOCHIP:
			c.debugInfo.inst = "JP V0, " + lib.FormatHex(v, 3)
			c.pc = v + uint16(c.readReg(0))
		default:
			c.debugInfo.inst = "JP V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(v, 3)
			c.pc = v + uint16(c.readReg(hi0))
		}

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

		spriteLen := lo0

		if lo0 == 0 {
			c.updateCompatibilityMode(lib.CM_SUPERCHIP)

			spriteLen = 2 * 16 // 2 col 16 rows
		}

		if c.ui.SelectedFrameBuffer == ui.SF_BOTH {
			spriteLen *= 2
		}

		sprite := make([]byte, spriteLen)

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
				if c.decodeInstruction() == 0xF000 {
					c.pc += 2
				}
			}
		case 0xA1:
			c.debugInfo.inst = "SKNP V" + lib.FormatHex(hi0, 1)

			if !c.ui.IsKeyPressed(c.readReg(hi0)) {
				c.pc += 2
				if c.decodeInstruction() == 0xF000 {
					c.pc += 2
				}
			}
		default:
			implemented = false
		}
	case 0xF:
		switch lo {
		case 0x00:
			c.pc += 2
			addr := c.decodeInstruction()

			c.debugInfo.inst = "LD I, " + lib.FormatHex(addr, 4)

			c.i = addr
		case 0x01:
			c.debugInfo.inst = "SFB " + lib.FormatHex(hi0, 1)

			c.updateCompatibilityMode(lib.CM_XOCHIP)
			c.ui.SelectFrameBuffer(hi0)
		case 0x02:
			c.debugInfo.inst = "LDP"

			var bytes [16]byte

			c.updateCompatibilityMode(lib.CM_XOCHIP)

			for i := range len(bytes) {
				bytes[i] = c.mem.Read(c.i + uint16(i))
			}

			c.apu.FillPatternBuffer(bytes)
		case 0x07:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", DT"
			c.writeReg(hi0, c.timer.GetDelay())
		case 0x0A:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", K"

			if c.pressedKey == nil {
				c.pressedKey = c.ui.GetPressedKey()

				return
			}

			if c.ui.IsKeyPressed(*c.pressedKey) {
				return
			}

			c.writeReg(hi0, *c.pressedKey)
			c.pressedKey = nil
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
			c.debugInfo.inst = "LD F, V" + lib.FormatHex(hi0, 1)
			digit := c.readReg(hi0)
			c.i = uint16(digit * 5)
		case 0x30:
			c.debugInfo.inst = "LD HF, V" + lib.FormatHex(hi0, 1)
			digit := c.readReg(hi0)
			c.i = uint16(digit*10 + 80)
		case 0x33:
			c.debugInfo.inst = "LD B, V" + lib.FormatHex(hi0, 1)
			v := fmt.Sprintf("%03d", c.readReg(hi0))
			c.mem.Write(c.i, v[0]-'0')
			c.mem.Write(c.i+1, v[1]-'0')
			c.mem.Write(c.i+2, v[2]-'0')
		case 0x3A:
			c.debugInfo.inst = "SP, V" + lib.FormatHex(hi0, 1)

			c.updateCompatibilityMode(lib.CM_XOCHIP)
			c.apu.SetPlaybackRate(c.readReg(hi0))
		case 0x55:
			c.debugInfo.inst = "LD " + lib.FormatHex(c.i, 2) + ", V" + lib.FormatHex(hi0, 1)

			i := c.i

			for x := range hi0 + 1 {
				c.mem.Write(i, c.readReg(x))
				i++
			}

			if c.compatibilityMode == lib.CM_CHIP8 || c.compatibilityMode == lib.CM_XOCHIP {
				c.i = i
			}
		case 0x65:
			c.debugInfo.inst = "LD V" + lib.FormatHex(hi0, 1) + ", " + lib.FormatHex(c.i, 2)

			i := c.i

			for x := range hi0 + 1 {
				c.writeReg(x, c.mem.Read(i))
				i++
			}

			if c.compatibilityMode == lib.CM_CHIP8 || c.compatibilityMode == lib.CM_XOCHIP {
				c.i = i
			}
		case 0x75:
			c.debugInfo.inst = "SF V" + lib.FormatHex(hi0, 1)
			c.updateCompatibilityMode(lib.CM_SUPERCHIP)

			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Println("failed to save flags: %w", err)

				break
			}

			flagDir := filepath.Join(homeDir, ".local/share/chip8-go")

			err = os.MkdirAll(flagDir, 0755)
			if err != nil {
				log.Println("failed to save flags: %w", err)

				break
			}

			storage := regStorage{Registers: c.reg}

			data, err := json.Marshal(storage)
			if err != nil {
				log.Println("failed to save flags: %w", err)

				break
			}

			romFileBaseName, _ := strings.CutSuffix(filepath.Base(c.romFileName), ".ch8")
			fileName := romFileBaseName + "-flags.json"

			err = os.WriteFile(filepath.Join(flagDir, fileName), data, 0644)
			if err != nil {
				log.Println("failed to save flags: %w", err)

				break
			}
		case 0x85:
			c.debugInfo.inst = "LF V" + lib.FormatHex(hi0, 1)
			c.updateCompatibilityMode(lib.CM_SUPERCHIP)

			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Println("failed to load flags: %w", err)

				break
			}

			flagDir := filepath.Join(homeDir, ".local/share/chip8-go")

			err = os.MkdirAll(flagDir, 0755)
			if err != nil {
				log.Println("failed to save flags: %w", err)

				break
			}

			romFileBaseName, _ := strings.CutSuffix(filepath.Base(c.romFileName), ".ch8")
			fileName := romFileBaseName + "-flags.json"

			data, err := os.ReadFile(filepath.Join(flagDir, fileName))
			if err != nil {
				_, err = os.Create(filepath.Join(flagDir, fileName))
				if err != nil {
					log.Println("failed to create flags file: %w", err)

					break
				}

				log.Println("failed to load flags: %w", err)

				break
			}

			var storage regStorage

			if len(data) > 0 {
				err = json.Unmarshal(data, &storage)
				if err != nil {
					log.Println("failed to load flags: %w", err)

					break
				}

				copy(c.reg[:], storage.Registers[:])
			}
		default:
			implemented = false
		}
	default:
		implemented = false
	}

	c.pc += 2

	log.Printf("Memory at 0x208-0x20C: %02X %02X %02X %02X %02X\n",
		c.mem.Read(0x208), c.mem.Read(0x209), c.mem.Read(0x20A), c.mem.Read(0x20B), c.mem.Read(0x20C))

	lib.Assert(implemented, fmt.Errorf("unimplemented instruction: %04X", inst))
}
