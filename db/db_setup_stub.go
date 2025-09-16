//go:build !devsetup
// +build !devsetup

package db

import "database/sql"

// MaybeSetupDb is called by db.go and does nothing in non-devsetup builds.
func MaybeSetupDb(db *sql.DB) error {
	return nil
}
