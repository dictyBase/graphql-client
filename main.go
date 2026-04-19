package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"
)

const (
	tabPadding   = 3
	defaultLimit = 10
)

const (
	PlasmidTypeAll         PlasmidType = "ALL"
	PlasmidTypeRegular     PlasmidType = "REGULAR"
	PlasmidTypeGoldenBraid PlasmidType = "GOLDEN_BRAID"
)

type PlasmidType string

func (PlasmidType) GetGraphQLType() string { return "PlasmidType" }

type PlasmidListFilter struct {
	PlasmidType PlasmidType `json:"plasmid_type"`
}

func (PlasmidListFilter) GetGraphQLType() string { return "PlasmidListFilter" }

type Plasmid struct {
	ID      graphql.ID `graphql:"id"`
	Name    string     `graphql:"name"`
	Summary string     `graphql:"summary"`
	InStock bool       `graphql:"in_stock"`
}

type ListPlasmidsQuery struct {
	ListPlasmids struct {
		NextCursor int       `graphql:"nextCursor"`
		TotalCount int       `graphql:"totalCount"`
		Plasmids   []Plasmid `graphql:"plasmids"`
	} `graphql:"listPlasmids(cursor: $cursor, limit: $limit, filter: $filter)"`
}

func listPlasmidsAction(ctx context.Context, cmd *cli.Command) error {
	query := new(ListPlasmidsQuery)
	err := graphql.NewClient(cmd.String("endpoint"), nil).
		Query(ctx, query, map[string]any{
			"cursor": 0,
			"limit":  cmd.Int("limit"),
			"filter": PlasmidListFilter{
				PlasmidType: PlasmidTypeAll,
			},
		})
	if err != nil {
		return fmt.Errorf("graphql query failed: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIN STOCK\tSUMMARY")
	fmt.Fprintln(w, "--\t----\t--------\t-------")
	for _, p := range query.ListPlasmids.Plasmids {
		fmt.Fprintf(w, "%s\t%s\t%t\t%s\n", p.ID, p.Name, p.InStock, p.Summary)
	}
	w.Flush()

	fmt.Fprintf(
		os.Stdout,
		"\nTotal: %d | Next cursor: %d\n",
		query.ListPlasmids.TotalCount,
		query.ListPlasmids.NextCursor,
	)
	return nil
}

func main() {
	cmd := &cli.Command{
		Name:  "gql-plasmid",
		Usage: "CLI to query plasmid data from a GraphQL endpoint",
		Commands: []*cli.Command{
			{
				Name:   "list-plasmids",
				Usage:  "List plasmids from the GraphQL endpoint",
				Action: listPlasmidsAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "endpoint",
						Value: "https://graphql.dictybase.dev/graphql",
						Usage: "GraphQL API endpoint URL",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: defaultLimit,
						Usage: "Number of plasmid entries to fetch",
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
