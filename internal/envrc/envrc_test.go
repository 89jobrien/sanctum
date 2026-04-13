package envrc

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestWalkOrder(t *testing.T) {
	home := t.TempDir()
	project := filepath.Join(home, "dev", "myproject")

	writeFile(t, filepath.Join(home, ".envrc"), "# home")
	writeFile(t, filepath.Join(home, "dev", ".envrc"), "# dev")
	writeFile(t, filepath.Join(project, ".envrc"), "# project")

	chain, err := Walk(project, home)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(chain.Files) != 3 {
		t.Fatalf("want 3 files, got %d: %v", len(chain.Files), chain.Files)
	}
	// innermost first
	if filepath.Base(filepath.Dir(chain.Files[0])) != "myproject" {
		t.Errorf("first file should be in myproject, got %s", chain.Files[0])
	}
}

func TestWalkNoFiles(t *testing.T) {
	home := t.TempDir()
	project := filepath.Join(home, "empty")
	if err := os.MkdirAll(project, 0o750); err != nil {
		t.Fatal(err)
	}

	chain, err := Walk(project, home)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(chain.Files) != 0 {
		t.Errorf("want 0 files, got %d", len(chain.Files))
	}
}

func TestExtract(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".envrc")
	writeFile(t, path, `
# comment — should be skipped
export FOO=$(op run -- echo op://my.1password.com/abc123/password)
export BAR="op://toptal.1password.com/xyz789/credential"
export SKIP="not-an-op-ref"
`)

	refs, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("want 2 refs, got %d: %v", len(refs), refs)
	}
	if refs[0].Vault != "my.1password.com" {
		t.Errorf("ref[0].Vault = %q", refs[0].Vault)
	}
	if refs[1].ItemID != "xyz789" {
		t.Errorf("ref[1].ItemID = %q", refs[1].ItemID)
	}
}

func TestIsNamedAccount(t *testing.T) {
	named := Ref{Vault: "my.1password.com"}
	uuid := Ref{Vault: "abc123def456"}
	if !named.IsNamedAccount() {
		t.Error("expected named account ref")
	}
	if uuid.IsNamedAccount() {
		t.Error("expected UUID ref to not be named account")
	}
}

func TestConflictsNone(t *testing.T) {
	refs := []Ref{
		{Vault: "my.1password.com"},
		{Vault: "my.1password.com"},
	}
	if c := Conflicts(refs); len(c) != 0 {
		t.Errorf("want no conflicts, got %v", c)
	}
}

func TestConflictsDetected(t *testing.T) {
	refs := []Ref{
		{Vault: "my.1password.com"},
		{Vault: "toptal.1password.com"},
	}
	if c := Conflicts(refs); len(c) != 2 {
		t.Errorf("want 2 conflict hosts, got %v", c)
	}
}

func TestConflictsUUIDNotCounted(t *testing.T) {
	// UUID-only refs alongside one named account should not trigger conflict.
	refs := []Ref{
		{Vault: "my.1password.com"},
		{Vault: "abc123def456"},
	}
	if c := Conflicts(refs); len(c) != 0 {
		t.Errorf("UUID refs should not trigger conflict, got %v", c)
	}
}
