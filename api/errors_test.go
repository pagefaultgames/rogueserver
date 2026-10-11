/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/pagefaultgames/rogueserver/api/savedata"
	"github.com/pagefaultgames/rogueserver/api/savedata/validate"
)

// TestStatusFor checks that each error category maps to its HTTP status, and
// that wrapping an error doesn't hide its category.
func TestStatusFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid save", validate.ErrNoPlaytime, http.StatusBadRequest},
		{"stale save", validate.ErrGreaterPlaytime, http.StatusBadRequest},
		{"wrapped stale save", fmt.Errorf("x: %w", validate.ErrWaveRegressed), http.StatusBadRequest},
		{"version parse failure", fmt.Errorf("%w: %w", validate.ErrVersionCompare, errors.New("bad")), http.StatusBadRequest},
		{"save not found", savedata.ErrSaveNotExist, http.StatusNotFound},
		{"unknown error", errors.New("db down"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFor(tt.err); got != tt.want {
				t.Errorf("statusFor(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
