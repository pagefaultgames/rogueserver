/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import "github.com/pagefaultgames/rogueserver/defs"

// Session reports whether next may replace prev in a session slot.
//
// Within the same run (same seed) the wave can't go backwards; it returns
// [ErrWaveRegressed] if it does. A different seed is assumed to be a new run
// and is always accepted.
func Session(prev, next defs.SessionSaveData) error {
	// TODO: Add more robust checks that validate sessions to prevent corruption.
	if prev.Seed == next.Seed && next.WaveIndex < prev.WaveIndex {
		return ErrWaveRegressed
	}
	return nil
}
