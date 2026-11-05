package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	EQ "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	N "github.com/IBM/fp-go/v2/number"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"

	eithererr "github.com/dictyBase/fp-go-loom/eithererr"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	MO "github.com/dictyBase/fp-go-loom/matchopt"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predbytes "github.com/dictyBase/fp-go-loom/predicate/bytes"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
)

const zeroDurationValue time.Duration = 0

var (
	isZeroCount   = F.Pipe1(0, EQ.Equals(predord.IntEq))
	isNotNullJSON = predord.NotEqualStr("null")
	isNonNullJSON = F.Pipe1(
		S.IsNonEmpty,
		P.And(isNotNullJSON),
	)
	hasPositiveDuration      = N.MoreThan(zeroDurationValue)
	hasValidLivenessEndpoint = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(cliEndpoint),
	)
	hasValidLivenessTimeout = F.Pipe1(
		hasPositiveDuration,
		P.ContraMap(cliTimeout),
	)
	errEmptyEndpoint = eithererr.ConstErr[livenessCLIArgs](
		"endpoint URL is required",
	)
	errInvalidTimeout = eithererr.ConstErr[livenessCLIArgs](
		"timeout must be greater than zero",
	)
	checkLivenessEndpoint = E.FromPredicate(
		hasValidLivenessEndpoint,
		errEmptyEndpoint,
	)
	checkLivenessTimeout = E.FromPredicate(
		hasValidLivenessTimeout,
		errInvalidTimeout,
	)
	errEmptyPayload = eithererr.ConstErr[[]byte](
		"empty graphql response payload",
	)
	checkNonEmptyPayload = E.FromPredicate(
		predbytes.HasPositiveLen,
		errEmptyPayload,
	)
	filterNonNullJSON  = O.FromPredicate(isNonNullJSON)
	mapRawJSONToString = O.Map(rawJSONToString)
	isNonEmptyOutcomes = predarrays.IsNonEmpty[LivenessProbeOutcome]()
	isFailedOutcome    = P.Not(outcomePassedLens.Get)
	zeroDurationThunk  = F.Constant(zeroDurationValue)
	traverseAllProbes  = IOE.TraverseArray(executeProbe)
	armPassStatus      = MO.Const(isZeroCount, probeStatusPass)
	armHealthyStatus   = MO.Const(isZeroCount, livenessStatusHealthy)
	armDownStatus      = MO.Const(isZeroCount, livenessStatusDown)
)

type livenessCLIArgs struct {
	Context context.Context
	Command *cli.Command
}

type livenessRunState struct {
	Ctx       context.Context
	Endpoint  string
	Timeout   time.Duration
	Client    *graphql.Client
	Probes    []LivenessProbe
	Specs     []LivenessOperationSpec
	StartedAt time.Time
	Outcomes  []LivenessProbeOutcome
}

type livenessProbeInput struct {
	Ctx     context.Context
	Timeout time.Duration
	Client  *graphql.Client
	Probe   LivenessProbe
}

func cliEndpoint(args livenessCLIArgs) string {
	return cliArgsCmdLens.Get(args).String(flagEndpoint)
}

func cliTimeout(args livenessCLIArgs) time.Duration {
	return cliArgsCmdLens.Get(args).Duration(flagTimeout)
}

func defaultLivenessProbes() []LivenessProbe {
	return A.ArrayConcatAll(
		catalogLivenessProbes(),
		searchAndPublicationLivenessProbes(),
		contentAndDetailLivenessProbes(),
		phenotypeLivenessProbes(),
	)
}

func validateLivenessArgs(
	args livenessCLIArgs,
) E.Either[error, livenessCLIArgs] {
	return F.Pipe2(
		args,
		checkLivenessEndpoint,
		E.Chain(checkLivenessTimeout),
	)
}

func toLivenessRunState(args livenessCLIArgs) livenessRunState {
	endpoint := cliEndpoint(args)
	return F.Pipe7(
		livenessRunState{},
		stateCtxLens.Set(cliArgsCtxLens.Get(args)),
		stateEndpointLens.Set(endpoint),
		stateTimeoutLens.Set(cliTimeout(args)),
		stateClientLens.Set(graphql.NewClient(endpoint, nil)),
		stateProbesLens.Set(defaultLivenessProbes()),
		stateSpecsLens.Set(defaultLivenessOperationSpecs()),
		stateStartedAtLens.Set(time.Now()),
	)
}

func rawJSONToString(raw json.RawMessage) string {
	return string(raw)
}

func unmarshalProbeObject(
	raw []byte,
) E.Either[error, map[string]json.RawMessage] {
	var fields map[string]json.RawMessage
	err := json.Unmarshal(raw, &fields)
	return E.TryCatchError(fields, err)
}

func validateProbePayload(
	rootField string,
) E.Kleisli[error, []byte, string] {
	extractRoot := F.Flow3(
		R.Lookup[json.RawMessage](rootField),
		mapRawJSONToString,
		O.Chain(filterNonNullJSON),
	)
	nullRootErr := F.Constant(errNullProbeRoot(rootField))
	return F.Flow4(
		checkNonEmptyPayload,
		E.Chain(unmarshalProbeObject),
		E.Map[error](extractRoot),
		E.Chain(E.FromOption[string](nullRootErr)),
	)
}

func asGraphQLErrors(err error) (graphql.Errors, bool) {
	var gqlErrs graphql.Errors
	ok := errors.As(err, &gqlErrs)
	return gqlErrs, ok
}

func graphQLErrorMessage(item graphql.Error) string {
	return item.Message
}

func extractFirstGraphQLError(errs graphql.Errors) O.Option[string] {
	slice := []graphql.Error(errs)
	return F.Pipe2(
		slice,
		A.Head[graphql.Error],
		O.Map(graphQLErrorMessage),
	)
}

func conciseGraphQLError(err error) string {
	return F.Pipe3(
		err,
		O.FromValidation(asGraphQLErrors),
		O.Chain(extractFirstGraphQLError),
		O.GetOrElse(err.Error),
	)
}

func rawExecProbe(input livenessProbeInput) IOE.IOEither[error, string] {
	validator := validateProbePayload(probeInputRootFieldLens.Get(input))
	baseCtx := probeInputCtxLens.Get(input)
	timeout := probeInputTimeoutLens.Get(input)
	client := probeInputClientLens.Get(input)
	operation := probeInputOperationLens.Get(input)
	queryDoc := probeInputQueryLens.Get(input)
	variables := probeInputVariablesLens.Get(input)
	return F.Pipe1(
		IOE.TryCatchError(func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(baseCtx, timeout)
			defer cancel()
			opOpt := graphql.OperationName(operation)
			raw, err := client.ExecRaw(
				ctx,
				queryDoc,
				variables,
				opOpt,
			)
			return raw, err
		}),
		IOE.ChainEitherK(validator),
	)
}

func failedOutcomeIO(
	probe LivenessProbe,
	started time.Time,
) func(error) IO.IO[LivenessProbeOutcome] {
	return func(err error) IO.IO[LivenessProbeOutcome] {
		msg := conciseGraphQLError(err)
		outcome := F.Pipe4(
			LivenessProbeOutcome{},
			outcomeProbeLens.Set(probe),
			outcomePassedLens.Set(false),
			outcomeLatencyLens.Set(time.Since(started)),
			outcomeErrorMsgLens.Set(msg),
		)
		return IO.Of(outcome)
	}
}

func passedOutcomeIO(
	probe LivenessProbe,
	started time.Time,
) func(string) IO.IO[LivenessProbeOutcome] {
	return func(string) IO.IO[LivenessProbeOutcome] {
		outcome := F.Pipe3(
			LivenessProbeOutcome{},
			outcomeProbeLens.Set(probe),
			outcomePassedLens.Set(true),
			outcomeLatencyLens.Set(time.Since(started)),
		)
		return IO.Of(outcome)
	}
}

func executeProbe(
	input livenessProbeInput,
) IOE.IOEither[error, LivenessProbeOutcome] {
	probe := probeInputProbeLens.Get(input)
	return IOE.Defer(func() IOE.IOEither[error, LivenessProbeOutcome] {
		started := time.Now()
		return F.Pipe3(
			input,
			rawExecProbe,
			IOE.Fold(
				failedOutcomeIO(probe, started),
				passedOutcomeIO(probe, started),
			),
			IOE.FromIO[error, LivenessProbeOutcome],
		)
	})
}

func sumProbeLatency(
	acc time.Duration,
	item LivenessProbeOutcome,
) time.Duration {
	return acc + outcomeLatencyLens.Get(item)
}

func meanProbeLatency(items []LivenessProbeOutcome) time.Duration {
	total := F.Pipe1(items, A.Reduce(sumProbeLatency, 0))
	count := time.Duration(len(items))
	return total / count
}

func averageProbeLatency(outcomes []LivenessProbeOutcome) time.Duration {
	return F.Pipe3(
		outcomes,
		O.FromPredicate(isNonEmptyOutcomes),
		O.Map(meanProbeLatency),
		O.GetOrElse(zeroDurationThunk),
	)
}

func operationStatus(failed int) string {
	return MO.First(probeStatusFail, []O.Option[string]{
		armPassStatus(failed),
	})
}

func overallLivenessStatus(passed, failed int) string {
	return MO.First(livenessStatusDegraded, []O.Option[string]{
		armHealthyStatus(failed),
		armDownStatus(passed),
	})
}

func summarizeOperation(
	outcomes []LivenessProbeOutcome,
	spec LivenessOperationSpec,
) OperationHealth {
	opName := specOperationLens.Get(spec)
	matchOp := F.Pipe1(
		predord.StrEq(opName),
		P.ContraMap(outcomeOperationLens.Get),
	)
	opOutcomes := F.Pipe1(outcomes, A.Filter(matchOp))
	passed := F.Pipe2(
		opOutcomes,
		A.Filter(outcomePassedLens.Get),
		A.Size[LivenessProbeOutcome],
	)
	total := A.Size(opOutcomes)
	failed := total - passed
	return F.Pipe7(
		OperationHealth{},
		opHealthOperationLens.Set(opName),
		opHealthRouteAreaLens.Set(specRouteAreaLens.Get(spec)),
		opHealthTotalLens.Set(total),
		opHealthPassedLens.Set(passed),
		opHealthFailedLens.Set(failed),
		opHealthLatencyLens.Set(averageProbeLatency(opOutcomes)),
		opHealthStatusLens.Set(operationStatus(failed)),
	)
}

func buildLivenessReport(state livenessRunState) LivenessReport {
	outcomes := stateOutcomesLens.Get(state)
	passed := F.Pipe2(
		outcomes,
		A.Filter(outcomePassedLens.Get),
		A.Size[LivenessProbeOutcome],
	)
	total := A.Size(outcomes)
	failed := total - passed
	ops := F.Pipe1(
		stateSpecsLens.Get(state),
		A.Map(F.Bind1st(summarizeOperation, outcomes)),
	)
	failures := F.Pipe1(outcomes, A.Filter(isFailedOutcome))
	return F.Pipe8(
		LivenessReport{},
		reportEndpointLens.Set(stateEndpointLens.Get(state)),
		reportStatusLens.Set(overallLivenessStatus(passed, failed)),
		reportTotalLens.Set(total),
		reportPassedLens.Set(passed),
		reportFailedLens.Set(failed),
		reportDurationLens.Set(time.Since(stateStartedAtLens.Get(state))),
		reportOperationsLens.Set(ops),
		reportFailuresLens.Set(failures),
	)
}

func newProbeInput(
	state livenessRunState,
	probe LivenessProbe,
) livenessProbeInput {
	return F.Pipe4(
		livenessProbeInput{},
		probeInputCtxLens.Set(stateCtxLens.Get(state)),
		probeInputTimeoutLens.Set(stateTimeoutLens.Get(state)),
		probeInputClientLens.Set(stateClientLens.Get(state)),
		probeInputProbeLens.Set(probe),
	)
}

func runAllProbes(
	state livenessRunState,
) IOE.IOEither[error, []LivenessProbeOutcome] {
	return F.Pipe3(
		state,
		stateProbesLens.Get,
		A.Map(F.Bind1st(newProbeInput, state)),
		traverseAllProbes,
	)
}

func evaluateLivenessState(
	state livenessRunState,
) IOE.IOEither[error, LivenessReport] {
	return F.Pipe2(
		IOE.Of[error](state),
		IOE.Bind(stateOutcomesLens.Set, runAllProbes),
		IOE.Map[error](buildLivenessReport),
	)
}

func RunLivenessCheckCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe10(
		livenessCLIArgs{},
		cliArgsCtxLens.Set(ctx),
		cliArgsCmdLens.Set(cmd),
		IOE.Of[error, livenessCLIArgs],
		IOE.ChainEitherK(checkLivenessEndpoint),
		IOE.ChainEitherK(checkLivenessTimeout),
		IOE.Map[error](toLivenessRunState),
		IOE.Bind(stateOutcomesLens.Set, runAllProbes),
		IOE.Map[error](buildLivenessReport),
		ioeutils.ToEither[error, LivenessReport],
		E.Fold(
			F.Identity[error],
			F.Bind1st(writeLivenessMarkdown, io.Writer(os.Stdout)),
		),
	)
}
