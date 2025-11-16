package timer

import (
	"fmt"
	"log"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
)

type Timer struct {
	audioDisabled bool

	delay uint8
	sound uint8

	device      sdl.AudioDeviceID
	audioStream *sdl.AudioStream
	beep        []byte
}

type Option func(*Timer)

func New(options ...Option) *Timer {
	t := &Timer{}

	for _, o := range options {
		o(t)
	}

	return t
}

func WithAudioDisabled(audioDisabled bool) Option {
	return func(t *Timer) {
		t.audioDisabled = audioDisabled
	}
}

func (t *Timer) Init() error {
	t.delay = 0
	t.sound = 0

	if t.audioDisabled {
		return nil
	}

	err := sdl.Init(sdl.INIT_AUDIO)
	if err != nil {
		return fmt.Errorf("failed to init sdl audio: %w", err)
	}

	spec := &sdl.AudioSpec{
		Freq:     44100,
		Format:   sdl.AUDIO_U8,
		Channels: 1,
	}

	t.device, err = sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDevice(spec)
	if err != nil {
		return fmt.Errorf("failed to get default playback audio device: %w", err)
	}

	if t.audioStream == nil {
		t.audioStream, err = sdl.CreateAudioStream(spec, spec)
		if err != nil {
			return fmt.Errorf("failed to create audio stream: %w", err)
		}
	}

	if t.audioStream.Device() == 0 {
		if err := t.device.BindAudioStream(t.audioStream); err != nil {
			return fmt.Errorf("failed to bind audio stream to device: %w", err)
		}
	}

	t.beep = generateBeep(int(spec.Freq), 440, 0.1)

	return nil
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
		if t.sound > 1 && !t.audioDisabled {
			available, err := t.audioStream.Available()
			if err != nil {
				log.Printf("failed to get available audio stream: %v", err)
			}

			if available < int32(len(t.beep)) {
				if err := t.audioStream.PutData(t.beep); err != nil {
					log.Printf("failed to put data to audio stream: %v", err)
				}
			}
		}

		t.sound--
	}
}

func generateBeep(sampleRate int, freq float64, seconds float64) []byte {
	period := 1.0 / freq                 // Period of one wave cycle in seconds
	cycles := math.Round(seconds * freq) // Number of complete cycles
	exactDuration := cycles * period     // Exact duration for complete cycles

	samples := int(float64(sampleRate) * exactDuration)
	buf := make([]byte, samples)

	center := 128.0 // Center point for unsigned 8-bit audio
	volume := 60.0  // Amplitude

	for i := 0; i < samples; i++ {
		// Generate sine wave
		sine := math.Sin(2 * math.Pi * freq * float64(i) / float64(sampleRate))

		// Generate 8-bit unsigned sample (centered at 128)
		sample := center + (volume * sine)

		// Clamp to valid range [0, 255]
		if sample < 0 {
			sample = 0
		} else if sample > 255 {
			sample = 255
		}

		buf[i] = uint8(sample)
	}

	return buf
}
