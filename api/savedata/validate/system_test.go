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

// stats is a helper to create a gameStats object with a given playTime.
func stats(pt any) map[string]any { return map[string]any{"playTime": pt} }

// TestPlaytime enforces the behavior of [Playtime]'s validation logic,
// ensuring it returns the correct errors for missing or malformed playtime.
func TestPlaytime(t *testing.T) {
	ok := defs.SystemSaveData{GameStats: stats(100.0)}
	tests := []struct {
		name       string
		prev, next defs.SystemSaveData
		want       error
	}{
		{"equal", ok, ok, nil},
		{"higher", ok, defs.SystemSaveData{GameStats: stats(200.0)}, nil},
		{"lower", ok, defs.SystemSaveData{GameStats: stats(50.0)}, ErrGreaterPlaytime},
		{"next gameStats nil", ok, defs.SystemSaveData{}, ErrNoPlaytime},                      // used to panic
		{"next gameStats not object", ok, defs.SystemSaveData{GameStats: "x"}, ErrNoPlaytime}, // used to panic
		{"next playTime missing", ok, defs.SystemSaveData{GameStats: map[string]any{}}, ErrNoPlaytime},
		{"next playTime not number", ok, defs.SystemSaveData{GameStats: stats("100")}, ErrNoPlaytime},
		{"prev gameStats nil", defs.SystemSaveData{}, ok, ErrNoPlaytime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Playtime(tt.prev, tt.next); !errors.Is(err, tt.want) {
				t.Errorf("Playtime() = %v, want %v", err, tt.want)
			}
		})
	}
}

// save builds a system save with valid playtime and the given version and
// migrators, so version tests aren't tripped up by the playtime check.
func save(pt float64, version string, migrators map[string]int) defs.SystemSaveData {
	return defs.SystemSaveData{
		GameStats:        stats(pt),
		GameVersion:      version,
		AppliedMigrators: migrators,
	}
}

// TestVersion checks [Version]'s minimum-version, downgrade, parse and
// migrator rules.
func TestVersion(t *testing.T) {
	m := map[string]int{"a": 1}
	tests := []struct {
		name       string
		prev, next defs.SystemSaveData
		want       error
	}{
		{"same version", save(0, MinGameVersion, m), save(0, MinGameVersion, m), nil},
		{"upgrade", save(0, MinGameVersion, m), save(0, "1.12.1", m), nil},
		{"prev below min is fine", save(0, "1.11.0", nil), save(0, MinGameVersion, nil), nil},
		{"next below min", save(0, MinGameVersion, nil), save(0, "1.12.0.9", nil), ErrVersionTooLow},
		{"downgrade", save(0, "1.13.0", nil), save(0, "1.12.1", nil), ErrExistingVersionGreater},
		{"next malformed", save(0, MinGameVersion, nil), save(0, "abc", nil), ErrVersionCompare},
		{"next empty", save(0, MinGameVersion, nil), save(0, "", nil), ErrVersionCompare},
		{"prev malformed", save(0, "abc", nil), save(0, MinGameVersion, nil), ErrVersionCompare},
		{"migrators desynced", save(0, MinGameVersion, m), save(0, MinGameVersion, nil), ErrMigratorsDesynced},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Version(tt.prev, tt.next); !errors.Is(err, tt.want) {
				t.Errorf("Version() = %v, want %v", err, tt.want)
			}
		})
	}
}

// TestMigrators checks that next must keep every migrator in prev with the
// same timestamp, and may add new ones.
func TestMigrators(t *testing.T) {
	tests := []struct {
		name       string
		prev, next map[string]int
		want       error
	}{
		{"both empty", nil, nil, nil},
		{"identical", map[string]int{"a": 1}, map[string]int{"a": 1}, nil},
		{"extra in next", map[string]int{"a": 1}, map[string]int{"a": 1, "b": 2}, nil},
		{"missing from next", map[string]int{"a": 1}, map[string]int{}, ErrMigratorsDesynced},
		{"next nil", map[string]int{"a": 1}, nil, ErrMigratorsDesynced},
		{"timestamp changed", map[string]int{"a": 1}, map[string]int{"a": 2}, ErrMigratorsDesynced},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Migrators(tt.prev, tt.next); !errors.Is(err, tt.want) {
				t.Errorf("Migrators() = %v, want %v", err, tt.want)
			}
		})
	}
}

// TestIds checks that both stored IDs must match.
func TestIds(t *testing.T) {
	next := defs.SystemSaveData{TrainerId: 10, SecretId: 20}
	tests := []struct {
		name                string
		trainerId, secretId int
		want                error
	}{
		{"match", 10, 20, nil},
		{"trainer differs", 11, 20, ErrIdMismatch},
		{"secret differs", 10, 21, ErrIdMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Ids(tt.trainerId, tt.secretId, next); !errors.Is(err, tt.want) {
				t.Errorf("Ids() = %v, want %v", err, tt.want)
			}
		})
	}
}

// TestSystem checks that [System] runs playtime before version, and skips
// version checks in beta builds.
func TestSystem(t *testing.T) {
	// versionErr is what a version failure looks like through System: an
	// error normally, nothing in beta builds.
	versionErr := ErrVersionTooLow
	if defs.IsBeta {
		versionErr = nil
	}

	good := save(100, MinGameVersion, nil)
	tests := []struct {
		name       string
		prev, next defs.SystemSaveData
		want       error
	}{
		{"all good", good, save(200, MinGameVersion, nil), nil},
		{"playtime checked first", good, save(50, "1.0.0", nil), ErrGreaterPlaytime},
		{"version checked outside beta", good, save(200, "1.0.0", nil), versionErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := System(tt.prev, tt.next); !errors.Is(err, tt.want) {
				t.Errorf("System() = %v, want %v", err, tt.want)
			}
		})
	}
}
