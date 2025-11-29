# chip8-go

A Golang CHIP-8 interpreter. Compatible with CHIP-8, SUPER-CHIP and XO-CHIP instruction sets.

Uses [go-sdl3](https://github.com/Zyko0/go-sdl3) for the UI and audio.

## Usage

```bash
NAME:
   chip8-go - chip8 interpreter

USAGE:
   chip8-go [arguments...]

OPTIONS:
   --debug, -d                             print debug logs
   --disable-audio                         disable audio beeps
   --speed float, -s float                 interpreter speed (default: 1)
   --scale int                             pixel and window scale factor (default: 4)
   --test-flag uint, -t uint               populate 0x1FF address before run (used by timendus tests) (default: 0)
   --compatibility-mode string, -m string  force compatibility mode (chip8, super, xo)
   --help, -h                              show help
   --pause-after int, -p int               pause execution after t ticks (default: 0)
   --exit-after int, -e int                exit after t ticks (default: 0)
   --headless                              disable ui
   --screenshot                            save screenshot on exit
```

## Test results

Automated screenshots from test runs done with [GitHub actions](./.github/workflows/golang-integration.yaml).

### Timendus

[CHIP-8 test suite](https://github.com/Timendus/chip8-test-suite)

|                CHIP-8 logo                |               IBM logo                |              Corax+               |              Flags              |
|:-----------------------------------------:|:-------------------------------------:|:---------------------------------:|:-------------------------------:|
| ![chip8-logo](./results/1-chip8-logo.jpg) | ![ibm-logo](./results/2-ibm-logo.jpg) | ![corax+](./results/3-corax+.jpg) | ![flags](./results/4-flags.jpg) |

|                Quirks (CHIP-8)                |              Quirks (SUPER-CHIP)              |            Quirks (XO-CHIP)             |
|:---------------------------------------------:|:---------------------------------------------:|:---------------------------------------:|
| ![quirks-chip8](./results/5-quirks-chip8.jpg) | ![quirks-super](./results/5-quirks-super.jpg) | ![quirks-xo](./results/5-quirks-xo.jpg) |

|               Scrolling SUPER-CHIP low resolution               |              Scrolling SUPER-CHIP high resolution               |             Scrolling XO-CHIP low resolution              |             Scrolling XO-CHIP high resolution             |
|:---------------------------------------------------------------:|:---------------------------------------------------------------:|:---------------------------------------------------------:|:---------------------------------------------------------:|
| ![scrolling-super-lores](./results/8-scrolling-super-lores.jpg) | ![scrolling-super-hires](./results/8-scrolling-super-hires.jpg) | ![scrolling-xo-lores](./results/8-scrolling-xo-lores.jpg) | ![scrolling-xo-hires](./results/8-scrolling-xo-hires.jpg) |

## Improvement ideas

- [x] Embed SDL3
- [x] Implement SUPER-CHIP
- [x] Change window title when game paused
- [x] Create Nix package
- [x] Implement XO-CHIP
- [x] Fix various crashing games (skyward crash when falling to bottom)
- [ ] Fix skyward bug when character goes into wall
- [ ] Adjust speed dynamically
- [ ] Load rom with drag-n-drop if not provided
