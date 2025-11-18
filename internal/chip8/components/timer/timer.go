package timer

import (
	"github.com/cterence/chip8-go/internal/chip8/components/apu"
	"github.com/cterence/chip8-go/internal/lib"
)

type Timer struct {
	apu               *apu.APU
	CompatibilityMode lib.CompatibilityMode

	delay uint8
	sound uint8
}

type Option func(*Timer)

func New(apu *apu.APU, options ...Option) *Timer {
	t := &Timer{}

	for _, o := range options {
		o(t)
	}

	t.apu = apu

	return t
}

func (t *Timer) Init() {
	t.delay = 0
	t.sound = 0
}

func (t *Timer) GetDelay() byte {
	return t.delay
}

func (t *Timer) SetDelay(v byte) {
	t.delay = v
}

func (t *Timer) SetSound(v byte) {
	t.sound = v
}

func (t *Timer) Tick() {
	if t.delay > 0 {
		t.delay--
	}

	if t.sound > 0 {
		if t.CompatibilityMode == lib.CM_XOCHIP || t.sound > 1 {
			t.apu.PlaySound()
		}

		t.sound--
	} else {
		t.apu.ResetPhase()
	}
}
