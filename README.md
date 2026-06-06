# AUR PkgMgr

[中文版本](./README.zh.md) →

## English

`aur-pkgmgr` is a TUI (Terminal User Interface) package manager for Arch Linux's `pacman`, `paru`, and `yay`, built with [bubbletea](https://github.com/charmbracelet/bubbletea).

### Features

- **Search** official repos and AUR packages
- **Browse** installed packages with real-time filtering
- **View** available updates
- **Install/Remove/Update** individual packages
- **Full system upgrade**
- **Package details** (dependencies, description, version, etc.)
- Auto-detects AUR helper (`paru` → `yay` → none)

### Installation

```bash
# Build
cd aur-pkgmgr
go build -o pkgmgr .

# Install to PATH
sudo cp pkgmgr /usr/local/bin/

# (Optional) Create pacman wrapper to support `pacman -tui`
sudo cp pacman-wrapper.sh /usr/local/bin/pacman
sudo chmod 755 /usr/local/bin/pacman
```

### Usage

Run directly:

```bash
./pkgmgr
```

Or with the wrapper installed:

```bash
pacman -tui
```

### Keybindings

| Key | Action |
|-----|--------|
| `1` `2` `3` | Switch tabs (Search / Installed / Updates) |
| `j`/`k` or `↑`/`↓` | Navigate list |
| `g` / `G` | Go to top / bottom |
| `/` | Search (Search tab) / Filter (Installed tab) |
| `Enter` | View package details |
| `i` / `r` / `u` | Install / Remove / Update |
| `U` | Full system upgrade |
| `R` | Reload package list |
| `?` | Help |
| `q` | Quit |

> **Note**: Install, remove, and update operations require sudo. If your sudo session has expired, run `sudo -v` first.

### Build Dependencies

- Go 1.21+
- GitHub libraries:
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/bubbles`
  - `github.com/charmbracelet/lipgloss`

### License

[MIT](LICENSE) © 2026 stevenx65
