# chip8-go

A Golang CHIP-8 emulator.

## Test results

Automated screenshots from test runs done with [GitHub actions](./.github/workflows/golang-integration.yaml).

### Timendus

[CHIP-8 test suite](https://github.com/Timendus/chip8-test-suite)

|                CHIP-8 logo                |               IBM logo                |              Corax+               |
|:-----------------------------------------:|:-------------------------------------:|:---------------------------------:|
| ![chip8-logo](./results/1-chip8-logo.jpg) | ![ibm-logo](./results/2-ibm-logo.jpg) | ![corax+](./results/3-corax+.jpg) |

|              Flags              |              Quirks               |               Scrolling Super-CHIP low resolution               |
|:-------------------------------:|:---------------------------------:|:---------------------------------------------------------------:|
| ![flags](./results/4-flags.jpg) | ![quirks](./results/5-quirks.jpg) | ![scrolling-super-lores](./results/8-scrolling-super-lores.jpg) |


|              Scrolling Super-CHIP high resolution               |             Scrolling XO-CHIP low resolution              |             Scrolling XO-CHIP high resolution             |
|:---------------------------------------------------------------:|:---------------------------------------------------------:|:---------------------------------------------------------:|
| ![scrolling-super-hires](./results/8-scrolling-super-hires.jpg) | ![scrolling-xo-lores](./results/8-scrolling-xo-lores.jpg) | ![scrolling-xo-hires](./results/8-scrolling-xo-hires.jpg) |

## Improvement ideas

- [x] Embed SDL3
- [ ] Implement SUPER-CHIP & XO-CHIP
- [ ] Load rom with drag-n-drop if not provided
