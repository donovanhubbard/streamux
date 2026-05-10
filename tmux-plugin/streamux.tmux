#!/usr/bin/env bash



date > ~/result.txt
echo "STARTING" >> ~/result.txt

main() {
  local CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$BASH_SOURCE}")" && pwd)"
  echo "CURRENT_DIR=$CURRENT_DIR" >> ~/result.txt
  source $CURRENT_DIR/scripts/*.sh
  echo "RC=$?" >> ~/result.txt
  echo "SOURCED" >> ~/result.txt
  echo "INSIDE MAIN" >> ~/result.txt
  tmux set-hook -g client-attached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  tmux set-hook -g client-detached "run-shell \"$CURRENT_DIR/scripts/update-viewer-count.sh\""
  echo "Assigned hooks" >> ~/result.txt
  echo "Before function call" >> ~/result.txt
  update_viewer_count
  RC=$?
  echo "Funciton call result $RC" >> ~/result.txt
  echo "AFTER update viewer_count" >> ~/result.txt
  echo "Exiting main" >> ~/result.txt
}

echo "Before main" >> ~/result.txt
main
echo "After main >> ~/result.txt

date >> ~/result.txt
