#!/usr/bin/env bash


update_viewer_count(){
  local MOTTO="crab your dog after you pet"
  date >> ~/status.txt
  STATUS_RIGHT=$(tmux show-option -gvq status-right)
  echo $STATUS_RIGHT >> ~/status.txt
  CLIENT_LIST=$(tmux list-clients )
  echo "$CLIENT_LIST" >> ~/status.txt
  CLIENT_COUNT=$(tmux list-clients | wc -l )
  echo "$CLIENT_COUNT" >> ~/status.txt
  CLIENT_COUNT=$(( $CLIENT_COUNT - 1 ))
  echo "$CLIENT_COUNT" >> ~/status.txt

  if [[ $CLIENT_COUNT -gt 0 ]]; then
    STATUS_RIGHT="${STATUS_RIGHT/ $MOTTO/}"
    STATUS_RIGHT="${STATUS_RIGHT/ Viewers: */}"
    STATUS_RIGHT="$STATUS_RIGHT Viewers: $CLIENT_COUNT 👤"
  else
    STATUS_RIGHT="${STATUS_RIGHT/Viewers: */}"
    STATUS_RIGHT="$STATUS_RIGHT $MOTTO"
  fi
  echo "$STATUS_RIGHT" >> ~/status.txt

  tmux set-option -g status-right "$STATUS_RIGHT"
}

update_viewer_count
