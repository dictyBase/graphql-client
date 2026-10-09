package main

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	T "github.com/IBM/fp-go/v2/tuple"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"

	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
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

func ParseStrainType(s string) E.Either[error, StrainType] {
	switch st := StrainType(s); st {
	case StrainTypeAll, StrainTypeRegular, StrainTypeGwdi, StrainTypeBacterial:
		return E.Right[error](st)
	default:
		return E.Left[StrainType](errInvalidStrainType(st))
	}
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
		clientFromCommand(input.F2),
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
		clientFromCommand(input.F2),
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
