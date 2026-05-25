#!/usr/bin/env bash

update_recording_light(){
  echo "Inside recording light" >> ~/app.log

  local RED_RECORDING="#[fg=red]⏺#[fg=#15161e]"
  local BLACK_RECORDING="#[fg=black]⏺#[fg=#15161e]"
  echo "TMUX='$TMUX'" >> ~/app.log
  local SOCK=$(echo $TMUX | cut -d',' -f1)
  echo "SOCK='$SOCK'">> ~/app.log

  STATUS_RIGHT=$(tmux show-option -gvq status-right)
  echo "Existing STATUS_RIGHT='$STATUS_RIGHT'" >> ~/app.log

  if [[ "$SOCK" == "/var/run/tmux/tmux.sock" ]]; then
    # connected to streaming sock
    ps -ef | grep streamux-gate | grep -v grep 2> /dev/null > /dev/null
    RESPONSE=$?

    if [[ $RESPONSE -eq 0 ]]; then
      STATUS_RIGHT="$STATUS_RIGHT $RED_RECORDING $STATUS_STYLE"
    else
      STATUS_RIGHT="$STATUS_RIGHT $BLACK_RECORDING $STATUS_STYLE"
    fi
  else
    echo "if is false" >> ~/app.log
  fi

  echo "STATUS_RIGHT='$STATUS_RIGHT'" >> ~/app.log
  tmux set-option -g status-right "$STATUS_RIGHT"
}

update_recording_light

