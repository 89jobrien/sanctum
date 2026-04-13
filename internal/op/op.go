// Package op provides an interface to the 1Password CLI (op).
// The Client interface is the port; ExecClient is the real adapter.
// Inject fakes in tests — no live op binary required.
package op

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrBinaryMissing is returned when the op binary is not found on PATH.
var ErrBinaryMissing = errors.New("op binary not found on PATH")

// Account represents a signed-in 1Password account.
type Account struct {
	// URL is the account sign-in address, e.g. "my.1password.com".
	URL string
	// Shorthand is the short name used in op CLI flags.
	Shorthand string
}

// Client is the port for 1Password operations.
type Client interface {
	// AccountList returns all currently signed-in accounts.
	// Returns ErrBinaryMissing if op is not on PATH.
	AccountList(ctx context.Context) ([]Account, error)

	// ItemGet resolves a single op:// ref, returning nil on success.
	// ref must be a full op:// URI.
	// Returns ErrBinaryMissing if op is not on PATH.
	ItemGet(ctx context.Context, ref string) error
}

// ExecClient is the real Client implementation that shells out to op.
type ExecClient struct{}

// NewExecClient returns an ExecClient.
func NewExecClient() *ExecClient {
	return &ExecClient{}
}

// AccountList runs `op account list` and parses the output.
func (c *ExecClient) AccountList(ctx context.Context) ([]Account, error) {
	if err := checkBinary(); err != nil {
		return nil, err
	}
	out, err := exec.CommandContext(ctx, "op", "account", "list", "--format=json").Output()
	if err != nil {
		return nil, fmt.Errorf("op: %w", err)
	}
	return parseAccountList(out)
}

// ItemGet runs `op item get` to confirm the ref resolves.
func (c *ExecClient) ItemGet(ctx context.Context, ref string) error {
	if err := checkBinary(); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "op", "item", "get", ref) //nolint:gosec
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("op: %w", err)
	}
	return nil
}

// checkBinary verifies op is available on PATH.
func checkBinary() error {
	if _, err := exec.LookPath("op"); err != nil {
		return ErrBinaryMissing
	}
	return nil
}

// parseAccountList parses `op account list --format=json` output.
// op returns a JSON array of objects with "url" and "shorthand" keys.
func parseAccountList(data []byte) ([]Account, error) {
	// Minimal manual parse to avoid pulling in encoding/json at package level
	// for what is essentially a line-count operation. Full parse used here.
	raw := strings.TrimSpace(string(data))
	if raw == "[]" || raw == "" {
		return nil, nil
	}

	// Delegate to encoding/json — import is stdlib.
	return parseAccountJSON(data)
}
