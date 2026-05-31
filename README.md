# lockbox

A CLI password vault that stores credentials in an AES-256-GCM encrypted file on disk.

## Installation

### Prerequisites
- **Go 1.21+** ([Download Go](https://go.dev/dl/))

### Option 1: `go install` (Recommended)

```bash
go install github.com/wasem1a1w-sketch/lockbox@latest
```

### Option 2: Build from source

```bash
git clone https://github.com/wasem1a1w-sketch/lockbox.git
cd lockbox
# On Windows:
go build -o lockbox.exe
# On Mac/Linux:
go build -o lockbox
```

## How it works

- Credentials are stored in `vault.lock` — an encrypted file using AES-256-GCM
- A master password derives the encryption key via PBKDF2 (100,000 SHA-256 iterations)
- Each credential has a unique **Index** identifier for targeted edit/delete operations

## Commands

### `--add 'domain:username:password'`

Add a new credential.

```
lockbox --add 'example.com:alice:myP@ss!'
```

### `--list`

Decrypt and display all stored credentials with their Index, account, username, password, and save date.

```
lockbox --list
```

### `--edit 'index:newpassword'`

Update the password of an existing credential by its Index. Other fields remain unchanged.

```
lockbox --edit '2:MyN3wP@ss!'
```

### `--reorder 'old_index:new_index'`

Move a credential's Index to a new position. Other credentials shift to fill the gap.

```
lockbox --reorder '3:1'
```

### `--delete index`

Delete a credential by its Index. Requires a positive integer.

```
lockbox --delete 2
```

### `--gen length`

Generate a cryptographically strong random password (upper/lower/digits/symbols). Specify the desired length as an argument.

```
lockbox --gen 20
```

## File structure

| File | Purpose |
|---|---|
| `main.go` | CLI entry point, flag parsing, command dispatch |
| `vault.go` | Credential model and CRUD operations |
| `password.go` | Random password generation |
| `crypto.go` | PBKDF2 key derivation, AES-256-GCM encrypt/decrypt |
| `storage.go` | Encrypted file persistence |
| `vault.lock` | Encrypted vault data (gitignored) |
