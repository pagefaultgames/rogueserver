/*
	Copyright (C) 2024 - 2025  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package defs

import (
	"bytes"
	"encoding/gob"
	"testing"
)

// System save data is persisted as zstd-compressed gob (see db/savedata.go), not JSON.
// StarterAttributes uses pointer fields so that "unset" and "set to the zero value" stay
// distinguishable over JSON — but gob has its own rules about nil pointers, and every player's
// save goes through it. These tests exist to prove the new field survives that encoder.

func TestStarterPreferencesGobRoundTrip(t *testing.T) {
	nickname := "lil man der"
	favorite := true
	notFavorite := false
	form := 0

	original := SystemSaveData{
		StarterPreferences: StarterPreferences{
			// only nickname set - every other field is a nil pointer
			4: {Nickname: &nickname},
			// a deliberate `false`, which must not be confused with "unset"
			7: {Favorite: &notFavorite},
			// a deliberate zero, same concern
			25: {Form: &form, Favorite: &favorite},
		},
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(original); err != nil {
		t.Fatalf("gob encode failed with nil pointer fields present: %v", err)
	}

	var decoded SystemSaveData
	if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("gob decode failed: %v", err)
	}

	if got := decoded.StarterPreferences[4].Nickname; got == nil || *got != nickname {
		t.Errorf("nickname did not survive round trip: got %v, want %q", got, nickname)
	}

	// The important one: a stored `false` must come back as `false`, not as nil. If this fails,
	// un-favoriting a starter would silently revert on the next load.
	if got := decoded.StarterPreferences[7].Favorite; got == nil {
		t.Error("favorite=false came back as nil - a removal would be silently discarded")
	} else if *got != false {
		t.Errorf("favorite did not survive round trip: got %v, want false", *got)
	}

	// Same concern for a numeric zero, which is a legitimate form index.
	if got := decoded.StarterPreferences[25].Form; got == nil {
		t.Error("form=0 came back as nil - the default form would be indistinguishable from unset")
	} else if *got != 0 {
		t.Errorf("form did not survive round trip: got %v, want 0", *got)
	}

	// Unset fields must stay unset rather than materializing as zero values.
	if decoded.StarterPreferences[4].Favorite != nil {
		t.Error("favorite was never set on species 4 but came back non-nil")
	}
}

// legacySystemSaveData stands in for a save written before starter preferences existed.
// gob matches fields by name, so encoding this and decoding into the current SystemSaveData is
// the same shape as a real player loading an old save after this change ships.
type legacySystemSaveData struct {
	TrainerId   int
	SecretId    int
	StarterData StarterData
}

func TestOldSavesDecodeWithoutStarterPreferences(t *testing.T) {
	legacy := legacySystemSaveData{
		TrainerId:   1234,
		SecretId:    5678,
		StarterData: StarterData{4: {CandyCount: 10}},
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(legacy); err != nil {
		t.Fatalf("failed to encode legacy save: %v", err)
	}

	var decoded SystemSaveData
	if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("a save written before this change failed to decode: %v", err)
	}

	if decoded.TrainerId != 1234 {
		t.Errorf("trainerId lost: got %d, want 1234", decoded.TrainerId)
	}
	if decoded.StarterPreferences != nil && len(decoded.StarterPreferences) != 0 {
		t.Errorf("expected empty starter preferences on a legacy save, got %v", decoded.StarterPreferences)
	}
}
