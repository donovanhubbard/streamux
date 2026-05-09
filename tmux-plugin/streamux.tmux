#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$CURRENT_DIR/scripts/*.sh"

main() {
  tmux set-hook -g client-attached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  tmux set-hook -g client-detached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  # tmux set-hook -g client-attached 'run-shell update_viewer_count'
  # tmux set-hook -g client-detached 'run-shell update_viewer_count'
  update_viewer_count
}

main
