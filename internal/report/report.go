// Package report formats sanctum diagnostic output.
// Human() writes a fixed-width table; JSON() writes compact JSON.
// No business logic lives here — pure formatting.
package report

import (
	"encoding/json"
	"fmt"
	"io"
)

// Report is the structured result of a sanctum check run.
type Report struct {
	OpAccounts  int      `json:"op_accounts"`
	EnvrcFiles  int      `json:"envrc_files"`
	OpRefs      int      `json:"op_refs"`
	Conflicts   []string `json:"conflicts"`
	LiteralRefs []string `json:"literal_refs"`
}

const colWidth = 14

// Human writes a fixed-width human-readable summary table to w.
// Errors go to w (not stderr) since this is a formatting function;
// callers control the destination writer.
func Human(w io.Writer, r Report) {
	status := func(ok bool, okMsg, failMsg string) string {
		if ok {
			return okMsg
		}
		return failMsg
	}

	accountStatus := status(r.OpAccounts > 0,
		fmt.Sprintf("OK  %d account(s)", r.OpAccounts),
		"WARN  no accounts signed in",
	)

	fmt.Fprintf(w, "%-*s  %s\n", colWidth, "1PASSWORD", accountStatus)
	fmt.Fprintf(w, "%-*s  %d file(s), %d op:// ref(s)\n",
		colWidth, "DIRENV CHAIN", r.EnvrcFiles, r.OpRefs)

	if len(r.Conflicts) == 0 {
		fmt.Fprintf(w, "%-*s  none\n", colWidth, "CONFLICTS")
	} else {
		fmt.Fprintf(w, "%-*s  %v\n", colWidth, "CONFLICTS", r.Conflicts)
	}

	if len(r.LiteralRefs) == 0 {
		fmt.Fprintf(w, "%-*s  none\n", colWidth, "LITERAL REFS")
	} else {
		fmt.Fprintf(w, "%-*s  %d found\n", colWidth, "LITERAL REFS", len(r.LiteralRefs))
		for _, ref := range r.LiteralRefs {
			fmt.Fprintf(w, "%*s  %s\n", colWidth, "", ref)
		}
	}
}

// JSON writes a compact JSON representation of r to w.
func JSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "")
	return enc.Encode(r)
}
