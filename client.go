package main

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
)

func newGraphQLClient(endpoint string) *graphql.Client {
	return graphql.NewClient(endpoint, nil)
}

func fetchPlasmids(
	ctx context.Context,
	client *graphql.Client,
	cursor int,
	limit int,
	filter PlasmidListFilter,
) (ListPlasmidsResult, error) {
	query := new(ListPlasmidsQuery)
	err := client.Query(ctx, query, map[string]any{
		"cursor": cursor,
		"limit":  limit,
		"filter": filter,
	})
	if err != nil {
		return ListPlasmidsResult{}, fmt.Errorf("graphql query failed: %w", err)
	}
	return query.toResult(), nil
}
