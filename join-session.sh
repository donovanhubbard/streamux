#/usr/bin/env bash

SOCK_ADDRESS="/var/run/tmux/tmux.sock"

OUTPUT=$(/opt/homebrew/bin/tmux -S "$SOCK_ADDRESS" list-sessions 2>&1)
echo "$OUTPUT" | grep "no server running on" > /dev/null
RC=$?

if [[ "$RC" == 0 ]]; then
  echo "I'm sorry but the streamer isn't active now. Try again later."
  exit
fi

/opt/homebrew/bin/tmux -S "$SOCK_ADDRESS"  attach-session -r

