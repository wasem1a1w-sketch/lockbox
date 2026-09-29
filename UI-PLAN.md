# lockbox Web UI Plan

Status: approved — building.

## Decisions (locked)

| Question | Decision |
|---|---|
| Architecture | **gcsim way** — React SPA served by local Go HTTP server (`127.0.0.1`), same packages as CLI |
| Design | **Modern shadcn-style** — Radix + CVA + Tailwind + lucide, dark-only, indigo accent (NOT Blueprint/gcsim look) |
| Auto-lock | **15 min inactivity** (`--lock-timeout` flag, default `15m`) |
| Edit scope | **All fields** — account, username, password editable |
| Reorder | **Drag-and-drop** via `@dnd-kit/sortable` (keyboard accessible) |

## Architecture

```
Browser (React SPA, shadcn-style dark, Vite build)
   │  same-origin fetch, JSON bodies, X-Lockbox-Token header
   ▼
`lockbox ui` → net/http server, bound 127.0.0.1:8787
   │  imports internal/{vault,storage,crypto,config} — zero logic duplication
   ▼
~/.local/share/lockbox/vault.lock (same file CLI uses; legacy format auto-upgraded)
```

Mirror of gcsim's `pkg/servermode` + Vite-proxy dev pattern, scoped to lockbox.

## Security model

- Bind `127.0.0.1` only (never 0.0.0.0)
- **Host check** — `r.Host` must be `127.0.0.1:port` or `localhost:port`, else 403 (DNS-rebinding guard)
- **Origin check** — non-empty `Origin` must match own origin, else 403
- **Random token** — 32-byte hex per boot; injected into served `index.html` AND bootstrap `GET /api/token` (dev mode); every `/api/*` request requires `X-Lockbox-Token` → 403 otherwise
- **No CORS headers** anywhere + JSON bodies + custom header → cross-origin browser requests fail preflight (CSRF closed)
- Master password / key / salt **memory only** — never logged, never persisted beyond vault file itself
- **Auto-lock**: janitor goroutine clears session after 15 min without API activity; API then returns 401 → UI shows unlock screen
- Session bound to vault path: if `vault_path` config changes mid-session → session invalidated
- `Cache-Control: no-store` on index + API responses

## API

| Endpoint | Body / Query | Returns |
|---|---|---|
| `GET /api/token` | — | `{token}` (bootstrap; no token required) |
| `GET /api/session` | — | `{vaultExists, unlocked}` |
| `POST /api/unlock` | `{password, confirm?}` | 200 \| 400 (new-vault confirm mismatch) \| 401 (bad pw) |
| `POST /api/lock` | — | 200 |
| `GET /api/vault` | `?search=&user=` | `{credentials:[{id,account,username,password,savedAt}]}` |
| `POST /api/vault` | `{account,username,password}` | 201 `{credential}` |
| `PUT /api/vault/{id}` | `{account,username,password}` | 200 `{credential}` |
| `DELETE /api/vault/{id}` | — | 200 |
| `POST /api/vault/reorder` | `{id, to}` (1-based target position) | 200 |
| `POST /api/generate` | `{length}` | `{password}` |
| `POST /api/change-master` | `{newPassword}` | 200 (rekeys session) |
| `GET /api/config` | — | `{vaultPath,kdfIterations,defaultGenLength,configPath}` |
| `POST /api/config` | `{key,value}` | `{key,value,sessionInvalidated}` |
| `GET /`, `GET /assets/*` | — | embedded SPA (token injected) |

Errors: JSON `{error}` everywhere. 403 = token/host/origin problem; 401 = locked (frontend → unlock screen).

## Backend files (Go)

- `internal/server/server.go` — mux, host/origin/token middleware, `127.0.0.1` listener, static serving with token injection
- `internal/server/session.go` — in-memory session `{key,salt,iterations,vaultPath,lastActivity}`, mutex; janitor (30s tick) enforces idle timeout
- `internal/server/handlers.go` — all endpoints; per-request `config.Load()` (always current); vault ops = Load → mutate → Save through `internal/{vault,storage}`
- `internal/server/static.go` — `//go:embed all:dist` + placeholder page when frontend not built
- `internal/server/server_test.go` — httptest: token/host/origin rejections, create-flow, wrong-pw 401, CRUD roundtrip, reorder, generate, config, auto-lock expiry, legacy vault → IDs persisted
- `cmd/lockbox/ui.go` — `lockbox ui [--port 8787] [--no-open] [--lock-timeout 15m]`, opens browser via `xdg-open`
- Register in `cmd/lockbox/root.go`

### Supporting changes (shared with CLI, CLI behavior unchanged)

- `internal/vault/vault.go`:
  - `EnsureIDs() int` — assign UUID to empty-ID rows; normalize zero `SortOrder`; returns changed count
  - `EditFields(id, account, username, password) error` — all-field edit (new; CLI keeps `EditPasswordByID`)
  - `SortByOrder()` — stable sort by display order
- `internal/storage/storage.go` — `LoadWithPassword` unmarshals via legacy-aware struct: old JSON `index` → `SortOrder` when `sort_order` missing (old CLI reordered Index only; file order ≠ display order)
- `internal/config/config.go` — extract `Set(key, value)` (validate, expand `~/`, write yaml 0600); CLI `config set` + server share it
- Server unlock runs `EnsureIDs` and persists immediately when changed → one-time legacy upgrade (UUIDs + order + current cipher format)

## Frontend files (`ui/`)

Scaffold: Vite + React 18 + TypeScript, Tailwind v4 (`@tailwindcss/vite`), Radix primitives + CVA + `clsx` + `tailwind-merge` + `lucide-react`, `@dnd-kit/{core,sortable,utilities}`.

```
ui/
├── package.json            # dev/build scripts; build = vite build + copy into internal/server/dist
├── vite.config.ts          # /api proxy → 127.0.0.1:8787 (dev)
├── index.html              # window.__LOCKBOX_TOKEN__ = "__LOCKBOX_TOKEN__" placeholder
└── src/
    ├── main.tsx, App.tsx   # session state machine: boot → locked | ready; 401 → locked
    ├── index.css           # dark tokens, indigo accent
    ├── lib/{api.ts,utils.ts}   # fetch wrapper (token header, 401 event, errors), cn()
    ├── hooks/              # useVault (search/user params, 250ms debounce), useToast, useClipboard
    ├── components/ui/      # button, input, label, dialog, dropdown-menu, tooltip, toast,
    │                       # slider, table, skeleton (shadcn-style, hand-written)
    └── features/
        ├── unlock/UnlockScreen      # unlock OR create (password + confirm, mismatch error, shake)
        ├── vault/VaultTable         # dnd sortable rows: # grip, account, username,
        │                            #   masked pw + per-row reveal/copy, date, actions dropdown
        ├── vault/Toolbar            # search (/ hotkey), global show toggle, + Add
        ├── dialogs/Add, Edit (all fields + inline generate), DeleteConfirm
        ├── dialogs/Generate         # length slider 8–64, regenerate, copy
        ├── dialogs/ChangeMaster     # new + confirm
        ├── dialogs/Settings         # vault path, iterations, gen length → POST config
        └── layout/Header            # lock button, settings menu, toast host
```

UX: toasts on every action, `Esc` closes dialogs, optimistic reorder with rollback, empty-state callout, copy → toast.

## Build glue

- `Makefile`: `make ui` = `cd ui && npm install && npm run build` + `go build -o lockbox ./cmd/lockbox`; `make test` = gofmt/vet/test
- `ui` build copies `ui/dist/*` → `internal/server/dist/` (tracked `.gitkeep` survives; built files gitignored)
- Dev: `go run ./cmd/lockbox ui --no-open` + `npm run dev` (Vite 5173 proxies `/api`)
- `.gitignore`: `ui/node_modules`, `ui/dist`, `internal/server/dist/index.html`, `internal/server/dist/assets/`
- README: "Web UI" section (build, usage, dev mode, security notes)

## Verification

1. `gofmt -l .` clean, `go vet ./...`, `go test ./...`
2. `cd ui && npm install && npm run build` → dist embedded
3. `go build -o lockbox ./cmd/lockbox`
4. curl smoke: token 403 enforcement, wrong Host 403, wrong Origin 403, create → unlock → CRUD roundtrip, reorder, generate, config write, auto-lock (short timeout), legacy fixture unlock → UUIDs persisted
5. Browser e2e on **copy of real vault** first, then real: unlock → list → add/edit/delete → drag reorder → generate → change-master → manual lock
6. CLI regression: `./lockbox list` still reads same file after UI session

## Out of scope (v1)

- WASM in-browser mode (gcsim dual-executor) — server-only, single local binary
- Mobile layout beyond basic responsiveness
- Import/export, folders/tags, password strength meter (follow-ups)

---

## Phase 2 — v2 frontend: fixes + React → Vue rewrite

Status: **executed** (2026-09-29).
Supersedes frontend-stack decisions above (React / Radix / dnd-kit / dark-only).
Backend, API, security, embed sections above remain valid unchanged.

### Decisions (locked)

| Question | Decision |
|---|---|
| Framework | **Vue 3.5**, `<script setup lang="ts">`; React removed entirely |
| Headless widgets | **reka-ui ^2.10** (Radix port: Dialog, DropdownMenu, Tooltip, Slider) |
| Icons | **lucide-vue-next ^1.0** (same icon set) |
| Drag reorder | **vuedraggable ^4.1.0** (SortableJS; must pin — npm `latest` tag is stale Vue 2) + **↑/↓ keydown on grip** to compensate lost dnd-kit keyboard sensor |
| Theme | **light + dark** — `.dark` class on `<html>`, `localStorage("lockbox-theme")`, system-preference fallback, no-flash init in `index.html`, sun/moon toggle in header |
| Scale | root `font-size: 18px` (uniform enlargement, kept from v1.1) |
| Generate default | **12** — UI `ref(12)` + Go `SetDefault("default_gen_length")` 20 → 12 |
| Typecheck | `vue-tsc` |

### Phase A — fixes (land inside Vue code; Go part standalone)

1. **Generate default = 12** — `internal/config/config.go` lines 36 + 88 (`v.SetDefault("default_gen_length", 20)` → `12`), so CLI `lockbox generate` and Settings dialog agree. Caveat: an existing `config.yaml` with explicit `20` overrides → `lockbox config set default_gen_length 12`.
2. **Dialog centering bug** — root cause proven from built CSS: Tailwind v4 emits `-translate-x-1/2 { translate: … }` (the **`translate` property**) while `@keyframes dialog-in` used `transform: translate(-50%,-50%) scale(.96)` → both applied during animation = −100% offset → box starts off-center and snaps to center when animation ends. Fix: keyframes animate only `opacity` + `scale` (never `transform`), same for `fade-in-up` / `toast-in` (also makes tooltip/dropdown popper positioning animation-proof).

### Phase B — stack swap

**Remove:** react, react-dom, @types/react*, @vitejs/plugin-react, @radix-ui/* ×5, @dnd-kit/* ×3, class-variance-authority, lucide-react.
**Add:** vue ^3.5, reka-ui ^2.10, lucide-vue-next ^1.0, vuedraggable ^4.1.0, @vitejs/plugin-vue (check peer vs vite ^6 at install), vue-tsc.
**Keep:** tailwindcss, @tailwindcss/vite, vite, typescript, clsx, tailwind-merge.

**Untouched:** `index.css` (tokens, light/dark palettes, 18px scale, ambient glows, fixed keyframes), `index.html` (token placeholder + theme init), `lib/api.ts`, `lib/theme.ts`, `lib/utils.ts`, build script (`vite build && cp -r dist/* ../internal/server/dist/`), Makefile, **all Go server code + tests**.

### Phase C — file map (21 `.tsx` → ~15 files)

| React (delete) | Vue (create) |
|---|---|
| `main.tsx` | `main.ts` — `createApp(App).mount("#root")` |
| `App.tsx` | `App.vue` — boot/locked/ready state machine, `LOCKED_EVENT` listener, `<Toasts/>`, `<TooltipProvider>` |
| `hooks/useToast.tsx` | `composables/useToast.ts` + `components/Toasts.vue` (same `toast(title, opts)` API) |
| `components/ui/{button,input,label,dialog,dropdown-menu,tooltip,slider}.tsx` | `components/ui/*.vue` — Button = plain variant→class map (CVA dropped); Dialog/Dropdown/Tooltip/Slider wrap reka-ui with identical classes/animations |
| `features/unlock/UnlockScreen.tsx` | `UnlockScreen.vue` |
| `features/vault/VaultScreen.tsx` | `VaultScreen.vue` + `components/Row.vue` — header (theme toggle, lock), toolbar, `<VueDraggable v-model="creds">` around rows; grip = SortableJS handle + ↑/↓ handler; optimistic reorder + rollback |
| `features/dialogs/{Credential,Delete,Generate,ChangeMaster,Settings}.tsx` | `features/dialogs/*.vue` — `v-model:open`, emits `saved`/`deleted`/`session-invalidated`; Generate slider default **12** |
| `vite.config.ts` | edit: `plugins: [vue(), tailwindcss()]` |
| `tsconfig.json` | edit: drop `jsx`, include `*.vue`, typecheck via `vue-tsc` |

Props pattern: `open` → `v-model:open`; `onOpenChange` → `emit("update:open", …)`; `useState` → `ref`; `useEffect` → `watch` / `onMounted`.

### Phase D — verification

1. `npm install` → `npx vue-tsc --noEmit` → `npm run build` (dist copied into `internal/server/dist/`)
2. `gofmt -l`, `go vet ./...`, `go test ./...`, `go build -o lockbox ./cmd/lockbox`
3. curl: token injected in served index, JS/CSS 200, CSS contains light palette + `font-size:18px` + fixed keyframes
4. Browser checklist: dialog centered from first frame · generate default **12** · theme toggle both modes (persists, no flash) · tooltips · drag reorder + ↑/↓ · CRUD + search + copy · lock/unlock
5. CLI regression: `./lockbox list` reads UI-written vault file

### Risk

Only behavioral delta vs v1: rows not keyboard-*draggable* (↑/↓ on focused grip compensates). API, security model, embed, Go tests untouched.
