#!/usr/bin/env bash

main() {
  local CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$BASH_SOURCE}")" && pwd)"
  source $CURRENT_DIR/scripts/*.sh
  tmux set-hook -g client-attached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  tmux set-hook -g client-detached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  update_viewer_count
}
main

