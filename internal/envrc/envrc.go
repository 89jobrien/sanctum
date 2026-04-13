// Package envrc traverses the .envrc chain from a directory up to $HOME
// and extracts op:// secret references from each file.
package envrc

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// refPattern matches op:// URIs in .envrc files.
// Groups: 1=vault-or-host, 2=item-uuid, 3=field
// Handles both named-account refs (op://host/uuid/field) and
// UUID-only refs (op://uuid/uuid/field).
var refPattern = regexp.MustCompile(
	`op://([^/\s]+)/([^/\s]+)/([^\s"']+)`,
)

// Ref represents a single op:// reference found in a .envrc file.
type Ref struct {
	// File is the absolute path of the .envrc containing this ref.
	File string
	// Raw is the full op:// string as written.
	Raw string
	// Vault is the first path segment — either a hostname or a UUID.
	Vault string
	// ItemID is the second path segment (item UUID).
	ItemID string
	// Field is the third path segment (field name).
	Field string
}

// IsNamedAccount returns true when Vault looks like a hostname
// (contains a dot), distinguishing it from a bare UUID.
func (r Ref) IsNamedAccount() bool {
	return strings.Contains(r.Vault, ".")
}

// Chain holds the ordered list of .envrc files from the target directory
// up to $HOME, and all refs extracted from them.
type Chain struct {
	// Files is the ordered list of .envrc paths (innermost first).
	Files []string
	// Refs is the flat list of all op:// refs across all files.
	Refs []Ref
}

// Walk collects .envrc files from dir up to (and including) home,
// then extracts all op:// refs. home should be os.UserHomeDir().
// Returns an empty Chain (not an error) if no .envrc files are found.
func Walk(dir, home string) (Chain, error) {
	files, err := walkFiles(dir, home)
	if err != nil {
		return Chain{}, err
	}

	var allRefs []Ref
	for _, f := range files {
		refs, err := Extract(f)
		if err != nil {
			// Non-fatal: skip unreadable files.
			continue
		}
		allRefs = append(allRefs, refs...)
	}

	return Chain{Files: files, Refs: allRefs}, nil
}

// walkFiles returns absolute paths of .envrc files from dir up to home.
func walkFiles(dir, home string) ([]string, error) {
	// Normalise to absolute paths for comparison.
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	home, err = filepath.Abs(home)
	if err != nil {
		return nil, err
	}

	var files []string
	current := dir
	for {
		candidate := filepath.Join(current, ".envrc")
		if _, err := os.Stat(candidate); err == nil {
			files = append(files, candidate)
		}

		if current == home {
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			// Filesystem root reached before home.
			break
		}
		current = parent
	}
	return files, nil
}

// Extract scans a single .envrc file and returns all op:// refs found.
func Extract(path string) ([]Ref, error) {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var refs []Ref
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comment lines.
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		matches := refPattern.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			refs = append(refs, Ref{
				File:   path,
				Raw:    m[0],
				Vault:  m[1],
				ItemID: m[2],
				Field:  m[3],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return refs, nil
}

// Conflicts returns named-account hostnames that conflict within the chain.
// A conflict is defined as more than one distinct named-account hostname
// appearing across all refs in the chain.
func Conflicts(refs []Ref) []string {
	seen := map[string]struct{}{}
	for _, r := range refs {
		if r.IsNamedAccount() {
			seen[r.Vault] = struct{}{}
		}
	}
	if len(seen) <= 1 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for host := range seen {
		out = append(out, host)
	}
	return out
}

// LiteralRefs returns any op:// URI strings found literally in the
// current process environment — indicating unresolved references.
func LiteralRefs() []string {
	var found []string
	for _, kv := range os.Environ() {
		if _, val, ok := strings.Cut(kv, "="); ok {
			if strings.HasPrefix(val, "op://") {
				found = append(found, kv)
			}
		}
	}
	return found
}
