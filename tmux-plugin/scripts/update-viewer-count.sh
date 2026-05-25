#!/usr/bin/env bash

update_viewer_count(){
  echo "Starting update viewer count" >> ~/app.log
  local CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$BASH_SOURCE}")" && pwd)"
  DISCORD_LINK=$(cat "$CURRENT_DIR/../discord.txt")
  DISCORD_DISPLAY="$DISCORD_LINK 🌐"
  local MOTTO="crab your dog after you pet"
  STATUS_RIGHT=$(tmux show-option -gvq status-right)
  echo "STATUS_RIGHT='$STATUS_RIGHT'" >> ~/app.log
  CLIENT_LIST=$(tmux list-clients )
  CLIENT_COUNT=$(tmux list-clients | wc -l )
  CLIENT_COUNT=$(( $CLIENT_COUNT - 1 ))

  if [[ $CLIENT_COUNT -gt 0 ]]; then
    STATUS_RIGHT="${STATUS_RIGHT/ $MOTTO/}"
    STATUS_RIGHT="${STATUS_RIGHT/Viewers: */}"
    STATUS_RIGHT="${STATUS_RIGHT/ $DISCORD_DISPLAY/}"
    STATUS_RIGHT="$STATUS_RIGHT ${DISCORD_DISPLAY}Viewers: $CLIENT_COUNT 👤"
  else
    STATUS_RIGHT="${STATUS_RIGHT/ $MOTTO/}"
    STATUS_RIGHT="${STATUS_RIGHT/Viewers: */}"
    STATUS_RIGHT="${STATUS_RIGHT/ $DISCORD_DISPLAY/}"
    STATUS_RIGHT="$STATUS_RIGHT $MOTTO"
  fi
  echo "STATUS_RIGHT='$STATUS_RIGHT'" >> ~/app.log

  tmux set-option -g status-right "$STATUS_RIGHT"
}

update_viewer_count

