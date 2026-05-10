#!/usr/bin/env bash

echo "top of update-viewer-count.sh" >> ~/result.txt

update_viewer_count(){
  echo "inside update_viewer_count" >>~/result.txt
  local CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$BASH_SOURCE}")" && pwd)"
  echo "CURRENT_DIR=$CURRENT_DIR" >> ~/result.txt
  DISCORD_LINK=$(cat "$CURRENT_DIR/../discord.txt")
  DISCORD_DISPLAY="$DISCORD_LINK 🌐"
  echo "discord='$DISCORD_LINK'" >> ~/result.txt
  local MOTTO="crab your dog after you pet"
  date >> ~/result.txt
  STATUS_RIGHT=$(tmux show-option -gvq status-right)
  echo $STATUS_RIGHT >> ~/result.txt
  CLIENT_LIST=$(tmux list-clients )
  echo "$CLIENT_LIST" >> ~/result.txt
  CLIENT_COUNT=$(tmux list-clients | wc -l )
  echo "$CLIENT_COUNT" >> ~/result.txt
  CLIENT_COUNT=$(( $CLIENT_COUNT - 1 ))
  echo "$CLIENT_COUNT" >> ~/result.txt
  echo "#####################" >> ~/result.txt
  echo "'$STATUS_RIGHT'" >> ~/result.txt

  if [[ $CLIENT_COUNT -gt 0 ]]; then
    STATUS_RIGHT="${STATUS_RIGHT/ $MOTTO/}"
    echo "'$STATUS_RIGHT'" >> ~/result.txt
    STATUS_RIGHT="${STATUS_RIGHT/Viewers: */}"
    echo "'$STATUS_RIGHT'" >> ~/result.txt
    STATUS_RIGHT="${STATUS_RIGHT/ $DISCORD_DISPLAY/}"
    STATUS_RIGHT="$STATUS_RIGHT ${DISCORD_DISPLAY}Viewers: $CLIENT_COUNT 👤"
  else
    STATUS_RIGHT="${STATUS_RIGHT/ $MOTTO/}"
    echo "'$STATUS_RIGHT'" >> ~/result.txt
    STATUS_RIGHT="${STATUS_RIGHT/Viewers: */}"
    echo "'$STATUS_RIGHT'" >> ~/result.txt
    STATUS_RIGHT="${STATUS_RIGHT/ $DISCORD_DISPLAY/}"
    STATUS_RIGHT="$STATUS_RIGHT $MOTTO"
  fi
  echo "'$STATUS_RIGHT'" >> ~/result.txt
  echo "#####################" >> ~/result.txt

  tmux set-option -g status-right "$STATUS_RIGHT"
  echo "Ending inside_viewer_count" >> ~/result.txt
}

update_viewer_count

echo "bottom of update-viewer-count.sh" >> ~/result.txt
