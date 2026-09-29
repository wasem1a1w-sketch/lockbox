#!/usr/bin/env bash
# Lockbox desktop launcher: ensure local server is up, then open an app-style window.
set -euo pipefail

BIN="@BIN@"
URL="http://127.0.0.1:8787"
LOG="${TMPDIR:-/tmp}/lockbox-ui.log"

# Start the server in the background if it is not already answering.
if ! curl -sf -m 1 "$URL/api/token" >/dev/null 2>&1; then
  nohup "$BIN" ui --no-open >>"$LOG" 2>&1 &
  for _ in $(seq 1 40); do
    if curl -sf -m 1 "$URL/api/token" >/dev/null 2>&1; then
      break
    fi
    sleep 0.15
  done
fi

if command -v google-chrome >/dev/null 2>&1; then
  exec google-chrome --app="$URL" --no-first-run --no-default-browser-check
elif command -v google-chrome-stable >/dev/null 2>&1; then
  exec google-chrome-stable --app="$URL" --no-first-run --no-default-browser-check
else
  exec xdg-open "$URL"
fi
