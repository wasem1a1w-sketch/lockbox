#!/usr/bin/env bash
# One-time setup: desktop launcher + always-on local server (no terminal needed).
set -euo pipefail

REPO="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$REPO/lockbox"
HOME_DIR="$HOME"

if [ ! -x "$BIN" ]; then
  echo "Binary not found at $BIN — run 'make build' first." >&2
  exit 1
fi

# 1. Launcher script (path-patched to this checkout)
mkdir -p "$HOME_DIR/.local/bin"
sed "s|^BIN=.*|BIN=\"$BIN\"|" "$REPO/scripts/lockbox-web.sh" > "$HOME_DIR/.local/bin/lockbox-web"
chmod 755 "$HOME_DIR/.local/bin/lockbox-web"
echo "✓ launcher: ~/.local/bin/lockbox-web"

# 2. App menu entry
mkdir -p "$HOME_DIR/.local/share/applications"
sed "s|@HOME@|$HOME_DIR|g" "$REPO/scripts/lockbox.desktop" > "$HOME_DIR/.local/share/applications/lockbox.desktop"
chmod 644 "$HOME_DIR/.local/share/applications/lockbox.desktop"
if command -v desktop-file-validate >/dev/null 2>&1; then
  desktop-file-validate "$HOME_DIR/.local/share/applications/lockbox.desktop"
fi
update-desktop-database "$HOME_DIR/.local/share/applications" 2>/dev/null || true
echo "✓ menu entry: ~/.local/share/applications/lockbox.desktop"

# 3. Icon
mkdir -p "$HOME_DIR/.local/share/icons/hicolor/scalable/apps"
cp "$REPO/scripts/lockbox-icon.svg" "$HOME_DIR/.local/share/icons/hicolor/scalable/apps/lockbox.svg"
gtk-update-icon-cache -f -t "$HOME_DIR/.local/share/icons/hicolor" 2>/dev/null || true
echo "✓ icon: hicolor/scalable/apps/lockbox.svg"

# 4. systemd user service (server ready from login, restarts on failure)
mkdir -p "$HOME_DIR/.config/systemd/user"
sed "s|^ExecStart=.*|ExecStart=$BIN ui --no-open|" "$REPO/scripts/lockbox-ui.service" \
  > "$HOME_DIR/.config/systemd/user/lockbox-ui.service"
chmod 644 "$HOME_DIR/.config/systemd/user/lockbox-ui.service"
systemctl --user daemon-reload
systemctl --user enable --now lockbox-ui.service
echo "✓ service: systemctl --user enable --now lockbox-ui"

# 5. Health check (restart service if an old server held the port)
sleep 0.5
if curl -sf -m 2 http://127.0.0.1:8787/api/token >/dev/null 2>&1; then
  echo "✓ server responding on 127.0.0.1:8787"
else
  echo "⏳ port busy with an old instance — restarting service…"
  systemctl --user restart lockbox-ui.service
  sleep 1
  curl -sf -m 2 http://127.0.0.1:8787/api/token >/dev/null 2>&1 \
    && echo "✓ server responding on 127.0.0.1:8787" \
    || { echo "✗ server not responding — check: journalctl --user -u lockbox-ui" >&2; exit 1; }
fi

echo
echo "Done. Open Lockbox any time:"
echo "  • Activities → search “Lockbox”"
echo "  • or: ~/.local/bin/lockbox-web"
