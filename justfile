# sanctum — 1Password and direnv session management plugin

# Set up local git hooks and install the plugin. Run once after cloning.
init:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "==> sanctum: plugin init"

    # 1. Wire local hooks
    git config core.hooksPath .githooks
    chmod +x .githooks/pre-commit .githooks/post-commit
    echo "    hooks: .githooks wired"

    # 2. Verify claude is available
    if ! command -v claude >/dev/null 2>&1; then
        echo "    ERROR: 'claude' not on PATH — install Claude Code first"
        echo "    https://claude.ai/code"
        exit 1
    fi

    # 3. Verify prerequisites
    if ! command -v op >/dev/null 2>&1; then
        echo "    WARNING: '1Password CLI (op)' not found"
        echo "    Install: brew install 1password-cli"
        echo "    Then sign in: op signin"
        printf "    Continue anyway? [y/N] "
        read -r ans
        [ "$ans" = "y" ] || [ "$ans" = "Y" ] || exit 1
    else
        echo "    op: $(op --version) found"
    fi

    if ! command -v direnv >/dev/null 2>&1; then
        echo "    NOTE: direnv not found — .envrc chain tracing will be skipped"
        echo "    Install: brew install direnv"
    else
        echo "    direnv: found"
    fi

    # 4. Register local marketplace if not already registered
    MARKETPLACE="$HOME/.claude/plugins/local-marketplace"
    if [ -d "$MARKETPLACE" ]; then
        claude plugin marketplace add "$MARKETPLACE" 2>/dev/null || true
        echo "    marketplace: local registered"
    else
        echo "    WARNING: local marketplace not found at $MARKETPLACE"
    fi

    # 5. Install / reinstall plugin
    claude plugin uninstall sanctum --force 2>/dev/null || true
    claude plugin install sanctum@local
    echo "    plugin: sanctum installed"

    echo ""
    echo "==> Done. Restart Claude Code to apply."

# Reinstall plugin without re-running full init
reinstall:
    #!/usr/bin/env bash
    claude plugin uninstall sanctum --force 2>/dev/null || true
    claude plugin install sanctum@local
    echo "[sanctum] reinstalled — restart Claude Code to apply"
