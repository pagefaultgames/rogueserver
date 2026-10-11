/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import (
	"errors"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// TestPurity enforces the rule in the package doc: validate does no I/O.
// It fails if the package (or anything it pulls in) depends on a non-standard
// package other than defs, or directly imports a standard package that talks
// to the outside world.
func TestPurity(t *testing.T) {
	const module = "github.com/pagefaultgames/rogueserver/"
	allowedNonStd := []string{
		module + "api/savedata/validate",
		module + "defs",
	}
	// Checked against direct imports only: fmt and errors pull in os
	// transitively, which is fine.
	forbiddenStd := []string{
		"os", "os/exec", "io/fs", "syscall", "log", "log/slog",
		"net", "net/http", "database/sql",
	}

	for _, pkg := range goList(t, "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}") {
		if !slices.Contains(allowedNonStd, pkg) {
			t.Errorf("validate depends on %s; only defs and the standard library are allowed", pkg)
		}
	}
	for _, pkg := range goList(t, "-f", `{{join .Imports "\n"}}`) {
		if slices.Contains(forbiddenStd, pkg) {
			t.Errorf("validate imports %s; it must not do I/O", pkg)
		}
	}
}

// goList runs `go list <args> .` on this package and returns one entry per line.
func goList(t *testing.T, args ...string) []string {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	out, err := exec.Command(goBin, append(append([]string{"list"}, args...), ".")...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("go list: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("go list: %v", err)
	}
	return strings.Fields(string(out))
}
