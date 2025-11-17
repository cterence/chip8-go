package lib

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"golang.org/x/exp/constraints"
)

const (
	BYTE_SIZE byte = 8
)

type CompatibilityMode uint8

const (
	CM_NONE CompatibilityMode = iota
	CM_CHIP8
	CM_SUPERCHIP
	CM_XOCHIP
)

func Assert(condition bool, errorMsg error) {
	if !condition {
		panic("assertion failed: " + errorMsg.Error())
	}
}

func Bit[T constraints.Unsigned](b T, pos T) byte {
	return byte((b >> pos) & 1)
}

func SetBit(b byte, pos byte) byte {
	Assert(pos < BYTE_SIZE, fmt.Errorf("pos must be lower than %d, actual %d", BYTE_SIZE, pos))

	return 1<<pos | b
}

func ResetBit(b byte, pos byte) byte {
	Assert(pos < BYTE_SIZE, fmt.Errorf("pos must be lower than %d, actual %d", BYTE_SIZE, pos))

	return ^(1 << pos) & b
}

func SetLogger(logLevel string) {
	opts := &slog.HandlerOptions{ReplaceAttr: changeTimeFormat}

	switch logLevel {
	case "debug":
		opts.Level = slog.LevelDebug
	case "info":
		opts.Level = slog.LevelInfo
	case "error":
		opts.Level = slog.LevelError
	default:
		panic("unsupported log level: " + logLevel)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	slog.SetDefault(logger)
}

func FormatHex[T constraints.Unsigned](v T, length int) string {
	switch length {
	case 1:
		return fmt.Sprintf("%01X", v)
	case 2:
		return fmt.Sprintf("%02X", v)
	case 3:
		return fmt.Sprintf("%03X", v)
	case 4:
		return fmt.Sprintf("%04X", v)
	default:
		return strconv.FormatUint(uint64(v), 16)
	}
}

func changeTimeFormat(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(a.Value.Time().Format("15:04:05"))
	}

	return a
}
