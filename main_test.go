package main

import "testing"

func TestParsePlasmidType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PlasmidType
		wantErr bool
	}{
		{
			name:  "ALL type",
			input: "ALL",
			want:  PlasmidTypeAll,
		},
		{
			name:  "REGULAR type",
			input: "REGULAR",
			want:  PlasmidTypeRegular,
		},
		{
			name:  "GOLDEN_BRAID type",
			input: "GOLDEN_BRAID",
			want:  PlasmidTypeGoldenBraid,
		},
		{
			name:    "empty string is invalid",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "INVALID",
			wantErr: true,
		},
		{
			name:    "lowercase is invalid",
			input:   "all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePlasmidType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePlasmidType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parsePlasmidType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
