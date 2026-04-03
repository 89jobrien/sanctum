---
name: op-resolver
description: This skill should be used when the user asks to "resolve secrets", "check
  1password", "debug env", "why isn't my op:// ref working", "trace direnv chain",
  "fix secret not loading", or invokes /joe-secrets:op-resolver. Also fires automatically
  at session start via the SessionStart hook.
---

# op-resolver

Validate 1Password authentication, trace the `.envrc` `source_up` chain, detect `op://`
URI conflicts between accounts, and report missing environment variables.

## Step 1: Validate 1Password Auth

```bash
op account list
```

Expected output: table showing at least two accounts (toptal.1password.com and my.1password.com).

If the command fails or shows no accounts: 1Password CLI is not authed. Instruct user:

> "Run `op signin` or open 1Password and unlock it, then retry."

## Step 2: Trace .envrc source_up Chain

Starting from the current working directory, trace upward:

```bash
d="$PWD"
while [ "$d" != "$HOME" ] && [ "$d" != "/" ]; do
  [ -f "$d/.envrc" ] && echo "$d/.envrc"
  d=$(dirname "$d")
done
```

For each `.envrc` found, check for `op://` references using the Grep tool (not bash grep).

Report the chain:

```
DIRENV CHAIN (CWD → HOME)
  /Users/joe/dev/minibox/.envrc  — 3 op:// refs
  /Users/joe/dev/.envrc          — 1 op:// ref (source_up)
  /Users/joe/.envrc              — 2 op:// refs (source_up)
```

## Step 3: Detect op:// Account Conflicts

Scan all `.envrc` files in the chain for `op://` refs and identify which account they target:

- `op://toptal.1password.com/...` → Toptal account
- `op://my.1password.com/...` or `op://Personal/...` → Personal account
- `op://<uuid>/...` — check UUID against `op account list` to identify account

If refs target multiple accounts in the same chain, flag as a potential conflict:

> "WARNING: .envrc chain references both Toptal and Personal 1Password accounts.
> Commands using `op run` may need `--account` flag to disambiguate."

## Step 4: Detect Literal op:// URIs in Shell

Claude's shell context cannot resolve `op://` URIs directly. If the environment
contains literal `op://` values (not resolved secrets), warn:

> "WARNING: Environment variable FOO contains a literal op:// URI.
> Use `op run -- <command>` to inject resolved values into commands."

## Step 5: Report Summary

```
1PASSWORD AUTH     OK (2 accounts)
DIRENV CHAIN       3 files found, 6 op:// refs total
ACCOUNT CONFLICTS  None detected
LITERAL OP:// REFS None in current environment
```

## Common Issues

| Issue | Fix |
|-------|-----|
| `op account list` fails | Run `op signin` or unlock 1Password |
| `source_up` not loading parent | Run `direnv reload` in each parent dir |
| Wrong account selected | Add `--account <uuid>` to `op run` |
| Literal op:// in env | Wrap command with `op run --` |
| op item not found | Use UUID not item name in op:// path |

## Always Use UUIDs

Never use item names in `op://` paths — they may not resolve correctly across accounts.
Use `op item list --vault <vault>` to get exact item UUIDs.

## Slash Command

Available as `/joe-secrets:op-resolver` for on-demand mid-session invocation.
