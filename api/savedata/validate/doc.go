/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

// Package validate decides whether an incoming save may replace the stored one.
//
// Functions in this package are pure: they compare two saves
// and return an error describing why the new one is unacceptable,
// or nil.
//
// Nothing in this package reads or writes storage, the environment, or the network, so
// callers can run checks inside a conditional-write callback and tests need no
// fakes. Keep it that way: this package imports only defs and the standard
// library.
//
// # Errors
//
// Every rejection wraps exactly one category error, which callers use to
// choose a response:
//
//   - [ErrInvalidSave]: the save is malformed or impossible (missing playtime,
//     unparseable version, mismatched trainer IDs). Retrying won't help.
//   - [ErrStale]: the save is well-formed but older than what's stored (lower
//     playtime, earlier wave in the same run). The client should reload.
//
// Test for the category with errors.Is; the specific error (for example
// [ErrGreaterPlaytime]) is for logs.
//
// Error messages beginning with "session out of date" are matched by the
// client. Don't reword them without checking pokerogue.
package validate
