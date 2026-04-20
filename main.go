package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const defaultLimit = 10

const defaultEndpoint = "https://graphql.dictybase.dev/graphql"

func main() {
	cmd := &cli.Command{
		Name:  "gql-plasmid",
		Usage: "CLI to query plasmid data from a GraphQL endpoint",
		Commands: []*cli.Command{
			{
				Name:   "list-plasmids",
				Usage:  "List plasmids from the GraphQL endpoint",
				Action: RunListPlasmidCLI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "endpoint",
						Value: defaultEndpoint,
						Usage: "GraphQL API endpoint URL",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: defaultLimit,
						Usage: "Number of plasmid entries to fetch",
					},
					&cli.StringFlag{
						Name:  "type",
						Value: string(PlasmidTypeAll),
						Usage: "Plasmid type to filter (ALL, REGULAR, GOLDEN_BRAID)",
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
