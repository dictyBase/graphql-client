package main

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	T "github.com/IBM/fp-go/v2/tuple"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"

	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
)

// GraphQL variable keys
const (
	gqlVarCursor = "cursor"
	gqlVarFilter = "filter"
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

type (
	listStrainCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	listStrainValidatedArgs = T.Tuple3[context.Context, *cli.Command, StrainType]
	listStrainFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, StrainListFilter]
)

type (
	filteredStrainCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	filteredStrainValidatedArgs = T.Tuple3[context.Context, *cli.Command, StrainType]
	filteredStrainFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, StrainAttributeFilter]
)

func toListPlasmidValidatedArgs(
	input listPlasmidCLIArgs,
) IOE.IOEither[error, listPlasmidValidatedArgs] {
	return F.Pipe2(
		ParsePlasmidType(input.F2.String(flagType)),
		IOE.FromEither[error, PlasmidType],
		IOE.Map[error](func(plasmidType PlasmidType) listPlasmidValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, plasmidType)
		}),
	)
}

func mapListPlasmidFetchArgs(input listPlasmidValidatedArgs) listPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String(flagEndpoint), nil),
		input.F2.Int(flagLimit),
		PlasmidListFilter{PlasmidType: input.F3},
	)
}

func fetchListPlasmidsIO(input listPlasmidFetchArgs) IOE.IOEither[error, ListPlasmidsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListPlasmidsQuery, error) {
			query := new(ListPlasmidsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				gqlVarCursor: 0,
				flagLimit:    input.F3,
				gqlVarFilter: input.F4,
			})
		}),
		IOE.MapLeft[*ListPlasmidsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListPlasmidsQuery).toResult),
	)
}

func RunListPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listPlasmidCLIArgs],
		IOE.Chain(toListPlasmidValidatedArgs),
		IOE.Map[error](mapListPlasmidFetchArgs),
		IOE.Chain(fetchListPlasmidsIO),
		ioeutils.ToEither[error, ListPlasmidsResult],
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
	switch pt := PlasmidType(s); pt {
	case PlasmidTypeAll, PlasmidTypeRegular, PlasmidTypeGoldenBraid:
		return E.Right[error, PlasmidType](pt)
	default:
		return E.Left[PlasmidType, error](errInvalidPlasmidType(pt))
	}
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
		Name:        stringPtr(cmd.String(flagName)),
		Summary:     stringPtr(cmd.String(flagSummary)),
	}
}

func toFilteredPlasmidValidatedArgs(
	input filteredPlasmidCLIArgs,
) IOE.IOEither[error, filteredPlasmidValidatedArgs] {
	return F.Pipe2(
		ParsePlasmidType(input.F2.String(flagType)),
		IOE.FromEither[error, PlasmidType],
		IOE.Map[error](func(plasmidType PlasmidType) filteredPlasmidValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, plasmidType)
		}),
	)
}

func mapFilteredPlasmidFetchArgs(input filteredPlasmidValidatedArgs) filteredPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String(flagEndpoint), nil),
		input.F2.Int(flagLimit),
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
				gqlVarCursor: 0,
				flagLimit:    input.F3,
				gqlVarFilter: input.F4,
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
		ioeutils.ToEither[error, ListPlasmidsResult],
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

func ParseStrainType(s string) E.Either[error, StrainType] {
	switch st := StrainType(s); st {
	case StrainTypeAll, StrainTypeRegular, StrainTypeGwdi, StrainTypeBacterial:
		return E.Right[error](st)
	default:
		return E.Left[StrainType](errInvalidStrainType(st))
	}
}

func toListStrainValidatedArgs(
	input listStrainCLIArgs,
) IOE.IOEither[error, listStrainValidatedArgs] {
	return F.Pipe2(
		ParseStrainType(input.F2.String(flagType)),
		IOE.FromEither[error, StrainType],
		IOE.Map[error](func(strainType StrainType) listStrainValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, strainType)
		}),
	)
}

func mapListStrainFetchArgs(input listStrainValidatedArgs) listStrainFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String(flagEndpoint), nil),
		input.F2.Int(flagLimit),
		StrainListFilter{StrainType: input.F3},
	)
}

func fetchListStrainsIO(input listStrainFetchArgs) IOE.IOEither[error, ListStrainsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListStrainsQuery, error) {
			query := new(ListStrainsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				gqlVarCursor: 0,
				flagLimit:    input.F3,
				gqlVarFilter: input.F4,
			})
		}),
		IOE.MapLeft[*ListStrainsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListStrainsQuery).toResult),
	)
}

func RunListStrainCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listStrainCLIArgs],
		IOE.Chain(toListStrainValidatedArgs),
		IOE.Map[error](mapListStrainFetchArgs),
		IOE.Chain(fetchListStrainsIO),
		ioeutils.ToEither[error, ListStrainsResult],
		E.Fold(
			func(err error) error { return err },
			func(result ListStrainsResult) error {
				writeStrainTable(os.Stdout, result.Strains)
				writeStrainSummary(os.Stdout, result)
				return nil
			},
		),
	)
}

func buildStrainAttributeFilter(
	strainType StrainType,
	cmd *cli.Command,
) StrainAttributeFilter {
	return StrainAttributeFilter{
		StrainType: strainType,
		Label:      stringPtr(cmd.String(flagLabel)),
		Summary:    stringPtr(cmd.String(flagSummary)),
	}
}

func toFilteredStrainValidatedArgs(
	input filteredStrainCLIArgs,
) IOE.IOEither[error, filteredStrainValidatedArgs] {
	return F.Pipe2(
		ParseStrainType(input.F2.String(flagType)),
		IOE.FromEither[error, StrainType],
		IOE.Map[error](func(strainType StrainType) filteredStrainValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, strainType)
		}),
	)
}

func mapFilteredStrainFetchArgs(input filteredStrainValidatedArgs) filteredStrainFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String(flagEndpoint), nil),
		input.F2.Int(flagLimit),
		buildStrainAttributeFilter(input.F3, input.F2),
	)
}

func fetchListFilteredStrainsIO(
	input filteredStrainFetchArgs,
) IOE.IOEither[error, ListStrainsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListFilteredStrainsQuery, error) {
			query := new(ListFilteredStrainsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				gqlVarCursor: 0,
				flagLimit:    input.F3,
				gqlVarFilter: input.F4,
			})
		}),
		IOE.MapLeft[*ListFilteredStrainsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListFilteredStrainsQuery).toResult),
	)
}

func RunListFilteredStrainCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, filteredStrainCLIArgs],
		IOE.Chain(toFilteredStrainValidatedArgs),
		IOE.Map[error](mapFilteredStrainFetchArgs),
		IOE.Chain(fetchListFilteredStrainsIO),
		ioeutils.ToEither[error, ListStrainsResult],
		E.Fold(
			func(err error) error { return err },
			func(result ListStrainsResult) error {
				writeStrainTable(os.Stdout, result.Strains)
				writeStrainSummary(os.Stdout, result)
				return nil
			},
		),
	)
}
