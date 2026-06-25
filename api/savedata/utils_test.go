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

package savedata

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
