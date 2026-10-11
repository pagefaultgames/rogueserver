/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import "errors"

// Categories. Every rejection reason wraps exactly one of these; callers test
// for them with errors.Is to decide how to respond.
var (
	// ErrInvalidSave means the save is malformed or impossible. Retrying the
	// same save won't help.
	ErrInvalidSave = errors.New("invalid save")

	// ErrStale means the save is well-formed but older than the stored one.
	// The client should reload.
	ErrStale = errors.New("session out of date")
)

// Rejection reasons. Each message is sent to the client as is, and messages
// beginning with "session out of date" are matched by the client
//
// don't reword them without verifying against what the pokerogue repository expects.
var (
	// ErrNoPlaytime is returned when a save has no gameStats.playTime,
	// indicating corruption or save tampering.
	ErrNoPlaytime error = &reason{"no playtime found", ErrInvalidSave}

	// ErrIdMismatch is returned when the save's trainer or secret ID does not
	// match the stored one, indicating corruption or save tampering.
	ErrIdMismatch error = &reason{"session out of date: stored trainer or secret ID does not match", ErrInvalidSave}

	// ErrVersionCompare is returned when a game version is missing or
	// malformed. Callers wrap the underlying parse error after it.
	ErrVersionCompare error = &reason{"failed to compare versions", ErrInvalidSave}

	// ErrGreaterPlaytime is returned when the stored save has more playtime
	// than the incoming one.
	ErrGreaterPlaytime error = &reason{"session out of date: existing playtime is greater", ErrStale}

	// ErrVersionTooLow is returned when the incoming save's game version is
	// below [MinGameVersion], indicating outdated client code.
	ErrVersionTooLow error = &reason{"session out of date: save version below minimum game version", ErrStale}

	// ErrExistingVersionGreater is returned when the stored save is from a
	// newer game version than the incoming one, indicating outdated client code.
	ErrExistingVersionGreater error = &reason{"session out of date: existing version is greater", ErrStale}

	// ErrMigratorsDesynced is returned when the incoming save is missing a
	// migrator that the stored save has, or has it with a different timestamp.
	ErrMigratorsDesynced error = &reason{"session out of date: migrators desynced", ErrStale}

	// ErrWaveRegressed is returned when an incoming session from the same run
	// (same seed) is at an earlier wave than the stored one.
	ErrWaveRegressed error = &reason{"session out of date: existing wave index is greater", ErrStale}
)

// reason is a rejection whose message is shown as is but which unwraps to
// its category, so errors.Is matches both the reason and the category.
type reason struct {
	msg      string
	category error
}

func (r *reason) Error() string { return r.msg }
func (r *reason) Unwrap() error { return r.category }
