// Command sanctum is the 1Password + direnv session diagnostic tool.
// It replaces inline bash in the sanctum plugin's SessionStart hook and
// provides on-demand diagnostics for the direnv → 1Password secret chain.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/89jobrien/sanctum/internal/envrc"
	"github.com/89jobrien/sanctum/internal/op"
	"github.com/89jobrien/sanctum/internal/report"
)

func main() {
	os.Exit(run())
}

// run returns an exit code. All errors are printed to stderr before returning.
// Using a separate function keeps os.Exit testable (defers run before Exit).
func run() int {
	opClient := op.NewExecClient()

	root := &cobra.Command{
		Use:           "sanctum",
		Short:         "1Password + direnv session diagnostics",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newCheckCmd(opClient),
		newTraceCmd(),
		newValidateCmd(opClient),
		newScaffoldCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "sanctum: %v\n", err)
		return 1
	}
	return 0
}

// Config carries shared options across subcommands.
type Config struct {
	Dir   string
	JSON  bool
	Quiet bool
}

// newCheckCmd builds the `sanctum check` subcommand.
func newCheckCmd(opClient op.Client) *cobra.Command {
	cfg := &Config{}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run full session-start diagnostics",
		Long: `check validates the 1Password auth state and scans the .envrc chain
for op:// secret references, conflicts, and unresolved literal refs.

Exit codes:
  0  clean — no issues found
  1  warnings — conflicts or literal refs detected
  2  fatal — op not authed or binary missing`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			code, err := runCheck(cmd.Context(), cfg, opClient)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sanctum check: %v\n", err)
			}
			// Signal the exit code via a dedicated mechanism so cobra
			// doesn't suppress it. We return nil to silence cobra's own
			// error printing; main() picks up the code via runCheck.
			os.Exit(code) //nolint:revive
			return nil
		},
	}

	home, _ := os.UserHomeDir()
	cmd.Flags().StringVar(&cfg.Dir, "dir", ".", "directory to start .envrc chain traversal")
	cmd.Flags().BoolVar(&cfg.JSON, "json", false, "output machine-readable JSON")
	cmd.Flags().BoolVar(&cfg.Quiet, "quiet", false, "suppress output; use exit code only")
	_ = home

	return cmd
}

// runCheck orchestrates phases 1-3 and returns an exit code.
func runCheck(ctx context.Context, cfg *Config, opClient op.Client) (int, error) {
	dir := cfg.Dir
	if dir == "" || dir == "." {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return 2, fmt.Errorf("get working directory: %w", err)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return 2, fmt.Errorf("get home directory: %w", err)
	}

	// Step 1: 1Password account check.
	accounts, err := opClient.AccountList(ctx)
	if err != nil {
		if !cfg.Quiet {
			fmt.Fprintf(os.Stderr, "sanctum: op auth check failed: %v\n", err)
		}
		return 2, nil //nolint:nilerr // fatal exit, error already printed
	}

	// Step 2-3: .envrc chain traversal + ref extraction.
	chain, err := envrc.Walk(dir, home)
	if err != nil {
		return 2, fmt.Errorf("envrc walk: %w", err)
	}

	// Step 4: conflict detection + literal ref scan.
	conflicts := envrc.Conflicts(chain.Refs)
	literalRefs := envrc.LiteralRefs()

	r := report.Report{
		OpAccounts:  len(accounts),
		EnvrcFiles:  len(chain.Files),
		OpRefs:      len(chain.Refs),
		Conflicts:   conflicts,
		LiteralRefs: literalRefs,
	}

	// Step 5: output.
	if !cfg.Quiet {
		w := os.Stdout
		if cfg.JSON {
			if err := report.JSON(w, r); err != nil {
				return 2, fmt.Errorf("json output: %w", err)
			}
		} else {
			report.Human(w, r)
		}
	}

	// Determine exit code.
	if len(conflicts) > 0 || len(literalRefs) > 0 {
		return 1, nil
	}
	return 0, nil
}

// newTraceCmd builds the `sanctum trace` subcommand (stub).
func newTraceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trace [dir]",
		Short: "Trace .envrc chain and list all op:// refs",
		RunE: func(_ *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			home, _ := os.UserHomeDir()
			chain, err := envrc.Walk(dir, home)
			if err != nil {
				return err
			}
			for _, f := range chain.Files {
				fmt.Println(f)
				for _, r := range chain.Refs {
					if r.File == f {
						flag := ""
						if !r.IsNamedAccount() {
							flag = "  (UUID-only ref)"
						}
						fmt.Printf("  %-60s  → %s%s\n", r.Raw, r.Vault, flag)
					}
				}
			}
			return nil
		},
	}
}

// newValidateCmd builds the `sanctum validate` subcommand (stub).
func newValidateCmd(opClient op.Client) *cobra.Command {
	var account string
	cmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Resolve each op:// ref and report pass/fail",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			home, _ := os.UserHomeDir()
			chain, err := envrc.Walk(dir, home)
			if err != nil {
				return err
			}
			anyFail := false
			for _, r := range chain.Refs {
				if account != "" && r.Vault != account {
					continue
				}
				if err := opClient.ItemGet(cmd.Context(), r.Raw); err != nil {
					fmt.Printf("FAIL  %s — %v\n", r.Raw, err)
					anyFail = true
				} else {
					fmt.Printf("PASS  %s\n", r.Raw)
				}
			}
			if anyFail {
				os.Exit(1)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&account, "account", "", "filter to refs for this account shorthand")
	return cmd
}

// newScaffoldCmd builds the `sanctum scaffold` subcommand (stub).
func newScaffoldCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "scaffold [dir]",
		Short: "Write a starter .envrc in dir",
		RunE: func(_ *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			target := dir + "/.envrc"
			if !force {
				if _, err := os.Stat(target); err == nil {
					return fmt.Errorf("%s already exists — use --force to overwrite", target)
				}
			}
			template := `# .envrc — managed by sanctum
# Secrets loaded via 1Password CLI
# Edit op:// refs then run: direnv allow

dotenv_if_exists .env.local

# Inject secrets via op run — add vars to ~/.secrets then:
# op run --env-file=~/.secrets -- direnv exec . <command>
`
			return os.WriteFile(target, []byte(template), 0o600)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing .envrc")
	return cmd
}
