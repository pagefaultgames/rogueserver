/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import (
	"errors"
	"fmt"
)

// Categories. Every rejection reason wraps exactly one of these; callers test
// for them with [errors.Is] to decide how to respond.
var (
	// ErrInvalidSave means the save is malformed or impossible. Retrying the
	// same save won't help.
	ErrInvalidSave = errors.New("invalid save")

	// ErrStale means the client's save disagrees with the stored one in a way
	// that reloading from the server fixes. The client should reload.
	//
	// The client reloads when an error message starts with this text, so
	// don't change it without updating pokerogue.
	ErrStale = errors.New("session out of date")
)

// Rejection reasons. Each message is its category's text followed by the
// specific reason, and is sent to the client as is.
var (
	// ErrNoPlaytime is returned when a save has no gameStats.playTime,
	// indicating corruption or save tampering.
	ErrNoPlaytime = fmt.Errorf("%w: no playtime found", ErrInvalidSave)

	// ErrVersionCompare is returned when a game version is missing or
	// malformed. Callers wrap the underlying parse error after it.
	ErrVersionCompare = fmt.Errorf("%w: failed to compare versions", ErrInvalidSave)

	// ErrIdMismatch is returned when the save's trainer or secret ID does not
	// match the stored one.
	ErrIdMismatch = fmt.Errorf("%w: stored trainer or secret ID does not match", ErrStale)

	// ErrGreaterPlaytime is returned when the stored save has more playtime
	// than the incoming one.
	ErrGreaterPlaytime = fmt.Errorf("%w: existing playtime is greater", ErrStale)

	// ErrVersionTooLow is returned when the incoming save's game version is
	// below [MinGameVersion], indicating outdated client code.
	ErrVersionTooLow = fmt.Errorf("%w: save version below minimum game version", ErrStale)

	// ErrExistingVersionGreater is returned when the stored save is from a
	// newer game version than the incoming one, indicating outdated client code.
	ErrExistingVersionGreater = fmt.Errorf("%w: existing version is greater", ErrStale)

	// ErrMigratorsDesynced is returned when the incoming save is missing a
	// migrator that the stored save has, or has it with a different timestamp.
	ErrMigratorsDesynced = fmt.Errorf("%w: migrators desynced", ErrStale)

	// ErrWaveRegressed is returned when an incoming session from the same run
	// (same seed) is at an earlier wave than the stored one.
	ErrWaveRegressed = fmt.Errorf("%w: existing wave index is greater", ErrStale)
)
