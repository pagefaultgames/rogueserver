/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import (
	"errors"
	"testing"
)

// TestReasons tests that the error messages and categories are correct for each error.
//
// Intended to prevent unintentional drift between the error messages and categories
// that are relied upon by the PokeRogue client code.
func TestReasons(t *testing.T) {
	tests := []struct {
		err      error
		msg      string
		category error
	}{
		{ErrNoPlaytime, "invalid save: no playtime found", ErrInvalidSave},
		{ErrVersionCompare, "invalid save: failed to compare versions", ErrInvalidSave},
		{ErrIdMismatch, "session out of date: stored trainer or secret ID does not match", ErrStale},
		{ErrGreaterPlaytime, "session out of date: existing playtime is greater", ErrStale},
		{ErrVersionTooLow, "session out of date: save version below minimum game version", ErrStale},
		{ErrExistingVersionGreater, "session out of date: existing version is greater", ErrStale},
		{ErrMigratorsDesynced, "session out of date: migrators desynced", ErrStale},
		{ErrWaveRegressed, "session out of date: existing wave index is greater", ErrStale},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.msg {
				t.Errorf("message = %q, want %q (the client matches these)", got, tt.msg)
			}
			for _, cat := range []error{ErrInvalidSave, ErrStale} {
				if want := cat == tt.category; errors.Is(tt.err, cat) != want {
					t.Errorf("errors.Is(%v) = %v, want %v", cat, !want, want)
				}
			}
		})
	}
}
