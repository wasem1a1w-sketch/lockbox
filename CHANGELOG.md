# Changelog

## [Unreleased] - Major Refactor: Cobra CLI, Config, Search/Filter

### Architecture
- **Split into internal packages** for separation of concerns:
  - `internal/vault` — Credential model, UUID-based identifiers, CRUD operations, filtering
  -   `internal/crypto` — PBKDF2 key derivation (100k iterations), AES-256-GCM encryption
  - `internal/storage` — Encrypted file persistence with auto directory creation
  - `internal/config` — Viper-based configuration with YAML, env var overrides
- **New entry point** at `cmd/lockbox/main.go` using Cobra

### CLI (Cobra Migration)
| Old Flag | New Command |
|----------|-------------|
| `--add 'a:b:c'` | `lockbox add "a:b:c"` |
| `--list` | `lockbox list` |
| `--edit 'idx:pass'` | `lockbox edit "idx:pass"` |
| `--delete idx` | `lockbox delete idx` |
| `--gen N` | `lockbox generate N` |
| `--reorder 'o:n'` | `lockbox reorder "o:n"` |
| `--change-master-password` | `lockbox change-master` |

**New global flags:**
- `--vault PATH` — Override vault file location
- `--iterations N` — Override PBKDF2 iterations

**Shell completion:**
```bash
lockbox completion bash > /etc/bash_completion.d/lockbox
lockbox completion zsh > "${fpath[1]}/_lockbox"
lockbox completion fish > ~/.config/fish/completions/lockbox.fish
```

### Configuration File
**Location:** `~/.config/lockbox/config.yaml`

```yaml
vault_path: "~/.local/share/lockbox/vault.lock"
kdf_iterations: 100000
default_gen_length: 20
```

**Environment variable overrides:**
- `LOCKBOX_VAULT_PATH`
- `LOCKBOX_KDF_ITERATIONS`
- `LOCKBOX_DEFAULT_GEN_LENGTH`

### Search & Filter (List Command)
```bash
# Search in account, username, or password
lockbox list --search github

# Filter by username
lockbox list --user alice

# Show passwords (hidden by default)
lockbox list --show-password

# Combine
lockbox list -s github -u alice -p
```

### Security Improvements
- **UUID v4 identifiers** replace fragile sequential indices
  - Stable across reorders/deletes
  - Commands accept both ID (full or 8-char prefix) and display index
- **Auto directory creation** on first save (0700 perms)
- **File permissions** 0600 maintained

### Vault Model Changes
```go
type Credential struct {
    ID        string    // UUID v4 (stable identifier)
    Account   string
    Username  string
    Password  string
    SavedAt   time.Time
    SortOrder int       // Display order (1-based, mutable)
}
```

- `ReOrder` now updates `SortOrder` only, IDs remain stable
- `Filter(search, username)` supports partial matching on account/user/pass

### Files Added
```
internal/vault/vault.go       # Vault model, CRUD, filtering
internal/vault/password.go    # Password generation
internal/crypto/crypto.go     # Key derivation, encryption
internal/storage/storage.go   # Encrypted file I/O
internal/config/config.go     # Config loading with Viper
cmd/lockbox/root.go           # Root command, global flags
cmd/lockbox/add.go            # Add credential
cmd/lockbox/list.go           # List with search/filter
cmd/lockbox/edit.go           # Edit password (by ID or index)
cmd/lockbox/delete.go         # Delete (by ID or index)
cmd/lockbox/generate.go       # Generate password
cmd/lockbox/reorder.go        # Reorder display position
cmd/lockbox/changemaster.go   # Change master password
cmd/lockbox/main.go           # Entry point, password reading
```

### Files Modified
- `go.mod` — Added dependencies: `github.com/spf13/cobra`, `github.com/spf13/viper`, `github.com/google/uuid`

### Files Removed (Legacy)
- `main.go` (root) — Replaced by `cmd/lockbox/main.go`
- `vault.go` (root) — Moved to `internal/vault/vault.go`
- `storage.go` (root) — Moved to `internal/storage/storage.go`
- `crypto.go` (root) — Moved to `internal/crypto/crypto.go`
- `password.go` (root) — Moved to `internal/vault/password.go`

### Build
```bash
go build -o lockbox.exe ./cmd/lockbox
```

### Migration Notes
1. **Existing vault files are compatible** — same encryption format, salt header
2. **Credentials gain UUIDs on next save** — `Add` assigns UUID; existing entries get UUID on next edit
3. **Config file auto-created** on first run with defaults
4. **Old flags no longer work** — use subcommands instead