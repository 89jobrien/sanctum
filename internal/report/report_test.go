package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestHumanClean(t *testing.T) {
	r := Report{OpAccounts: 2, EnvrcFiles: 3, OpRefs: 6}
	var buf bytes.Buffer
	Human(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "OK  2 account(s)") {
		t.Errorf("expected account status in output:\n%s", out)
	}
	if !strings.Contains(out, "3 file(s), 6 op:// ref(s)") {
		t.Errorf("expected chain stats in output:\n%s", out)
	}
	if !strings.Contains(out, "none") {
		t.Errorf("expected 'none' for conflicts/literal refs:\n%s", out)
	}
}

func TestHumanNoAccounts(t *testing.T) {
	r := Report{OpAccounts: 0}
	var buf bytes.Buffer
	Human(&buf, r)
	if !strings.Contains(buf.String(), "WARN") {
		t.Errorf("expected WARN when no accounts:\n%s", buf.String())
	}
}

func TestHumanConflicts(t *testing.T) {
	r := Report{
		OpAccounts: 1,
		Conflicts:  []string{"my.1password.com", "toptal.1password.com"},
	}
	var buf bytes.Buffer
	Human(&buf, r)
	if !strings.Contains(buf.String(), "toptal.1password.com") {
		t.Errorf("expected conflict hosts in output:\n%s", buf.String())
	}
}

func TestJSON(t *testing.T) {
	r := Report{OpAccounts: 2, EnvrcFiles: 1, OpRefs: 3, Conflicts: nil, LiteralRefs: nil}
	var buf bytes.Buffer
	if err := JSON(&buf, r); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.OpAccounts != 2 {
		t.Errorf("op_accounts: got %d, want 2", decoded.OpAccounts)
	}
}
