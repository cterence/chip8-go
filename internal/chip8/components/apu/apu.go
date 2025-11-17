package apu

import (
	"fmt"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
)

type APU struct {
	audioDisabled bool

	device      sdl.AudioDeviceID
	audioStream *sdl.AudioStream
	beep        []byte
}

type Option func(*APU)

func New(options ...Option) *APU {
	a := &APU{}

	for _, o := range options {
		o(a)
	}

	return a
}

func (a *APU) Init() error {
	if a.audioDisabled {
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

	a.device, err = sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDevice(spec)
	if err != nil {
		return fmt.Errorf("failed to get default playback audio device: %w", err)
	}

	if a.audioStream == nil {
		a.audioStream, err = sdl.CreateAudioStream(spec, spec)
		if err != nil {
			return fmt.Errorf("failed to create audio stream: %w", err)
		}
	}

	if a.audioStream.Device() == 0 {
		if err := a.device.BindAudioStream(a.audioStream); err != nil {
			return fmt.Errorf("failed to bind audio stream to device: %w", err)
		}
	}

	a.beep = generateBeep(int(spec.Freq), 440, 0.1)

	return nil
}

func WithAudioDisabled(audioDisabled bool) Option {
	return func(a *APU) {
		a.audioDisabled = audioDisabled
	}
}

func (a *APU) PlayBeep() error {
	if a.audioDisabled {
		return nil
	}

	available, err := a.audioStream.Available()
	if err != nil {
		return fmt.Errorf("failed to get available audio stream: %v", err)
	}

	if available < int32(len(a.beep)) {
		if err := a.audioStream.PutData(a.beep); err != nil {
			return fmt.Errorf("failed to put data to audio stream: %v", err)
		}
	}

	return nil
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
