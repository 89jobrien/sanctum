# sanctum binary — design spec

Date: 2026-04-06

## Overview

A Go CLI binary that replaces the inline bash in the sanctum plugin's SessionStart hook and
provides on-demand diagnostics for the direnv → 1Password secret chain. The plugin's skills
and hooks call into `sanctum` rather than reimplementing shell logic.

---

## Architecture

```
sanctum/
├── cmd/sanctum/        # main package — CLI entry point (cobra)
├── internal/
│   ├── op/             # 1Password: auth check, item resolution, ref parsing
│   ├── envrc/          # .envrc chain traversal, op:// ref extraction
│   ├── tailscale/      # host reachability checks (ts status / ping)
│   └── report/         # structured output: human-readable + JSON
├── docs/superpowers/specs/
└── hooks/op-resolver-startup.sh  # replaced with: exec sanctum check --json
```

### Dependency rules

- `cmd/sanctum` imports `internal/*` only
- `internal/op`, `internal/envrc`, `internal/tailscale` are independent — no cross-imports
- `internal/report` imported by `cmd/sanctum` only
- All external calls (`op`, `tailscale`) go through interfaces injected at `main` — testable
  without live binaries

---

## Commands

### `sanctum check`

Full session-start diagnostic. Runs all checks and prints a summary.

```
sanctum check [--json] [--dir <path>]
```

Steps (in order):
1. Validate `op account list` — count accounts, flag if zero
2. Trace `.envrc` chain from `--dir` (default: `$PWD`) up to `$HOME`
3. Count `op://` refs per file; detect literal `op://` values in current env
4. Detect account conflicts (refs targeting multiple 1Password accounts in same chain)
5. Print summary table or `--json` object

Output (human):
```
1PASSWORD     OK  2 accounts
DIRENV CHAIN  3 files, 6 op:// refs
CONFLICTS     none
LITERAL REFS  none
```

Output (`--json`):
```json
{
  "op_accounts": 2,
  "envrc_files": 3,
  "op_refs": 6,
  "conflicts": [],
  "literal_refs": []
}
```

Exit codes: 0 = clean, 1 = warnings present, 2 = fatal (op not authed, binary missing).

### `sanctum trace [dir]`

Trace the `.envrc` chain from `dir` (default: CWD) and list every `op://` ref with its
vault, item UUID, and target account.

```
sanctum trace [dir]
```

Output:
```
/Users/joe/dev/minibox/.envrc
  op://toptal.1password.com/<uuid>/credential   → Toptal
  op://my.1password.com/<uuid>/password         → Personal

/Users/joe/dev/.envrc (source_up)
  op://<uuid>/<uuid>/token                      → Personal (UUID-only ref)
```

Flags unresolvable refs (item not found, account mismatch) inline.

### `sanctum validate [dir]`

For each `op://` ref found in the chain, call `op item get <uuid>` to confirm it resolves.
Reports pass/fail per ref.

```
sanctum validate [dir] [--account <account-shorthand>]
```

Output:
```
PASS  op://toptal.1password.com/<uuid>/credential
FAIL  op://<uuid>/<uuid>/token — item not found
```

Exit 1 if any ref fails.

### `sanctum scaffold [dir]`

Write a starter `.envrc` in `dir`. Detects existing patterns in parent chain (e.g.,
`op run --env-file=~/.secrets --`) and follows them. Does not overwrite an existing `.envrc`
without `--force`.

```
sanctum scaffold [dir] [--force]
```

Writes:
```bash
# .envrc — managed by sanctum
# Secrets loaded via 1Password CLI
# Edit op:// refs then run: direnv allow

dotenv_if_exists .env.local

# Inject secrets via op run — add vars to ~/.secrets then:
# op run --env-file=~/.secrets -- direnv exec . <command>
```

---

## Error Handling

- All `op` subprocess errors are wrapped with `fmt.Errorf("op: %w", err)`
- Missing binary (`op`, `tailscale`) → clear message + exit 2, never panic
- Partial failures in `validate` → report per-ref, continue, exit 1 at end
- No global state — `Config` struct passed through call chain

---

## Output

- Default: human-readable, 100-col width
- `--json`: machine-readable for hook consumption
- `--quiet`: suppress all output, use exit code only
- Errors always go to stderr

---

## Testing

- `internal/op`, `internal/envrc`, `internal/tailscale` each have an interface the real
  implementation satisfies — tests inject fakes
- Integration tests (build tag `integration`) require live `op` auth and real `.envrc` files
- `go test ./...` runs unit tests only; `go test -tags integration ./...` runs all

---

## Hook Integration

`hooks/op-resolver-startup.sh` becomes a thin wrapper:

```bash
#!/usr/bin/env bash
exec sanctum check --json
```

The JSON output is consumed by the Claude Code hook runtime directly.

---

## Tailscale (future, not v1)

`sanctum check` will optionally ping hosts referenced in `.envrc` (e.g. `MINIBOX_HOST`).
Deferred — design is in place via `internal/tailscale` stub, not implemented in v1.

---

## Out of Scope (v1)

- Watching for `.envrc` changes (fsnotify)
- Secret rotation or writing to 1Password
- TUI / interactive mode
- dumcp integration
