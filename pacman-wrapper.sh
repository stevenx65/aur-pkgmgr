#!/bin/bash
# pacman wrapper — intercepts -tui/--tui to launch pkgmgr TUI
# Install: sudo install -m 755 this-script /usr/local/bin/pacman
# (Make sure /usr/local/bin is before /usr/bin in your PATH)

for arg in "$@"; do
    if [ "$arg" = "-tui" ] || [ "$arg" = "--tui" ]; then
        exec /usr/local/bin/pkgmgr
    fi
done

exec /usr/bin/pacman "$@"
