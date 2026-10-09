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
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predstrings "github.com/dictyBase/fp-go-loom/predicate/strings"
)

// GraphQL variable keys
const (
	gqlVarCursor = "cursor"
	gqlVarFilter = "filter"
	gqlVarInput  = "input"
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
		return E.Right[error](pt)
	default:
		return E.Left[PlasmidType](errInvalidPlasmidType(pt))
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

// createOrderCLIArgs is the single state type of the create-order pipeline;
// it carries the request context and the CLI command through every step.
type createOrderCLIArgs struct {
	Context context.Context
	Command *cli.Command
}

func ParseStatus(s string) E.Either[error, StatusEnum] {
	switch st := StatusEnum(s); st {
	case StatusInPreparation, StatusGrowing, StatusCancelled, StatusShipped:
		return E.Right[error](st)
	default:
		return E.Left[StatusEnum](errInvalidStatus(s))
	}
}

// validateEmail checks one email address and tags the error with its role.
func validateEmail(field string) E.Kleisli[error, string, string] {
	return E.FromPredicate(
		predstrings.HasAtSign,
		func(string) error { return errInvalidEmail(field) },
	)
}

// validateOrderEmails checks the consumer, payer, and purchaser email flags
// of the create-order command and passes the state through unchanged.
func validateOrderEmails(args createOrderCLIArgs) E.Either[error, createOrderCLIArgs] {
	return F.Pipe2(
		T.MakeTuple3(
			args.Command.String(flagConsumer),
			args.Command.String(flagPayer),
			args.Command.String(flagPurchaser),
		),
		E.TraverseTuple3(
			validateEmail("consumer"),
			validateEmail("payer"),
			validateEmail("purchaser"),
		),
		E.Map[error](func(T.Tuple3[string, string, string]) createOrderCLIArgs {
			return args
		}),
	)
}

func validateOrderStatus(args createOrderCLIArgs) E.Either[error, createOrderCLIArgs] {
	return F.Pipe2(
		args.Command.String(flagStatus),
		ParseStatus,
		E.Map[error](func(StatusEnum) createOrderCLIArgs { return args }),
	)
}

func validateOrderItems(args createOrderCLIArgs) E.Either[error, createOrderCLIArgs] {
	return F.Pipe2(
		args.Command.StringSlice(flagItems),
		E.FromPredicate(
			predarrays.IsNonEmpty[string](),
			func([]string) error { return errNoOrderItems },
		),
		E.Map[error](func([]string) createOrderCLIArgs { return args }),
	)
}

func toCreateOrderInput(args createOrderCLIArgs) CreateOrderInfo {
	cmd := args.Command
	input := CreateOrderInput{
		Courier:          cmd.String(flagCourier),
		CourierAccount:   cmd.String(flagCourierAccount),
		Comments:         cmd.String(flagComments),
		Payment:          cmd.String(flagPayment),
		PurchaseOrderNum: cmd.String(flagPONum),
		Status:           StatusEnum(cmd.String(flagStatus)),
		Consumer:         cmd.String(flagConsumer),
		Payer:            cmd.String(flagPayer),
		Purchaser:        cmd.String(flagPurchaser),
		Items:            cmd.StringSlice(flagItems),
		ConsumerInfo:     fakeConsumerInfo(),
		PayerInfo:        fakePayerInfo(),
	}
	return CreateOrderInfo{
		Endpoint: cmd.String(flagEndpoint),
		Ctx:      args.Context,
		Input:    input,
	}
}

// mutateCreateOrder sends the createOrder mutation with the given input
// through a client built from the command's endpoint flag.
func mutateCreateOrder(info CreateOrderInfo) IOE.IOEither[error, string] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*CreateOrderMutation, error) {
			mutation := new(CreateOrderMutation)
			vars := map[string]any{gqlVarInput: info.Input}
			client := graphql.NewClient(info.Endpoint, nil)
			return mutation, client.Mutate(info.Ctx, mutation, vars)
		}),
		IOE.MapLeft[*CreateOrderMutation](func(err error) error {
			return fmt.Errorf("graphql mutation failed: %w", err)
		}),
		IOE.Map[error](func(mutation *CreateOrderMutation) string {
			return string(mutation.CreateOrder.ID)
		}),
	)
}

// submitCreateOrder builds the order input from the command flags and runs
// the createOrder mutation, yielding the new order id.
func submitCreateOrder(args createOrderCLIArgs) IOE.IOEither[error, string] {
	return F.Pipe3(
		args,
		IOE.Of[error, createOrderCLIArgs],
		IOE.Map[error](toCreateOrderInput),
		IOE.Chain(mutateCreateOrder),
	)
}

func RunCreateOrderCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe7(
		createOrderCLIArgs{Context: ctx, Command: cmd},
		IOE.Of[error, createOrderCLIArgs],
		IOE.ChainEitherK(validateOrderEmails),
		IOE.ChainEitherK(validateOrderStatus),
		IOE.ChainEitherK(validateOrderItems),
		IOE.Chain(submitCreateOrder),
		ioeutils.ToEither[error, string],
		E.Fold(
			func(err error) error { return err },
			func(orderID string) error {
				writeCreatedOrder(os.Stdout, orderID)
				return nil
			},
		),
	)
}
