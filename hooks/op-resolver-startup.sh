#!/usr/bin/env bash
# op-resolver-startup.sh
# SessionStart hook: validate 1Password auth and trace direnv chain.
# Outputs a system message for Claude summarizing session readiness.
set -euo pipefail

OUTPUT=""
WARNINGS=""

# Step 1: Check 1Password auth
if ! op account list &>/dev/null; then
  WARNINGS="${WARNINGS}WARNING: 1Password CLI not authed — run 'op signin' before using op:// refs.\n"
else
  ACCOUNT_COUNT=$(op account list 2>/dev/null | tail -n +2 | wc -l | tr -d ' ')
  OUTPUT="${OUTPUT}1Password: ${ACCOUNT_COUNT} account(s) authed.\n"
fi

# Step 2: Trace .envrc chain from CWD
ENVRC_COUNT=0
OP_REF_COUNT=0
d="${CLAUDE_PROJECT_DIR:-$PWD}"
while [ "$d" != "$HOME" ] && [ "$d" != "/" ]; do
  if [ -f "$d/.envrc" ]; then
    ENVRC_COUNT=$((ENVRC_COUNT + 1))
    refs=$(grep -c "op://" "$d/.envrc" 2>/dev/null || echo 0)
    OP_REF_COUNT=$((OP_REF_COUNT + refs))
  fi
  d=$(dirname "$d")
done

OUTPUT="${OUTPUT}Direnv chain: ${ENVRC_COUNT} .envrc file(s) found, ${OP_REF_COUNT} op:// refs.\n"

# Step 3: Check for literal op:// URIs in environment
LITERAL_COUNT=$(env | grep -c "op://" 2>/dev/null || echo 0)
if [ "$LITERAL_COUNT" -gt 0 ]; then
  WARNINGS="${WARNINGS}WARNING: ${LITERAL_COUNT} env var(s) contain literal op:// URIs — use 'op run --' to resolve.\n"
fi

# Compose system message
MSG="[joe-secrets session-start]\n${OUTPUT}"
if [ -n "$WARNINGS" ]; then
  MSG="${MSG}${WARNINGS}"
fi

# Output JSON for Claude
python3 -c "
import json, sys
msg = sys.argv[1].replace('\\\\n', '\n').strip()
print(json.dumps({'systemMessage': msg}))
" "$MSG"
