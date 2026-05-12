#!/usr/bin/env bash

set -e

SOCK_DIR="/var/run/tmux"

mkdir -p "$SOCK_DIR"

chown donovan:tmux-share /var/run/tmux

