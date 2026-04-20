package main

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	R "github.com/IBM/fp-go/v2/record"
	T "github.com/IBM/fp-go/v2/tuple"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"
)

type (
	listPlasmidCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	listPlasmidValidatedArgs = T.Tuple3[context.Context, *cli.Command, PlasmidType]
	listPlasmidFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, PlasmidListFilter]
	listPlasmidResultTuple   = T.Tuple2[ListPlasmidsResult, error]
)

func toListPlasmidValidatedArgs(
	input listPlasmidCLIArgs,
) IOE.IOEither[error, listPlasmidValidatedArgs] {
	return F.Pipe4(
		plasmidTypeMap,
		R.Lookup[PlasmidType](input.F2.String("type")),
		E.FromOption[PlasmidType](func() error {
			return errInvalidPlasmidType(PlasmidType(input.F2.String("type")))
		}),
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
	return IOE.TryCatchError(func() (ListPlasmidsResult, error) {
		return fetchPlasmids(
			input.F1,
			input.F2,
			0,
			input.F3,
			input.F4,
		)
	})
}

func toListPlasmidResultErrorTuple(err error) listPlasmidResultTuple {
	return T.MakeTuple2(ListPlasmidsResult{}, err)
}

func toListPlasmidResultSuccessTuple(result ListPlasmidsResult) listPlasmidResultTuple {
	return T.MakeTuple2(result, error(nil))
}

func displayListPlasmidResultTuple(result listPlasmidResultTuple) error {
	displayResults(result.F1)
	return result.F2
}

func foldListPlasmidCLIResult(result E.Either[error, ListPlasmidsResult]) error {
	return F.Pipe1(
		result,
		E.Fold(
			F.Flow2(
				toListPlasmidResultErrorTuple,
				T.Second[ListPlasmidsResult, error],
			),
			F.Flow2(
				toListPlasmidResultSuccessTuple,
				displayListPlasmidResultTuple,
			),
		),
	)
}

func evaluateListPlasmidCLIResult(result IOE.IOEither[error, ListPlasmidsResult]) error {
	return F.Pipe1(
		result(),
		foldListPlasmidCLIResult,
	)
}

func RunListPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listPlasmidCLIArgs],
		IOE.Chain(toListPlasmidValidatedArgs),
		IOE.Map[error](mapListPlasmidFetchArgs),
		IOE.Chain(fetchListPlasmidsIO),
		evaluateListPlasmidCLIResult,
	)
}
