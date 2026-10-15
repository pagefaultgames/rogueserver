/*
 * SPDX-FileCopyrightText: 2024-2026 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package validate

import "testing"

func Test_compareGameVersion(t *testing.T) {
	tests := []struct {
		v1      string
		v2      string
		want    int
		wantErr bool
	}{
		{
			v1:   "1.0.4",
			v2:   "1.0.4",
			want: 0,
		},
		{
			v1:   "1.0.4",
			v2:   "1.0.4.0",
			want: 0,
		},
		{
			v1:   "1.0.4.0",
			v2:   "1.0.4",
			want: 0,
		},
		{
			v1:   "1.0.4",
			v2:   "1.0.4.1",
			want: -1,
		},
		{
			v1:   "1.0.4.1",
			v2:   "1.0.4",
			want: 1,
		},
		{
			v1:   "1.12.0",
			v2:   "1.12.0.1",
			want: -1,
		},
		{
			v1:   "1.12.0.1",
			v2:   "1.10.0",
			want: 1,
		},
		{
			v1:   "1.9.9",
			v2:   "1.10.0.0",
			want: -1,
		},
		{
			v1:   "1.10.0.0",
			v2:   "1.9.9",
			want: 1,
		},
		{
			v1:   "1.11.15.0",
			v2:   "1.11.9.0",
			want: 1,
		},
		{
			v1:      "1.a.4",
			v2:      "1.0.4",
			wantErr: true,
		},
		{
			v1:      "1.0.4",
			v2:      "1.0.beta",
			wantErr: true,
		},
		{
			v1:      "1.0",
			v2:      "1.0.4",
			wantErr: true,
		},
		{
			v1:      "1.0.4",
			v2:      "1.0.4.1.2",
			wantErr: true,
		},
		{
			v1:      "1..4",
			v2:      "1.0.4",
			wantErr: true,
		},
		{
			v1:      "",
			v2:      "1.0.4",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		testName := tt.v1 + " vs " + tt.v2
		t.Run(testName, func(t *testing.T) {
			got, err := CompareGameVersion(tt.v1, tt.v2)
			if tt.wantErr {
				if err == nil {
					t.Errorf("compareGameVersion() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("compareGameVersion() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("compareGameVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
