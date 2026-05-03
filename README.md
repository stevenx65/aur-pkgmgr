# AUR PkgMgr

[English](#english) | [中文](#中文)

---

## 中文

`aur-pkgmgr` 是一个基于 [bubbletea](https://github.com/charmbracelet/bubbletea) 构建的终端用户界面（TUI）包管理器，用于 Arch Linux 的 `pacman`、`paru` 和 `yay`。

### 功能

- **搜索**官方仓库和 AUR 包
- **浏览**已安装的包列表，支持实时筛选
- **查看**可用更新
- **安装/删除/更新**单个包
- **全系统升级**
- **包详情**查看（依赖、描述、版本等）
- 自动检测 AUR 助手（`paru` → `yay` → 无）

### 安装

```bash
# 构建
cd aur-pkgmgr
go build -o pkgmgr .

# 安装到 PATH
sudo cp pkgmgr /usr/local/bin/

# （可选）创建 pacman 包装器以支持 `pacman -tui`
sudo cp pacman-wrapper.sh /usr/local/bin/pacman
sudo chmod 755 /usr/local/bin/pacman
```

### 使用

直接运行：

```bash
./pkgmgr
```

如果安装了包装器：

```bash
pacman -tui
```

### 快捷键

| 按键 | 功能 |
|------|------|
| `1` `2` `3` | 切换标签页（搜索/已安装/更新） |
| `j`/`k` 或 `↑`/`↓` | 导航列表 |
| `g` / `G` | 跳至顶部/底部 |
| `/` | 搜索（搜索标签页）/ 筛选（已安装标签页） |
| `Enter` | 查看包详情 |
| `i` / `r` / `u` | 安装 / 删除 / 更新 |
| `U` | 全系统升级 |
| `R` | 重新加载包列表 |
| `?` | 帮助 |
| `q` | 退出 |

> **注意**：安装、删除和更新操作需要 sudo 权限。如果 sudo 会话已过期，请先运行 `sudo -v`。

### 构建依赖

- Go 1.21+
- GitHub 库：
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/bubbles`
  - `github.com/charmbracelet/lipgloss`

---

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
