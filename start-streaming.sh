#!/usr/bin/env bash

SOCK_ADDRESS=/var/run/tmux/tmux.sock

tmux -S "$SOCK_ADDRESS" list-sessions

tmux -S "$SOCK_ADDRESS"  new-session \; server-access -a -r tmuxview  \; run-shell -b -E "chmod 770 '$SOCK_ADDRESS'"

