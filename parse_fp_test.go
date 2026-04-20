package main

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
)

func TestParsePlasmidTypeFP(t *testing.T) {
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
			result := ParsePlasmidType(tt.input)
			val, err := E.Unwrap(result)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, val)
		})
	}
}
