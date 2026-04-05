---
name: onboard
description: Use when the user says "onboard me", "set up sanctum", "what does sanctum do", "walk
  me through setup", or invokes /sanctum:onboard. Guides through installation and verification of
  the sanctum 1Password + direnv session management plugin.
---

# onboard — sanctum plugin setup

## Overview

**sanctum** provides automatic secrets validation at every Claude session start:

1. Validates `op account list` — confirms 1Password is authed
2. Traces `.envrc` `source_up` chain from CWD
3. Counts `op://` refs and detects literal URIs leaked into environment
4. Returns a system message to Claude with readiness summary

The `SessionStart` hook fires automatically — no manual invocation needed.

## Step 1: Prerequisites

```bash
which op direnv claude
op account list
```

- `op` — 1Password CLI (required): `brew install 1password-cli`
- `direnv` — optional; chain tracing skipped gracefully without it
- `claude` — Claude Code CLI (required)

If `op` is missing or not signed in:
```bash
brew install 1password-cli
op signin
```

## Step 2: Clone and Init

```bash
git clone https://github.com/89jobrien/sanctum ~/dev/sanctum
cd ~/dev/sanctum
just init
```

`just init` will:
1. Set `core.hooksPath = .githooks` for auto-reinstall on source changes
2. Check for `op` and `direnv`, prompting if missing
3. Register the local plugin marketplace
4. Install via `claude plugin install sanctum@local`

If `just` is not installed: `brew install just`.

## Step 3: Verify SessionStart Hook

Start a fresh Claude session (close and reopen, or `claude` in a new terminal).
Within the first response, you should see:

```
1Password: 2 account(s) authed.
Direnv chain: N .envrc file(s) found, N op:// refs.
```

If this block is absent:
```bash
ls -l ~/dev/sanctum/hooks/
claude plugin list | grep sanctum
```

## Step 4: On-Demand Secret Resolution

Run `/sanctum:op-resolver` mid-session to re-validate secrets or trace a broken `.envrc` chain.

Triggers automatically when Claude detects unresolved `op://` refs in error output.

## Onboarding Complete

> sanctum runs silently at every session start.
> If 1Password is locked or `op://` refs are unresolved, it surfaces the issue immediately
> so you can fix it before hitting a runtime failure.
>
> Install **atelier** alongside sanctum for the full session-start experience.
