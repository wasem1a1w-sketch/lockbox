# lockbox

A CLI password vault that stores credentials in an AES-256-GCM encrypted file on disk.

## Installation

### Prerequisites
- **Go 1.22+** ([Download Go](https://go.dev/dl/))

### Option 1: `go install` (Recommended)

```bash
go install github.com/wasem1a1w-sketch/lockbox@latest
```

### Option 2: Build from source

```bash
git clone https://github.com/wasem1a1w-sketch/lockbox.git
cd lockbox
# On Windows:
go build -o lockbox.exe ./cmd/lockbox
# On Mac/Linux:
go build -o lockbox ./cmd/lockbox
```

## How it works

- Credentials are stored in an encrypted file (default: `~/.local/share/lockbox/vault.lock`)
- A master password derives the encryption key via PBKDF2 (100,000 SHA-256 iterations)
- Each credential has a unique **UUID** identifier for targeted edit/delete operations

## Commands

### `add "domain:username:password"`

Add a new credential.

```bash
lockbox add "example.com:alice:myP@ss!"
```

### `list`

Decrypt and display all stored credentials with their ID, account, username, password, and save date. Passwords are hidden by default.

```bash
lockbox list
lockbox list --show-password
lockbox list --search github
lockbox list --user alice
```

### `edit "id_or_index:newpassword"`

Update the password of an existing credential by its ID or display index.

```bash
lockbox edit "2:MyN3wP@ss!"
lockbox edit "a1b2c3d4:MyN3wP@ss!"
```

### `reorder "old_index:new_index"`

Move a credential to a new display position. Other credentials shift to fill the gap.

```bash
lockbox reorder "3:1"
```

### `delete id_or_index`

Delete a credential by its ID or display index.

```bash
lockbox delete 2
lockbox delete "a1b2c3d4"
```

### `generate [length]`

Generate a cryptographically strong random password (upper/lower/digits/symbols). Default length is 20.

```bash
lockbox generate 20
```

### `change-master`

Change the master password for the entire vault. You will be prompted for the current password, then asked to enter and confirm a new one. The vault is re-encrypted with a fresh salt and a new key derived from the new password.

```bash
lockbox change-master
```

### `config`

Manage configuration settings.

```bash
lockbox config show
lockbox config set vault_path ~/my-vault.lock
lockbox config set kdf_iterations 100000
lockbox config set default_gen_length 12
```

## Global Flags

```bash
--vault PATH        # Override vault file location
--iterations N      # Override PBKDF2 iterations
```

## Configuration

**Location:** `~/.config/lockbox/config.yaml`

```yaml
vault_path: "~/.local/share/lockbox/vault.lock"
kdf_iterations: 100000
default_gen_length: 12
```

**Environment variable overrides:**
- `LOCKBOX_VAULT_PATH`
- `LOCKBOX_KDF_ITERATIONS`
- `LOCKBOX_DEFAULT_GEN_LENGTH`

## Shell Completion

```bash
# Bash
lockbox completion bash > /etc/bash_completion.d/lockbox

# Zsh
lockbox completion zsh > "${fpath[1]}/_lockbox"

# Fish
lockbox completion fish > ~/.config/fish/completions/lockbox.fish
```

## Web UI

Local browser UI served by the same binary.

```bash
make ui                                        # build frontend (once) + binary
./lockbox ui                                   # serve http://127.0.0.1:8787
./lockbox ui --port 8787 --no-open             # custom port, don't open browser
./lockbox ui --lock-timeout 10m                # auto-lock idle session (default 15m)
```

### Desktop launcher (no terminal)

One-time setup — after this you never touch a terminal:

```bash
make install-desktop
```

What it installs:

- **App-menu entry** ("Lockbox") with a padlock icon → opens a Chrome
  `--app` window (no tabs, no address bar) straight to the vault
- **systemd user service** (`lockbox-ui`) → server is already running from
  login and auto-restarts on failure, so the window opens in <1s
- **Self-healing launcher** (`~/.local/bin/lockbox-web`) → if the server
  is somehow down, it starts it before opening the window

Open it: `Activities` → search "Lockbox", or run `~/.local/bin/lockbox-web`.
Logs: `journalctl --user -u lockbox-ui`. Remove everything: `make uninstall-desktop`.

Development mode (hot reload, Vite on :5173 proxying to Go on :8787):

```bash
make build            # once, so the Go server exists
make dev              # terminal 1: vite dev server
./lockbox ui --no-open # terminal 2: API server
```

If frontend assets are missing, the server returns a page telling you to run `make ui`.

Security: binds 127.0.0.1 only; rejects non-loopback Host and cross-origin
requests; per-boot token injected into the page (re-fetch falls back to
`GET /api/token`); session is memory-only and auto-locks after inactivity.

## Security

- **AES-256-GCM** authenticated encryption
- **PBKDF2** key derivation with 100,000 iterations
- Master password input is masked (not echoed to terminal)
- File permissions: 0600 (vault), 0700 (directories)
- Random salt per vault file

## File Structure

```
lockbox/
├── cmd/lockbox/          # CLI entry point and commands
│   ├── main.go           # Entry point, password reading
│   ├── root.go           # Root Cobra command, global flags
│   ├── add.go            # Add credential
│   ├── list.go           # List with search/filter
│   ├── edit.go           # Edit password (by ID or index)
│   ├── delete.go         # Delete (by ID or index)
│   ├── generate.go       # Generate password
│   ├── reorder.go        # Reorder display position
│   ├── changemaster.go   # Change master password
│   ├── config.go         # Config management
│   ├── ui.go             # Web UI server command
│   └── vault.go          # Shared vault-unlocking helper
├── internal/
│   ├── vault/            # Credential model, CRUD, filtering
│   ├── crypto/           # PBKDF2 key derivation, AES-256-GCM
│   ├── storage/          # Encrypted file I/O (incl. legacy format)
│   ├── config/           # Viper-based configuration
│   └── server/           # Local HTTP API + embedded SPA
├── ui/                   # Vue 3 + Vite frontend source
├── Makefile              # build / ui / dev / test targets
├── go.mod
├── CHANGELOG.md
└── README.md
```
