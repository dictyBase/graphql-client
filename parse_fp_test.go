package main

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
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

func TestStringPtr(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		result := stringPtr("")
		require.Nil(t, result)
	})

	t.Run("non-empty string returns pointer", func(t *testing.T) {
		result := stringPtr("pDM123")
		require.NotNil(t, result)
		require.Equal(t, "pDM123", *result)
	})
}

func TestBuildPlasmidAttributeFilter(t *testing.T) {
	t.Run("only plasmid type, no attribute filters", func(t *testing.T) {
		cmd := buildTestCommand(t, nil)
		filter := buildPlasmidAttributeFilter(PlasmidTypeAll, cmd)
		require.Equal(t, PlasmidTypeAll, filter.PlasmidType)
		require.Nil(t, filter.Name)
		require.Nil(t, filter.Summary)
	})

	t.Run("all attribute filters set", func(t *testing.T) {
		cmd := buildTestCommand(t, map[string]string{
			"name":    "pDM123",
			"summary": "expression vector",
		})
		filter := buildPlasmidAttributeFilter(PlasmidTypeRegular, cmd)
		require.Equal(t, PlasmidTypeRegular, filter.PlasmidType)
		require.NotNil(t, filter.Name)
		require.Equal(t, "pDM123", *filter.Name)
		require.NotNil(t, filter.Summary)
		require.Equal(t, "expression vector", *filter.Summary)
	})

	t.Run("partial attribute filters", func(t *testing.T) {
		cmd := buildTestCommand(t, map[string]string{
			"name": "pDM",
		})
		filter := buildPlasmidAttributeFilter(PlasmidTypeGoldenBraid, cmd)
		require.Equal(t, PlasmidTypeGoldenBraid, filter.PlasmidType)
		require.NotNil(t, filter.Name)
		require.Equal(t, "pDM", *filter.Name)
		require.Nil(t, filter.Summary)
	})
}

func buildTestCommand(t *testing.T, flags map[string]string) *cli.Command {
	t.Helper()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
			&cli.StringFlag{Name: "summary"},
		},
	}
	for k, v := range flags {
		switch k {
		case "name":
			cmd.Set("name", v)
		case "summary":
			cmd.Set("summary", v)
		}
	}
	return cmd
}
