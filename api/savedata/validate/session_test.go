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

	"github.com/pagefaultgames/rogueserver/defs"
)

// run builds a session save with the given seed and wave.
func run(seed string, wave int) defs.SessionSaveData {
	return defs.SessionSaveData{Seed: seed, WaveIndex: wave}
}

// TestSession checks that the wave can't go backwards within a run, and that
// a different seed is treated as a new run.
func TestSession(t *testing.T) {
	tests := []struct {
		name       string
		prev, next defs.SessionSaveData
		want       error
	}{
		{"same wave", run("a", 10), run("a", 10), nil},
		{"forward", run("a", 10), run("a", 11), nil},
		{"backwards in same run", run("a", 10), run("a", 9), ErrWaveRegressed},
		{"new run at lower wave", run("a", 10), run("b", 1), nil},
		{"empty prev", defs.SessionSaveData{}, run("a", 1), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Session(tt.prev, tt.next); !errors.Is(err, tt.want) {
				t.Errorf("Session() = %v, want %v", err, tt.want)
			}
		})
	}
}
