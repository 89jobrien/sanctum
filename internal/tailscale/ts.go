// Package tailscale provides host-reachability checks via the tailscale CLI.
// This is a v1 stub — the interface is defined but not implemented.
// Full implementation is deferred per the design spec.
package tailscale

import "context"

// Client is the port for Tailscale operations.
type Client interface {
	// Ping checks whether host is reachable via Tailscale.
	Ping(ctx context.Context, host string) error
}

// NoopClient satisfies Client but performs no operations.
// Used in v1 where Tailscale checks are deferred.
type NoopClient struct{}

// Ping always returns nil (no-op).
func (NoopClient) Ping(_ context.Context, _ string) error { return nil }
