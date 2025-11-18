package apu

import (
	"fmt"
	"log"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/cterence/chip8-go/internal/lib"
)

type APU struct {
	CompatibilityMode lib.CompatibilityMode
	audioDisabled     bool

	device       sdl.AudioDeviceID
	audioStream  *sdl.AudioStream
	pattern      [16]byte
	sampleRate   int32
	playbackRate float64
	phase        float64
}

const (
	TPS                 = 30
	PATTERN_BUFFER_BITS = 128
)

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

	a.pattern = [16]byte{
		0xF0, 0x0, 0x0, 0x0,
		0xF0, 0x0, 0x0, 0x0,
		0xF0, 0x0, 0x0, 0x0,
		0xF0, 0x0, 0x0, 0x0,
	}
	a.playbackRate = 4000
	a.sampleRate = 44100

	spec := &sdl.AudioSpec{
		Freq:     a.sampleRate,
		Format:   sdl.AUDIO_U8,
		Channels: 1,
	}

	err := sdl.Init(sdl.INIT_AUDIO)
	if err != nil {
		return fmt.Errorf("failed to init sdl audio: %w", err)
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

	return nil
}

func WithAudioDisabled(audioDisabled bool) Option {
	return func(a *APU) {
		a.audioDisabled = audioDisabled
	}
}

func (a *APU) PlaySound() {
	if a.audioDisabled {
		return
	}

	a.playPatternBuffer()
}

func (a *APU) FillPatternBuffer(bytes [16]byte) {
	copy(a.pattern[:], bytes[:])
}

func (a *APU) SetPlaybackRate(pitch byte) {
	pow := (float64(pitch) - 64) / 4
	a.playbackRate = 4000 * math.Pow(2, pow)
}

func (a *APU) playPatternBuffer() {
	available, err := a.audioStream.Available()
	if err != nil {
		log.Printf("failed to get available audio stream: %v", err)
	}

	sound := a.generateSound()

	if available < int32(len(sound)) {
		if err := a.audioStream.PutData(sound); err != nil {
			log.Printf("failed to put data to audio stream: %v", err)
		}
	}
}

func (a *APU) generateSound() []byte {
	numSamples := int(a.sampleRate / TPS)
	sound := make([]byte, numSamples)
	step := a.playbackRate / float64(a.sampleRate)

	for i := range len(sound) {
		idx := int(a.phase) % PATTERN_BUFFER_BITS
		byteIdx := idx / lib.BYTE_SIZE
		bitIdx := byte(7 - (idx % lib.BYTE_SIZE))
		b := lib.Bit(a.pattern[byteIdx], bitIdx)

		if b == 1 {
			sound[i] = 192
		} else {
			sound[i] = 64
		}

		a.phase += step
	}

	a.phase = math.Mod(a.phase, float64(PATTERN_BUFFER_BITS))

	return sound
}
