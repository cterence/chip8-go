package lib_test

import (
	"testing"

	"github.com/cterence/chip8-go/internal/lib"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	b := byte(0b00010010)

	t.Run("Bit", func(t *testing.T) {
		assert.Equal(t, byte(0), lib.Bit(b, 0))
		assert.Equal(t, byte(1), lib.Bit(b, 1))
		assert.Equal(t, byte(0), lib.Bit(b, 2))
		assert.Equal(t, byte(0), lib.Bit(b, 3))
		assert.Equal(t, byte(1), lib.Bit(b, 4))
		assert.Equal(t, byte(0), lib.Bit(b, 5))
		assert.Equal(t, byte(0), lib.Bit(b, 6))
		assert.Equal(t, byte(0), lib.Bit(b, 7))
		assert.Panics(t, func() { lib.Bit(b, 8) })
	})

	t.Run("SetBit", func(t *testing.T) {
		assert.Equal(t, byte(0b00010011), lib.SetBit(b, 0))
		assert.Equal(t, byte(0b00010010), lib.SetBit(b, 1))
		assert.Equal(t, byte(0b00010110), lib.SetBit(b, 2))
		assert.Equal(t, byte(0b00011010), lib.SetBit(b, 3))
		assert.Equal(t, byte(0b00010010), lib.SetBit(b, 4))
		assert.Equal(t, byte(0b00110010), lib.SetBit(b, 5))
		assert.Equal(t, byte(0b01010010), lib.SetBit(b, 6))
		assert.Equal(t, byte(0b10010010), lib.SetBit(b, 7))
		assert.Panics(t, func() { lib.SetBit(b, 8) })
	})

	t.Run("ResetBit", func(t *testing.T) {
		assert.Equal(t, byte(0b00010010), lib.ResetBit(b, 0))
		assert.Equal(t, byte(0b00010000), lib.ResetBit(b, 1))
		assert.Equal(t, byte(0b00010010), lib.ResetBit(b, 2))
		assert.Equal(t, byte(0b00010010), lib.ResetBit(b, 3))
		assert.Equal(t, byte(0b00000010), lib.ResetBit(b, 4))
		assert.Equal(t, byte(0b00010010), lib.ResetBit(b, 5))
		assert.Equal(t, byte(0b00010010), lib.ResetBit(b, 6))
		assert.Equal(t, byte(0b00010010), lib.ResetBit(b, 7))
		assert.Panics(t, func() { lib.ResetBit(b, 8) })
	})
}
