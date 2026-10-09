package main

import (
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"
)

// stringPtr returns a nil pointer for an empty string, otherwise a pointer
// to the value.
func stringPtr(s string) *string {
	return F.Pipe2(
		s,
		O.FromPredicate(P.IsNonZero[string]()),
		O.Fold(
			func() *string { return nil },
			func(s string) *string { return &s },
		),
	)
}

// clientFromCommand builds the GraphQL client for the command's endpoint.
func clientFromCommand(cmd *cli.Command) *graphql.Client {
	endpoint := cmd.String(flagEndpoint)

	return graphql.NewClient(endpoint, nil)
}
