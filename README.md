# joe-secrets

1Password and direnv session management — secrets validation and env chain tracing.

## Installation

Install alongside `joe-dev` for the full session-start experience:

```bash
claude --plugin-dir ~/.claude/plugins/joe-secrets
```

## Skills

| Skill | Trigger |
|-------|---------|
| op-resolver | "resolve secrets", "check 1password", "debug env", `/joe-secrets:op-resolver` |

## Hooks

`SessionStart` — runs automatically on every new Claude session:

1. Validates `op account list`
2. Traces `.envrc` `source_up` chain from CWD
3. Counts `op://` refs and detects literal URIs in environment
4. Returns system message to Claude with readiness summary

## Slash Command

`/joe-secrets:op-resolver` — invoke op-resolver on demand mid-session.

## Prerequisites

- 1Password CLI (`op`) installed and on PATH
- `direnv` installed (optional — chain tracing degrades gracefully without it)
- Install `joe-dev` for `handon` skill that session-start chains into
