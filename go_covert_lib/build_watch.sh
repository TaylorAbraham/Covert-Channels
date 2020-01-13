#!/bin/sh

if ! type "inotifywatch" >/dev/null # If inotifywatch is NOT a recognized command
then
    echo "inotify-tools are not installed. Installing..."
    sudo apt install inotify-tools -y
fi

if ! type "tmux" >/dev/null
then
    echo "tmux is not installed. Installing..."
    sudo apt install tmux -y
fi

tmux kill-session -t 'Covert-Channels' # Kill any existing CC sessions
tmux new-session -d -s 'Covert-Channels'
tmux new -s 'Covert-Channels'
tmux rename-window 'Main'
tmux send-keys 'ls' 'C-m'
tmux send-keys 'gs' 'C-m'
tmux send-keys 'gb' 'C-m'
tmux new-window
tmux rename-window 'Background Tasks'
tmux send-keys 'sudo ./main -p 8080' 'C-m'
tmux split-window -v
tmux send-keys 'cd client && ./client_watch.sh' 'C-m'
tmux split-window -h
tmux split-window -h -t 1
tmux send-keys 'sudo ./main -p 8081' 'C-m'

tmux -2 attach-session -t 'Covert-Channels'
