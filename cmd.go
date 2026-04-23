package main

import (
	"context"
	"fmt"
	"os"

	R "github.com/IBM/fp-go/v2/record"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	T "github.com/IBM/fp-go/v2/tuple"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"
)

type (
	listPlasmidCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	listPlasmidValidatedArgs = T.Tuple3[context.Context, *cli.Command, PlasmidType]
	listPlasmidFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, PlasmidListFilter]
)

type (
	filteredPlasmidCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	filteredPlasmidValidatedArgs = T.Tuple3[context.Context, *cli.Command, PlasmidType]
	filteredPlasmidFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, PlasmidAttributeFilter]
)

func toListPlasmidValidatedArgs(
	input listPlasmidCLIArgs,
) IOE.IOEither[error, listPlasmidValidatedArgs] {
	return F.Pipe2(
		ParsePlasmidType(input.F2.String("type")),
		IOE.FromEither[error, PlasmidType],
		IOE.Map[error](func(plasmidType PlasmidType) listPlasmidValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, plasmidType)
		}),
	)
}

func mapListPlasmidFetchArgs(input listPlasmidValidatedArgs) listPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String("endpoint"), nil),
		input.F2.Int("limit"),
		PlasmidListFilter{PlasmidType: input.F3},
	)
}

func fetchListPlasmidsIO(input listPlasmidFetchArgs) IOE.IOEither[error, ListPlasmidsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListPlasmidsQuery, error) {
			query := new(ListPlasmidsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				"cursor": 0,
				"limit":  input.F3,
				"filter": input.F4,
			})
		}),
		IOE.MapLeft[*ListPlasmidsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListPlasmidsQuery).toResult),
	)
}

func toEither[ERR, A any](ioe IOE.IOEither[ERR, A]) E.Either[ERR, A] {
	return ioe()
}

func RunListPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listPlasmidCLIArgs],
		IOE.Chain(toListPlasmidValidatedArgs),
		IOE.Map[error](mapListPlasmidFetchArgs),
		IOE.Chain(fetchListPlasmidsIO),
		toEither[error, ListPlasmidsResult],
		E.Fold(
			func(err error) error { return err },
			func(result ListPlasmidsResult) error {
				writePlasmidTable(os.Stdout, result.Plasmids)
				writeSummary(os.Stdout, result)
				return nil
			},
		),
	)
}

func ParsePlasmidType(s string) E.Either[error, PlasmidType] {
	return F.Pipe2(
		plasmidTypeMap,
		R.Lookup[PlasmidType](s),
		E.FromOption[PlasmidType](func() error {
			return errInvalidPlasmidType(PlasmidType(s))
		}),
	)
}

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

func buildPlasmidAttributeFilter(
	plasmidType PlasmidType,
	cmd *cli.Command,
) PlasmidAttributeFilter {
	return PlasmidAttributeFilter{
		PlasmidType: plasmidType,
		Name:        stringPtr(cmd.String("name")),
		Summary:     stringPtr(cmd.String("summary")),
	}
}

func toFilteredPlasmidValidatedArgs(
	input filteredPlasmidCLIArgs,
) IOE.IOEither[error, filteredPlasmidValidatedArgs] {
	return F.Pipe2(
		ParsePlasmidType(input.F2.String("type")),
		IOE.FromEither[error, PlasmidType],
		IOE.Map[error](func(plasmidType PlasmidType) filteredPlasmidValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, plasmidType)
		}),
	)
}

func mapFilteredPlasmidFetchArgs(input filteredPlasmidValidatedArgs) filteredPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String("endpoint"), nil),
		input.F2.Int("limit"),
		buildPlasmidAttributeFilter(input.F3, input.F2),
	)
}

func fetchListFilteredPlasmidsIO(
	input filteredPlasmidFetchArgs,
) IOE.IOEither[error, ListPlasmidsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListFilteredPlasmidsQuery, error) {
			query := new(ListFilteredPlasmidsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				"cursor": 0,
				"limit":  input.F3,
				"filter": input.F4,
			})
		}),
		IOE.MapLeft[*ListFilteredPlasmidsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListFilteredPlasmidsQuery).toResult),
	)
}

func RunListFilteredPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, filteredPlasmidCLIArgs],
		IOE.Chain(toFilteredPlasmidValidatedArgs),
		IOE.Map[error](mapFilteredPlasmidFetchArgs),
		IOE.Chain(fetchListFilteredPlasmidsIO),
		toEither[error, ListPlasmidsResult],
		E.Fold(
			func(err error) error { return err },
			func(result ListPlasmidsResult) error {
				writePlasmidTable(os.Stdout, result.Plasmids)
				writeSummary(os.Stdout, result)
				return nil
			},
		),
	)
}
