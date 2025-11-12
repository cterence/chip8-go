package memory

import (
	"fmt"

	"github.com/cterence/chip8-go-v2/internal/lib"
)

type Memory struct {
	ram [RAM_SIZE]byte
}

const (
	RAM_SIZE uint16 = 4096

	INTERPRETER_RAM_START uint16 = 0
	INTERPRETER_RAM_END   uint16 = 0x1FF

	PROGRAM_RAM_START uint16 = INTERPRETER_RAM_END + 1
	PROGRAM_RAM_END   uint16 = 0xFFF
	PROGRAM_RAM_SIZE  uint16 = PROGRAM_RAM_END - PROGRAM_RAM_START
)

func New() *Memory {
	m := &Memory{}

	return m
}

func (m *Memory) Read(a uint16) byte {
	lib.Assert(a < RAM_SIZE, fmt.Errorf("address must be lower than 0x%03X, actual 0x%03X", RAM_SIZE, a))
	lib.Assert(a > INTERPRETER_RAM_END, fmt.Errorf("illegal access to interpreter ram section: 0x%03X", a))

	return m.ram[a]
}

func (m *Memory) Write(a uint16, v byte) {
	lib.Assert(a < RAM_SIZE, fmt.Errorf("address must be lower than 0x%03X, actual 0x%03X", RAM_SIZE, a))
	lib.Assert(a > INTERPRETER_RAM_END, fmt.Errorf("illegal access to interpreter ram section: 0x%03X", a))

	m.ram[a] = v
}
