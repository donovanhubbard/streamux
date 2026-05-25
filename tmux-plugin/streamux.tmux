#!/usr/bin/env bash

main() {
  echo "Starting main" > ~/app.log
  local CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$BASH_SOURCE}")" && pwd)"
  echo "CURRENT_DIR='$CURRENT_DIR'" >> ~/app.log
  source "$CURRENT_DIR/scripts/update-recording-light.sh"
  source "$CURRENT_DIR/scripts/update-viewer-count.sh"
  echo "After source" >> ~/app.log
  tmux set-hook -g client-attached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  tmux set-hook -g client-detached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  echo "After hooks" >> ~/app.log
  update_viewer_count
  echo "Done" >> ~/app.log
}
main

