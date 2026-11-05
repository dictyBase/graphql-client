package main

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	S "github.com/IBM/fp-go/v2/string"
	T "github.com/IBM/fp-go/v2/tuple"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"

	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	predstrings "github.com/dictyBase/fp-go-loom/predicate/strings"
)

// GraphQL variable keys
const (
	gqlVarCursor = "cursor"
	gqlVarFilter = "filter"
	gqlVarInput  = "input"
)

// minStockQueryLen mirrors the server-side minimum query length of the
// searchStocks operation.
const minStockQueryLen = 2

var (
	nonEmptyStringOpt   = O.FromPredicate(S.IsNonEmpty)
	hasMinStockQueryLen = predord.MinStrLen(minStockQueryLen)
	errStockQueryConst  = F.Constant1[string](errStockQueryTooShort)
	checkStockQueryLen  = E.FromPredicate(
		hasMinStockQueryLen,
		errStockQueryConst,
	)
	isNonEmptyOrderItems = predarrays.IsNonEmpty[string]()
	errNoOrderItemsConst = F.Constant1[[]string](errNoOrderItems)
	checkOrderItemsList  = E.FromPredicate(
		isNonEmptyOrderItems,
		errNoOrderItemsConst,
	)
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

func wrapGraphQLQueryErr(err error) error {
	return fmt.Errorf("graphql query failed: %w", err)
}

func wrapGraphQLMutationErr(err error) error {
	return fmt.Errorf("graphql mutation failed: %w", err)
}

func attachPlasmidType(
	input listPlasmidCLIArgs,
	plasmidType PlasmidType,
) listPlasmidValidatedArgs {
	return T.MakeTuple3(input.F1, input.F2, plasmidType)
}

func toListPlasmidValidatedArgs(
	input listPlasmidCLIArgs,
) E.Either[error, listPlasmidValidatedArgs] {
	return F.Pipe2(
		input.F2.String(flagType),
		ParsePlasmidType,
		E.Map[error](F.Bind1st(attachPlasmidType, input)),
	)
}

func mapListPlasmidFetchArgs(
	input listPlasmidValidatedArgs,
) listPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String(flagEndpoint), nil),
		input.F2.Int(flagLimit),
		F.Pipe1(PlasmidListFilter{}, plasmidFilterTypeLens.Set(input.F3)),
	)
}

func fetchListPlasmidsIO(
	input listPlasmidFetchArgs,
) IOE.IOEither[error, ListPlasmidsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListPlasmidsQuery, error) {
			query := new(ListPlasmidsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				gqlVarCursor: 0,
				flagLimit:    input.F3,
				gqlVarFilter: input.F4,
			})
		}),
		IOE.MapLeft[*ListPlasmidsQuery](wrapGraphQLQueryErr),
		IOE.Map[error]((*ListPlasmidsQuery).toResult),
	)
}

func renderPlasmidListResult(result ListPlasmidsResult) error {
	writePlasmidTable(os.Stdout, result.Plasmids)
	writeSummary(os.Stdout, result)
	return nil
}

func RunListPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listPlasmidCLIArgs],
		IOE.ChainEitherK(toListPlasmidValidatedArgs),
		IOE.Map[error](mapListPlasmidFetchArgs),
		IOE.Chain(fetchListPlasmidsIO),
		ioeutils.ToEither[error, ListPlasmidsResult],
		E.Fold(
			F.Identity[error],
			renderPlasmidListResult,
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
		nonEmptyStringOpt,
		O.ToNillable2[string],
	)
}

func buildPlasmidAttributeFilter(
	plasmidType PlasmidType,
	cmd *cli.Command,
) PlasmidAttributeFilter {
	return F.Pipe3(
		PlasmidAttributeFilter{},
		plasmidAttrTypeLens.Set(plasmidType),
		plasmidAttrNameLens.Set(stringPtr(cmd.String(flagName))),
		plasmidAttrSummaryLens.Set(stringPtr(cmd.String(flagSummary))),
	)
}

func toFilteredPlasmidValidatedArgs(
	input filteredPlasmidCLIArgs,
) E.Either[error, filteredPlasmidValidatedArgs] {
	return F.Pipe2(
		input.F2.String(flagType),
		ParsePlasmidType,
		E.Map[error](F.Bind1st(attachPlasmidType, input)),
	)
}

func mapFilteredPlasmidFetchArgs(
	input filteredPlasmidValidatedArgs,
) filteredPlasmidFetchArgs {
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
		IOE.MapLeft[*ListFilteredPlasmidsQuery](wrapGraphQLQueryErr),
		IOE.Map[error]((*ListFilteredPlasmidsQuery).toResult),
	)
}

func RunListFilteredPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, filteredPlasmidCLIArgs],
		IOE.ChainEitherK(toFilteredPlasmidValidatedArgs),
		IOE.Map[error](mapFilteredPlasmidFetchArgs),
		IOE.Chain(fetchListFilteredPlasmidsIO),
		ioeutils.ToEither[error, ListPlasmidsResult],
		E.Fold(
			F.Identity[error],
			renderPlasmidListResult,
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

func attachStrainType(
	input listStrainCLIArgs,
	strainType StrainType,
) listStrainValidatedArgs {
	return T.MakeTuple3(input.F1, input.F2, strainType)
}

func toListStrainValidatedArgs(
	input listStrainCLIArgs,
) E.Either[error, listStrainValidatedArgs] {
	return F.Pipe2(
		input.F2.String(flagType),
		ParseStrainType,
		E.Map[error](F.Bind1st(attachStrainType, input)),
	)
}

func mapListStrainFetchArgs(
	input listStrainValidatedArgs,
) listStrainFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String(flagEndpoint), nil),
		input.F2.Int(flagLimit),
		F.Pipe1(StrainListFilter{}, strainFilterTypeLens.Set(input.F3)),
	)
}

func fetchListStrainsIO(
	input listStrainFetchArgs,
) IOE.IOEither[error, ListStrainsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListStrainsQuery, error) {
			query := new(ListStrainsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				gqlVarCursor: 0,
				flagLimit:    input.F3,
				gqlVarFilter: input.F4,
			})
		}),
		IOE.MapLeft[*ListStrainsQuery](wrapGraphQLQueryErr),
		IOE.Map[error]((*ListStrainsQuery).toResult),
	)
}

func renderStrainListResult(result ListStrainsResult) error {
	writeStrainTable(os.Stdout, result.Strains)
	writeStrainSummary(os.Stdout, result)
	return nil
}

func RunListStrainCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listStrainCLIArgs],
		IOE.ChainEitherK(toListStrainValidatedArgs),
		IOE.Map[error](mapListStrainFetchArgs),
		IOE.Chain(fetchListStrainsIO),
		ioeutils.ToEither[error, ListStrainsResult],
		E.Fold(
			F.Identity[error],
			renderStrainListResult,
		),
	)
}

func buildStrainAttributeFilter(
	strainType StrainType,
	cmd *cli.Command,
) StrainAttributeFilter {
	return F.Pipe3(
		StrainAttributeFilter{},
		strainAttrTypeLens.Set(strainType),
		strainAttrLabelLens.Set(stringPtr(cmd.String(flagLabel))),
		strainAttrSummaryLens.Set(stringPtr(cmd.String(flagSummary))),
	)
}

func toFilteredStrainValidatedArgs(
	input filteredStrainCLIArgs,
) E.Either[error, filteredStrainValidatedArgs] {
	return F.Pipe2(
		input.F2.String(flagType),
		ParseStrainType,
		E.Map[error](F.Bind1st(attachStrainType, input)),
	)
}

func mapFilteredStrainFetchArgs(
	input filteredStrainValidatedArgs,
) filteredStrainFetchArgs {
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
		IOE.MapLeft[*ListFilteredStrainsQuery](wrapGraphQLQueryErr),
		IOE.Map[error]((*ListFilteredStrainsQuery).toResult),
	)
}

func RunListFilteredStrainCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, filteredStrainCLIArgs],
		IOE.ChainEitherK(toFilteredStrainValidatedArgs),
		IOE.Map[error](mapFilteredStrainFetchArgs),
		IOE.Chain(fetchListFilteredStrainsIO),
		ioeutils.ToEither[error, ListStrainsResult],
		E.Fold(
			F.Identity[error],
			renderStrainListResult,
		),
	)
}

// searchStockCLIArgs is the single state type of the search-stocks
// pipeline; it carries the request context and the CLI command through
// every step.
type searchStockCLIArgs struct {
	Context context.Context
	Command *cli.Command
}

// searchStockFetchArgs carries the resolved query inputs for the
// searchStocks GraphQL call.
type searchStockFetchArgs struct {
	Ctx    context.Context
	Client *graphql.Client
	Query  string
	Limit  int
	Entity StockEntityType
}

// validateStockQuery checks that the --query flag meets the server's
// minimum query length and passes the state through unchanged.
func validateStockQuery(
	args searchStockCLIArgs,
) E.Either[error, searchStockCLIArgs] {
	cmd := searchCLICmdLens.Get(args)
	return F.Pipe2(
		cmd.String(flagQuery),
		checkStockQueryLen,
		E.Map[error](F.Constant1[string](args)),
	)
}

// parseStockEntity validates the --entity flag of the search-stocks
// command and passes the state through unchanged.
func parseStockEntity(
	args searchStockCLIArgs,
) E.Either[error, searchStockCLIArgs] {
	cmd := searchCLICmdLens.Get(args)
	return F.Pipe2(
		cmd.String(flagEntity),
		ParseStockEntityType,
		E.Map[error](F.Constant1[StockEntityType](args)),
	)
}

// toSearchStockFetchArgs builds the GraphQL call inputs from the command
// flags of the search-stocks command.
func toSearchStockFetchArgs(args searchStockCLIArgs) searchStockFetchArgs {
	cmd := searchCLICmdLens.Get(args)
	endpoint := cmd.String(flagEndpoint)
	client := graphql.NewClient(endpoint, nil)
	entity := StockEntityType(cmd.String(flagEntity))
	return F.Pipe5(
		searchStockFetchArgs{},
		searchFetchCtxLens.Set(searchCLICtxLens.Get(args)),
		searchFetchClientLens.Set(client),
		searchFetchQueryLens.Set(cmd.String(flagQuery)),
		searchFetchLimitLens.Set(cmd.Int(flagLimit)),
		searchFetchEntityLens.Set(entity),
	)
}

// fetchSearchStocksIO sends the searchStocks query through a client
// built from the command's endpoint flag.
func fetchSearchStocksIO(
	input searchStockFetchArgs,
) IOE.IOEither[error, SearchStocksResult] {
	client := searchFetchClientLens.Get(input)
	ctx := searchFetchCtxLens.Get(input)
	queryStr := searchFetchQueryLens.Get(input)
	limit := searchFetchLimitLens.Get(input)
	entity := searchFetchEntityLens.Get(input)
	return F.Pipe2(
		IOE.TryCatchError(func() (*SearchStocksQuery, error) {
			query := new(SearchStocksQuery)
			return query, client.Query(ctx, query, map[string]any{
				flagQuery:  queryStr,
				flagLimit:  limit,
				flagEntity: entity,
			})
		}),
		IOE.MapLeft[*SearchStocksQuery](wrapGraphQLQueryErr),
		IOE.Map[error]((*SearchStocksQuery).toResult),
	)
}

func ParseStockEntityType(s string) E.Either[error, StockEntityType] {
	switch et := StockEntityType(s); et {
	case StockEntityAll, StockEntityStrain, StockEntityPlasmid:
		return E.Right[error](et)
	default:
		return E.Left[StockEntityType](errInvalidStockEntityType(et))
	}
}

func renderStockSearchResult(result SearchStocksResult) error {
	writeStockSearchTable(os.Stdout, result.Results)
	writeStockSearchSummary(os.Stdout, result)
	return nil
}

func RunSearchStockCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe9(
		searchStockCLIArgs{},
		searchCLICtxLens.Set(ctx),
		searchCLICmdLens.Set(cmd),
		IOE.Of[error, searchStockCLIArgs],
		IOE.ChainEitherK(validateStockQuery),
		IOE.ChainEitherK(parseStockEntity),
		IOE.Map[error](toSearchStockFetchArgs),
		IOE.Chain(fetchSearchStocksIO),
		ioeutils.ToEither[error, SearchStocksResult],
		E.Fold(
			F.Identity[error],
			renderStockSearchResult,
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
		F.Constant1[string](errInvalidEmail(field)),
	)
}

// validateOrderEmails checks the consumer, payer, and purchaser email flags
// of the create-order command and passes the state through unchanged.
func validateOrderEmails(
	args createOrderCLIArgs,
) E.Either[error, createOrderCLIArgs] {
	cmd := orderCLICmdLens.Get(args)
	return F.Pipe2(
		T.MakeTuple3(
			cmd.String(flagConsumer),
			cmd.String(flagPayer),
			cmd.String(flagPurchaser),
		),
		E.TraverseTuple3(
			validateEmail("consumer"),
			validateEmail("payer"),
			validateEmail("purchaser"),
		),
		E.Map[error](F.Constant1[T.Tuple3[string, string, string]](args)),
	)
}

func validateOrderStatus(
	args createOrderCLIArgs,
) E.Either[error, createOrderCLIArgs] {
	cmd := orderCLICmdLens.Get(args)
	return F.Pipe2(
		cmd.String(flagStatus),
		ParseStatus,
		E.Map[error](F.Constant1[StatusEnum](args)),
	)
}

func validateOrderItems(
	args createOrderCLIArgs,
) E.Either[error, createOrderCLIArgs] {
	cmd := orderCLICmdLens.Get(args)
	return F.Pipe2(
		cmd.StringSlice(flagItems),
		checkOrderItemsList,
		E.Map[error](F.Constant1[[]string](args)),
	)
}

func toCreateOrderInput(args createOrderCLIArgs) CreateOrderInfo {
	cmd := orderCLICmdLens.Get(args)
	input := F.Pipe12(
		CreateOrderInput{},
		orderCourierLens.Set(cmd.String(flagCourier)),
		orderCourierAcctLens.Set(cmd.String(flagCourierAccount)),
		orderCommentsLens.Set(cmd.String(flagComments)),
		orderPaymentLens.Set(cmd.String(flagPayment)),
		orderPONumLens.Set(cmd.String(flagPONum)),
		orderStatusLens.Set(StatusEnum(cmd.String(flagStatus))),
		orderConsumerLens.Set(cmd.String(flagConsumer)),
		orderPayerLens.Set(cmd.String(flagPayer)),
		orderPurchaserLens.Set(cmd.String(flagPurchaser)),
		orderItemsLens.Set(cmd.StringSlice(flagItems)),
		orderConsumerInfoLens.Set(fakeConsumerInfo()),
		orderPayerInfoLens.Set(fakePayerInfo()),
	)
	return F.Pipe3(
		CreateOrderInfo{},
		orderInfoEndpointLens.Set(cmd.String(flagEndpoint)),
		orderInfoCtxLens.Set(orderCLICtxLens.Get(args)),
		orderInfoInputLens.Set(input),
	)
}

// mutateCreateOrder sends the createOrder mutation with the given input
// through a client built from the command's endpoint flag.
func mutateCreateOrder(info CreateOrderInfo) IOE.IOEither[error, string] {
	endpoint := orderInfoEndpointLens.Get(info)
	ctx := orderInfoCtxLens.Get(info)
	input := orderInfoInputLens.Get(info)
	return F.Pipe2(
		IOE.TryCatchError(func() (*CreateOrderMutation, error) {
			mutation := new(CreateOrderMutation)
			client := graphql.NewClient(endpoint, nil)
			return mutation, client.Mutate(
				ctx,
				mutation,
				map[string]any{gqlVarInput: input},
			)
		}),
		IOE.MapLeft[*CreateOrderMutation](wrapGraphQLMutationErr),
		IOE.Map[error]((*CreateOrderMutation).toOrderID),
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

func renderCreatedOrderResult(orderID string) error {
	writeCreatedOrder(os.Stdout, orderID)
	return nil
}

func RunCreateOrderCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe9(
		createOrderCLIArgs{},
		orderCLICtxLens.Set(ctx),
		orderCLICmdLens.Set(cmd),
		IOE.Of[error, createOrderCLIArgs],
		IOE.ChainEitherK(validateOrderEmails),
		IOE.ChainEitherK(validateOrderStatus),
		IOE.ChainEitherK(validateOrderItems),
		IOE.Chain(submitCreateOrder),
		ioeutils.ToEither[error, string],
		E.Fold(
			F.Identity[error],
			renderCreatedOrderResult,
		),
	)
}
