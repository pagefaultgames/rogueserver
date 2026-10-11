/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

// ValidMigrators confirms that every migrator existing in the old save exists in the new save with identical timestamp.
func ValidMigrators(new map[string]int, old map[string]int) bool {
	for migrator, timestamp := range old {
		newTimestamp := new[migrator]
		if newTimestamp != timestamp {
			return false
		}
	}
	return true
}
