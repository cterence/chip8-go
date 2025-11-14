package debugger

import (
	"strings"

	"github.com/cterence/chip8-go/internal/chip8/components/cpu"
	"github.com/cterence/chip8-go/internal/chip8/components/memory"
)

type Debugger struct {
	cpu *cpu.CPU
	mem *memory.Memory
}

func New(cpu *cpu.CPU, mem *memory.Memory) *Debugger {
	d := &Debugger{}

	d.cpu = cpu
	d.mem = mem

	return d
}

func (d *Debugger) DebugLog() string {
	var debugLog strings.Builder

	debugLog.WriteString("CPU | " + d.cpu.DebugInfo())

	return debugLog.String()
}
