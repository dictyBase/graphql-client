package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const defaultLimit = 10

const defaultEndpoint = "https://graphql.dictybase.dev/graphql"

func newRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "graphql-client",
		Usage: "CLI to query stock data from a GraphQL endpoint",
		Commands: []*cli.Command{
			newListPlasmidsCommand(),
			newListFilteredPlasmidsCommand(),
			newListStrainsCommand(),
			newListFilteredStrainsCommand(),
		},
	}
}

func newListPlasmidsCommand() *cli.Command {
	return &cli.Command{
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
	}
}

func newListFilteredPlasmidsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-filtered-plasmids",
		Usage:  "List plasmids with attribute-level filtering from the GraphQL endpoint",
		Action: RunListFilteredPlasmidCLI,
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
			&cli.StringFlag{
				Name:  "name",
				Usage: "Filter by plasmid name",
			},
			&cli.StringFlag{
				Name:  "summary",
				Usage: "Filter by plasmid summary",
			},
		},
	}
}

func newListStrainsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-strains",
		Usage:  "List strains from the GraphQL endpoint",
		Action: RunListStrainCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "endpoint",
				Value: defaultEndpoint,
				Usage: "GraphQL API endpoint URL",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: defaultLimit,
				Usage: "Number of strain entries to fetch",
			},
			&cli.StringFlag{
				Name:  "type",
				Value: string(StrainTypeAll),
				Usage: "Strain type to filter (ALL, REGULAR, GWDI, BACTERIAL)",
			},
		},
	}
}

func newListFilteredStrainsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-filtered-strains",
		Usage:  "List strains with attribute-level filtering from the GraphQL endpoint",
		Action: RunListFilteredStrainCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "endpoint",
				Value: defaultEndpoint,
				Usage: "GraphQL API endpoint URL",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: defaultLimit,
				Usage: "Number of strain entries to fetch",
			},
			&cli.StringFlag{
				Name:  "type",
				Value: string(StrainTypeAll),
				Usage: "Strain type to filter (ALL, REGULAR, GWDI, BACTERIAL)",
			},
			&cli.StringFlag{
				Name:  "label",
				Usage: "Filter by strain label",
			},
			&cli.StringFlag{
				Name:  "summary",
				Usage: "Filter by strain summary",
			},
		},
	}
}

func main() {
	if err := newRootCommand().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
