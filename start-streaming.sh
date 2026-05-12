#!/usr/bin/env bash

tmux -S /var/run/tmux/tmux.sock new-session \; server-access -a -r tmuxview  \; run-shell -b -E "chmod 770 '$SOCK_ADDRESS'"
