# chip8-go

A Golang CHIP-8 interpreter.

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
- [ ] Implement XO-CHIP
- [ ] Adjust speed dynamically
- [ ] Change window title when game paused
- [ ] Load rom with drag-n-drop if not provided
- [ ] Create Nix package
