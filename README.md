# Blutui

**Blutui** is a terminal user interface (TUI) application for controlling Bluesound devices and browsing music libraries (Local and Tidal) from your terminal. It provides a fast, keyboard-driven interface for playback control, browsing artists and albums, and managing your listening experience—all without leaving your shell.

---

## Features

- **Bluesound Integration:** Control playback, volume, mute, repeat modes, and more on Bluesound devices via HTTP API.
- **Music Library Browsing:** Browse and search your local and Tidal music libraries, view artists, albums, and tracks.
- **Fast TUI:** Built with [tcell](https://github.com/gdamore/tcell) and [tview](https://github.com/mkozjak/tview) for responsive, mouse-enabled terminal UI.
- **Keyboard Shortcuts:** Efficient navigation and control with comprehensive keybindings.
- **Status Bar:** Real-time player status and feedback.
- **Help Screen:** In-app help for all keybindings.
- **Caching:** Local caching for faster library browsing.

---

## Installation

### Prerequisites

- Go 1.22.2 or newer
- A Bluesound device accessible on your network

### Build from Source

```sh
git clone https://github.com/mkozjak/blutui.git
cd blutui
make
```

This will build the `blutui` binary in the `bin/` directory.

---

## Usage

Run the application from your terminal:

```sh
./bin/blutui
```

By default, Blutui will attempt to connect to your Bluesound device at `http://bluesound.local:11000`. You may need to adjust your device's hostname or network settings if this does not work.

### Flags

- `--version` : Display the application version.

---

## Keybindings

| Key / Combo         | Action                                      |
|---------------------|---------------------------------------------|
| `1`                 | Show local library                          |
| `2`                 | Show Tidal library                          |
| `↵` (Enter)         | Start playback                              |
| `x`                 | Play selected song only                     |
| `p`                 | Play/Pause                                  |
| `s`                 | Stop                                        |
| `>`                 | Next song                                   |
| `<`                 | Previous song                               |
| `+`                 | Volume up                                   |
| `-`                 | Volume down                                 |
| `m`                 | Toggle mute                                 |
| `r`                 | Toggle repeat mode (none, all, one)         |
| `Ctrl+f`            | Page down                                   |
| `Ctrl+b`            | Page up                                     |
| `Ctrl+d`            | Half page down                              |
| `Ctrl+u`            | Half page up                                |
| `o`                 | Jump to currently playing artist            |
| `f`                 | Search artists                              |
| `u`                 | Update library                              |
| `h`                 | Show help screen                            |
| `q`                 | Quit app                                    |

Press `h` at any time to view the help screen with all keybindings.

---

## Contributing

Contributions are welcome! Please open issues or submit pull requests on [GitHub](https://github.com/mkozjak/blutui).

### Development

- Code is organized in `internal/` modules for app logic, player, library, keyboard, and UI components.
- Use `make` to build and test.
- Please follow idiomatic Go practices and document your code.

---

## License

This project is licensed under the **GNU Affero General Public License v3.0**. See the [LICENSE](./LICENSE) file for details.

---

## Acknowledgements

- [tcell](https://github.com/gdamore/tcell) and [tview](https://github.com/rivo/tview) for TUI components.
- Bluesound for their open HTTP API.

---

## Author

Maintained by [mkozjak](https://github.com/mkozjak).
