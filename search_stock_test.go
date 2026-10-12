package main

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// searchStockTestCommand builds a search-stocks command whose flags carry
// valid default values.
func searchStockTestCommand() *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagQuery, Value: "pDM"},
			&cli.IntFlag{Name: flagLimit, Value: defaultLimit},
			&cli.StringFlag{Name: flagEntity, Value: string(StockEntityAll)},
			&cli.StringFlag{Name: flagEndpoint, Value: "https://unit.test/graphql"},
		},
	}
}

func searchStockTestArgs() searchStockCLIArgs {
	return searchStockCLIArgs{Context: context.Background(), Command: searchStockTestCommand()}
}

func TestParseStockEntityTypeFP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StockEntityType
		wantErr bool
	}{
		{
			name:  "ALL entity",
			input: string(StockEntityAll),
			want:  StockEntityAll,
		},
		{
			name:  "STRAIN entity",
			input: string(StockEntityStrain),
			want:  StockEntityStrain,
		},
		{
			name:  "PLASMID entity",
			input: "PLASMID",
			want:  StockEntityPlasmid,
		},
		{
			name:    "empty entity is invalid",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid entity",
			input:   "GWDI",
			wantErr: true,
		},
		{
			name:    "lowercase entity is invalid",
			input:   "strain",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseStockEntityType(tt.input)
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

func TestValidateStockQuery(t *testing.T) {
	t.Run("too short query fails", func(t *testing.T) {
		args := searchStockTestArgs()
		require.NoError(t, args.Command.Set(flagQuery, "p"))
		_, err := E.Unwrap(validateStockQuery(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least 2 characters")
	})

	t.Run("minimum length query passes", func(t *testing.T) {
		args := searchStockTestArgs()
		require.NoError(t, args.Command.Set(flagQuery, "DB"))
		result, err := E.Unwrap(validateStockQuery(args))
		require.NoError(t, err)
		require.Equal(t, args, result)
	})

	t.Run("normal query passes through the state", func(t *testing.T) {
		args := searchStockTestArgs()
		result, err := E.Unwrap(validateStockQuery(args))
		require.NoError(t, err)
		require.Equal(t, args, result)
	})
}

func TestParseStockEntity(t *testing.T) {
	t.Run("invalid entity fails", func(t *testing.T) {
		args := searchStockTestArgs()
		require.NoError(t, args.Command.Set(flagEntity, "ORGANISM"))
		_, err := E.Unwrap(parseStockEntity(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "entity")
	})

	t.Run("valid entity passes through the state", func(t *testing.T) {
		args := searchStockTestArgs()
		result, err := E.Unwrap(parseStockEntity(args))
		require.NoError(t, err)
		require.Equal(t, args, result)
	})
}

func TestToSearchStockFetchArgs(t *testing.T) {
	args := searchStockTestArgs()
	require.NoError(t, args.Command.Set(flagEntity, "STRAIN"))
	require.NoError(t, args.Command.Set(flagLimit, "25"))

	fetch := toSearchStockFetchArgs(args)
	cmd := args.Command
	require.NotNil(t, fetch.Client)
	require.Equal(t, cmd.String(flagQuery), fetch.Query)
	require.Equal(t, 25, fetch.Limit)
	require.Equal(t, StockEntityStrain, fetch.Entity)
	require.Equal(t, args.Context, fetch.Ctx)
}
