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
//     unparseable version). Retrying won't help.
//   - [ErrStale]: the client's save disagrees with what's stored in a way that
//     reloading fixes (lower playtime, earlier wave in the same run,
//     mismatched trainer IDs). The client should reload.
//
// Test for the category with errors.Is; the specific error (for example
// [ErrGreaterPlaytime]) is for logs.
//
// Each error message starts with its category's text. The client reloads when
// a message starts with [ErrStale]'s text, so don't change it without
// updating pokerogue.
package validate
