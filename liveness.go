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
	J "github.com/IBM/fp-go/v2/json"
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
	extractNonNullRawJSON = F.Flow2(
		O.Map(rawJSONToString),
		O.Chain(O.FromPredicate(isNonNullJSON)),
	)
	isNonEmptyOutcomes  = predarrays.IsNonEmpty[LivenessProbeOutcome]()
	isFailedOutcome     = P.Not(outcomePassedLens.Get)
	countPassedOutcomes = F.Flow2(
		A.Filter(outcomePassedLens.Get),
		A.Size[LivenessProbeOutcome],
	)
	filterFailedOutcomes = A.Filter(isFailedOutcome)
	zeroDurationThunk    = F.Constant(zeroDurationValue)
	traverseAllProbes    = IOE.TraverseArray(executeProbe)
	armPassStatus        = MO.Const(isZeroCount, probeStatusPass)
	armHealthyStatus     = MO.Const(isZeroCount, livenessStatusHealthy)
	armDownStatus        = MO.Const(isZeroCount, livenessStatusDown)
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
	Report    LivenessReport
}

type opHealthBuildState struct {
	spec     LivenessOperationSpec
	outcomes []LivenessProbeOutcome
	health   OperationHealth
}

type livenessProbeInput struct {
	Ctx       context.Context
	Timeout   time.Duration
	Client    *graphql.Client
	Probe     LivenessProbe
	StartedAt time.Time
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
	return F.Pipe7(
		livenessRunState{},
		stateCtxLens.Set(cliArgsCtxLens.Get(args)),
		stateEndpointLens.Set(cliEndpoint(args)),
		stateTimeoutLens.Set(cliTimeout(args)),
		stateClientLens.Set(graphql.NewClient(cliEndpoint(args), nil)),
		stateProbesLens.Set(defaultLivenessProbes()),
		stateSpecsLens.Set(defaultLivenessOperationSpecs()),
		stateStartedAtLens.Set(time.Now()),
	)
}

func rawJSONToString(raw json.RawMessage) string {
	return string(raw)
}

func validateProbePayload(
	rootField string,
	raw []byte,
) E.Either[error, string] {
	return F.Pipe5(
		raw,
		checkNonEmptyPayload,
		E.Chain(J.Unmarshal[map[string]json.RawMessage]),
		E.Map[error](R.Lookup[json.RawMessage](rootField)),
		E.Map[error](extractNonNullRawJSON),
		E.Chain(E.FromOption[string](F.Constant(errNullProbeRoot(rootField)))),
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
	return F.Pipe2(
		[]graphql.Error(errs),
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
	rootField := probeInputRootFieldLens.Get(input)
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
		IOE.ChainEitherK(F.Bind1st(validateProbePayload, rootField)),
	)
}

func stampProbeStart(input livenessProbeInput) IO.IO[livenessProbeInput] {
	return func() livenessProbeInput {
		return F.Pipe1(input, probeInputStartedAtLens.Set(time.Now()))
	}
}

func buildFailedOutcome(
	input livenessProbeInput,
	err error,
) IO.IO[LivenessProbeOutcome] {
	return F.Pipe5(
		LivenessProbeOutcome{},
		outcomeProbeLens.Set(probeInputProbeLens.Get(input)),
		outcomePassedLens.Set(false),
		outcomeLatencyLens.Set(time.Since(probeInputStartedAtLens.Get(input))),
		outcomeErrorMsgLens.Set(conciseGraphQLError(err)),
		IO.Of[LivenessProbeOutcome],
	)
}

func buildPassedOutcome(
	input livenessProbeInput,
	_ string,
) IO.IO[LivenessProbeOutcome] {
	return F.Pipe4(
		LivenessProbeOutcome{},
		outcomeProbeLens.Set(probeInputProbeLens.Get(input)),
		outcomePassedLens.Set(true),
		outcomeLatencyLens.Set(time.Since(probeInputStartedAtLens.Get(input))),
		IO.Of[LivenessProbeOutcome],
	)
}

func runTimedProbe(
	input livenessProbeInput,
) IOE.IOEither[error, LivenessProbeOutcome] {
	return F.Pipe3(
		input,
		rawExecProbe,
		IOE.Fold(
			F.Bind1st(buildFailedOutcome, input),
			F.Bind1st(buildPassedOutcome, input),
		),
		IOE.FromIO[error, LivenessProbeOutcome],
	)
}

func executeProbe(
	input livenessProbeInput,
) IOE.IOEither[error, LivenessProbeOutcome] {
	return F.Pipe3(
		input,
		stampProbeStart,
		IOE.FromIO[error, livenessProbeInput],
		IOE.Chain(runTimedProbe),
	)
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

func matchOperationName(opName string) P.Predicate[LivenessProbeOutcome] {
	return F.Pipe1(
		predord.StrEq(opName),
		P.ContraMap(outcomeOperationLens.Get),
	)
}

func setOpBuildOperation(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(s, opBuildOperationLens.Set(opBuildSpecOpLens.Get(s)))
}

func setOpBuildRouteArea(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(s, opBuildRouteAreaLens.Set(opBuildSpecRouteLens.Get(s)))
}

func filterOpBuildOutcomes(s opHealthBuildState) opHealthBuildState {
	return F.Pipe2(
		opBuildOutcomesLens.Get(s),
		A.Filter(matchOperationName(opBuildOperationLens.Get(s))),
		F.Pipe1(s, F.Flip(opBuildOutcomesLens.Set)),
	)
}

func setOpBuildTotal(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(s, opBuildTotalLens.Set(A.Size(opBuildOutcomesLens.Get(s))))
}

func setOpBuildPassed(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(
		s,
		opBuildPassedLens.Set(countPassedOutcomes(opBuildOutcomesLens.Get(s))),
	)
}

func setOpBuildFailed(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(
		s,
		opBuildFailedLens.Set(
			opBuildTotalLens.Get(s)-opBuildPassedLens.Get(s),
		),
	)
}

func setOpBuildLatency(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(
		s,
		opBuildLatencyLens.Set(
			averageProbeLatency(opBuildOutcomesLens.Get(s)),
		),
	)
}

func setOpBuildStatus(s opHealthBuildState) opHealthBuildState {
	return F.Pipe1(
		s,
		opBuildStatusLens.Set(operationStatus(opBuildFailedLens.Get(s))),
	)
}

func summarizeOperation(
	outcomes []LivenessProbeOutcome,
	spec LivenessOperationSpec,
) OperationHealth {
	return F.Pipe11(
		opHealthBuildState{},
		opBuildSpecLens.Set(spec),
		opBuildOutcomesLens.Set(outcomes),
		setOpBuildOperation,
		setOpBuildRouteArea,
		filterOpBuildOutcomes,
		setOpBuildTotal,
		setOpBuildPassed,
		setOpBuildFailed,
		setOpBuildLatency,
		setOpBuildStatus,
		opBuildHealthLens.Get,
	)
}

func setReportEndpoint(s livenessRunState) livenessRunState {
	return F.Pipe1(s, stateReportEndpointLens.Set(stateEndpointLens.Get(s)))
}

func setReportTotal(s livenessRunState) livenessRunState {
	return F.Pipe1(
		s,
		stateReportTotalLens.Set(A.Size(stateOutcomesLens.Get(s))),
	)
}

func setReportPassed(s livenessRunState) livenessRunState {
	return F.Pipe1(
		s,
		stateReportPassedLens.Set(countPassedOutcomes(stateOutcomesLens.Get(s))),
	)
}

func setReportFailed(s livenessRunState) livenessRunState {
	return F.Pipe1(
		s,
		stateReportFailedLens.Set(
			stateReportTotalLens.Get(s)-stateReportPassedLens.Get(s),
		),
	)
}

func setReportStatus(s livenessRunState) livenessRunState {
	return F.Pipe1(
		s,
		stateReportStatusLens.Set(
			overallLivenessStatus(
				stateReportPassedLens.Get(s),
				stateReportFailedLens.Get(s),
			),
		),
	)
}

func setReportDuration(s livenessRunState) livenessRunState {
	return F.Pipe1(
		s,
		stateReportDurationLens.Set(time.Since(stateStartedAtLens.Get(s))),
	)
}

func setReportOperations(s livenessRunState) livenessRunState {
	return F.Pipe2(
		stateSpecsLens.Get(s),
		A.Map(F.Bind1st(summarizeOperation, stateOutcomesLens.Get(s))),
		F.Pipe1(s, F.Flip(stateReportOperationsLens.Set)),
	)
}

func setReportFailures(s livenessRunState) livenessRunState {
	return F.Pipe2(
		stateOutcomesLens.Get(s),
		filterFailedOutcomes,
		F.Pipe1(s, F.Flip(stateReportFailuresLens.Set)),
	)
}

func buildLivenessReport(state livenessRunState) LivenessReport {
	return F.Pipe9(
		state,
		setReportEndpoint,
		setReportTotal,
		setReportPassed,
		setReportFailed,
		setReportStatus,
		setReportDuration,
		setReportOperations,
		setReportFailures,
		stateReportLens.Get,
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
