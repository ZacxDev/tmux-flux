#!/usr/bin/env bash
# tmux-flux tmux integration
# Add to tmux.conf: run-shell /path/to/tmux-flux.tmux
# Or manually add: bind-key C-q display-popup -E -w 60% -h 60% "tmux-flux"

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find tmux-flux binary
if command -v tmux-flux &> /dev/null; then
    TMUX_FLUX="tmux-flux"
elif [[ -x "$CURRENT_DIR/tmux-flux" ]]; then
    TMUX_FLUX="$CURRENT_DIR/tmux-flux"
elif [[ -x "$HOME/.local/bin/tmux-flux" ]]; then
    TMUX_FLUX="$HOME/.local/bin/tmux-flux"
else
    echo "tmux-flux not found in PATH, current directory, or ~/.local/bin"
    exit 1
fi

# Bind Ctrl+Q to open tmux-flux in a popup
tmux bind-key C-q display-popup -E -w 80% -h 80% "$TMUX_FLUX"
