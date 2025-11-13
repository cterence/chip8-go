package timer

type Timer struct {
	delay uint8
	sound uint8
}

func New() *Timer {
	t := &Timer{}

	return t
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
		t.sound--
	}
}
