/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import (
	"fmt"

	"github.com/pagefaultgames/rogueserver/defs"
)

// System reports whether next may replace prev as a player's system save.
// It checks playtime first, then (outside beta builds) game version and
// migrators, returning the first failure.
func System(prev, next defs.SystemSaveData) error {
	if err := Playtime(prev, next); err != nil {
		return err
	}
	if !defs.IsBeta {
		return Version(prev, next)
	}
	return nil
}

// playTime extracts gameStats.playTime, reporting false if gameStats is
// missing, isn't an object, or has no numeric playTime.
func playTime(data defs.SystemSaveData) (float64, bool) {
	stats, ok := data.GameStats.(map[string]any)
	if !ok {
		return 0, false
	}
	pt, ok := stats["playTime"].(float64)
	return pt, ok
}

// Playtime reports whether next has at least as much total playtime as prev.
//
// It returns [ErrNoPlaytime] if either save lacks gameStats.playTime, and
// [ErrGreaterPlaytime] if next has less.
func Playtime(prev, next defs.SystemSaveData) error {
	nextPT, ok := playTime(next)
	if !ok {
		return ErrNoPlaytime
	}
	prevPT, ok := playTime(prev)
	if !ok {
		return ErrNoPlaytime
	}
	if nextPT < prevPT {
		return ErrGreaterPlaytime
	}
	return nil
}

// Version reports whether next may replace prev based on game version and
// applied migrators.
//
// It returns [ErrVersionTooLow] if next is below
// [MinGameVersion], [ErrExistingVersionGreater] if prev is from a newer
// version, [ErrMigratorsDesynced] if migrators disagree, and an error
// wrapping [ErrVersionCompare] if either version can't be parsed.
func Version(prev, next defs.SystemSaveData) error {
	cmp, err := CompareGameVersion(MinGameVersion, next.GameVersion)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrVersionCompare, err)
	}
	if cmp > 0 {
		return ErrVersionTooLow
	}

	cmp, err = CompareGameVersion(prev.GameVersion, next.GameVersion)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrVersionCompare, err)
	}
	if cmp > 0 {
		return ErrExistingVersionGreater
	}

	return Migrators(prev.AppliedMigrators, next.AppliedMigrators)
}

// Migrators reports whether every migrator applied to prev is also applied
// to next with the same timestamp.
//
// Extra migrators in next are allowed, since
// a newer client may have applied new ones.
//
// It returns [ErrMigratorsDesynced]
// if any migrator from prev is missing from next or has a different timestamp.
func Migrators(prev, next map[string]int) error {
	for migrator, timestamp := range prev {
		if next[migrator] != timestamp {
			return ErrMigratorsDesynced
		}
	}
	return nil
}

// Ids reports whether next carries trainerId and secretId. It
// returns [ErrIdMismatch] if either differs.
func Ids(trainerId, secretId int, next defs.SystemSaveData) error {
	if next.TrainerId != trainerId || next.SecretId != secretId {
		return ErrIdMismatch
	}
	return nil
}
