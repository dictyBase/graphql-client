package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func runListPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	filterPlasmidType, err := parsePlasmidType(cmd.String("type"))
	if err != nil {
		return err
	}

	client := newGraphQLClient(cmd.String("endpoint"))
	result, err := fetchPlasmids(
		ctx,
		client,
		0,
		cmd.Int("limit"),
		PlasmidListFilter{
			PlasmidType: filterPlasmidType,
		},
	)
	if err != nil {
		return fmt.Errorf("graphql query failed: %w", err)
	}

	displayResults(result)
	return nil
}
