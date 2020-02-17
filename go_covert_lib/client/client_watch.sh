#!/bin/bash

GREEN='\033[0;32m'
NC='\033[0m'

while true; do
  npm run build
  $@ &
  PID=$!
  inotifywait -r -e modify --exclude 'node_modules\/' .
  kill $PID
  echo -e "${GREEN}Changes detected! Reloading...${NC}"
done
