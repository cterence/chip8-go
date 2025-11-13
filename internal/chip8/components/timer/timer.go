package timer

import (
	"log/slog"
	"time"
)

type Timer struct {
	delay           uint8
	sound           uint8
	lastTimerUpdate time.Time
}

const (
	UPDATES_PER_SECOND   = 60
	TARGET_UPDATE_PERIOD = time.Second / UPDATES_PER_SECOND
)

func New() *Timer {
	t := &Timer{}

	return t
}

func (t *Timer) Init() {
	t.lastTimerUpdate = time.Now()
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

func (t *Timer) Tick(tickTime time.Time) {
	tickDuration := tickTime.Sub(t.lastTimerUpdate)
	if tickDuration > TARGET_UPDATE_PERIOD {
		if t.delay > 0 {
			t.delay--
			slog.Debug("delay timer", "value", t.delay)
		}

		if t.sound > 0 {
			t.sound--
			slog.Debug("sound timer", "value", t.sound)
		}

		t.lastTimerUpdate = time.Now()
	}
}
