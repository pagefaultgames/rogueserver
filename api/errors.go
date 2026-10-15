/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package api

import (
	"errors"
	"net/http"

	"github.com/pagefaultgames/rogueserver/api/savedata"
	"github.com/pagefaultgames/rogueserver/api/savedata/validate"
)

// statusFor returns the HTTP status for an error from save handling.
func statusFor(err error) int {
	switch {
	case errors.Is(err, validate.ErrInvalidSave), errors.Is(err, validate.ErrStale):
		return http.StatusBadRequest
	case errors.Is(err, savedata.ErrSaveNotExist):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
