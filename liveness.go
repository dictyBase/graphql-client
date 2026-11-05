package main

import (
	"context"
	"encoding/json"
	"errors"
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
}

type livenessProbeInput struct {
	Ctx     context.Context
	Timeout time.Duration
	Client  *graphql.Client
	Probe   LivenessProbe
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
	endpointErr := eithererr.ConstErr[string]("endpoint URL is required")
	timeoutErr := eithererr.ConstErr[time.Duration](
		"timeout must be greater than zero",
	)
	positiveDuration := N.MoreThan(zeroDurationValue)
	checkTimeout := func(string) E.Either[error, time.Duration] {
		return F.Pipe1(
			args.Command.Duration(flagTimeout),
			E.FromPredicate(positiveDuration, timeoutErr),
		)
	}
	keepArgs := F.Constant1[time.Duration](args)
	return F.Pipe3(
		args.Command.String(flagEndpoint),
		E.FromPredicate(S.IsNonEmpty, endpointErr),
		E.Chain(checkTimeout),
		E.Map[error](keepArgs),
	)
}

func toLivenessRunState(args livenessCLIArgs) livenessRunState {
	endpoint := args.Command.String(flagEndpoint)
	return livenessRunState{
		Ctx:       args.Context,
		Endpoint:  endpoint,
		Timeout:   args.Command.Duration(flagTimeout),
		Client:    graphql.NewClient(endpoint, nil),
		Probes:    defaultLivenessProbes(),
		Specs:     defaultLivenessOperationSpecs(),
		StartedAt: time.Now(),
	}
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
	emptyErr := eithererr.ConstErr[[]byte]("empty graphql response payload")
	lookupRoot := R.Lookup[json.RawMessage](rootField)
	extractRoot := F.Flow3(
		lookupRoot,
		O.Map(rawJSONToString),
		O.Chain(O.FromPredicate(isNonNullJSON)),
	)
	nullRootErr := func() error {
		return errNullProbeRoot(rootField)
	}
	return F.Flow4(
		E.FromPredicate(predbytes.HasPositiveLen, emptyErr),
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

func extractFirstGraphQLError(errs graphql.Errors) O.Option[string] {
	slice := []graphql.Error(errs)
	return F.Pipe2(
		slice,
		A.Head[graphql.Error],
		O.Map(func(item graphql.Error) string { return item.Message }),
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
	validator := validateProbePayload(input.Probe.RootField)
	return F.Pipe1(
		IOE.TryCatchError(func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(input.Ctx, input.Timeout)
			defer cancel()
			opOpt := graphql.OperationName(input.Probe.Operation)
			raw, err := input.Client.ExecRaw(
				ctx,
				input.Probe.Query,
				input.Probe.Variables,
				opOpt,
			)
			return raw, err
		}),
		IOE.ChainEitherK(validator),
	)
}

func executeProbe(
	input livenessProbeInput,
) IOE.IOEither[error, LivenessProbeOutcome] {
	return IOE.Defer(func() IOE.IOEither[error, LivenessProbeOutcome] {
		started := time.Now()
		onFail := func(err error) IO.IO[LivenessProbeOutcome] {
			msg := conciseGraphQLError(err)
			return IO.Of(LivenessProbeOutcome{
				Probe:    input.Probe,
				Passed:   false,
				Latency:  time.Since(started),
				ErrorMsg: msg,
			})
		}
		onPass := func(string) IO.IO[LivenessProbeOutcome] {
			return IO.Of(LivenessProbeOutcome{
				Probe:   input.Probe,
				Passed:  true,
				Latency: time.Since(started),
			})
		}
		rawEffect := rawExecProbe(input)
		return F.Pipe2(
			rawEffect,
			IOE.Fold(onFail, onPass),
			IOE.FromIO[error, LivenessProbeOutcome],
		)
	})
}

func isPassedOutcome(outcome LivenessProbeOutcome) bool {
	return outcome.Passed
}

func outcomeOperation(outcome LivenessProbeOutcome) string {
	return outcome.Probe.Operation
}

func sumProbeLatency(
	acc time.Duration,
	item LivenessProbeOutcome,
) time.Duration {
	return acc + item.Latency
}

func averageProbeLatency(outcomes []LivenessProbeOutcome) time.Duration {
	nonEmpty := predarrays.IsNonEmpty[LivenessProbeOutcome]()
	computeMean := func(items []LivenessProbeOutcome) time.Duration {
		total := F.Pipe1(items, A.Reduce(sumProbeLatency, 0))
		count := time.Duration(len(items))
		return total / count
	}
	zeroDuration := F.Constant(zeroDurationValue)
	return F.Pipe3(
		outcomes,
		O.FromPredicate(nonEmpty),
		O.Map(computeMean),
		O.GetOrElse(zeroDuration),
	)
}

func operationStatus(failed int) string {
	armPass := MO.Const(isZeroCount, probeStatusPass)
	return MO.First(probeStatusFail, []O.Option[string]{
		armPass(failed),
	})
}

func overallLivenessStatus(passed, failed int) string {
	armHealthy := MO.Const(isZeroCount, livenessStatusHealthy)
	armDown := MO.Const(isZeroCount, livenessStatusDown)
	return MO.First(livenessStatusDegraded, []O.Option[string]{
		armHealthy(failed),
		armDown(passed),
	})
}

func summarizeOperation(
	outcomes []LivenessProbeOutcome,
) func(LivenessOperationSpec) OperationHealth {
	return func(spec LivenessOperationSpec) OperationHealth {
		matchOp := F.Pipe1(
			predord.StrEq(spec.Operation),
			P.ContraMap(outcomeOperation),
		)
		opOutcomes := F.Pipe1(outcomes, A.Filter(matchOp))
		passed := F.Pipe2(
			opOutcomes,
			A.Filter(isPassedOutcome),
			A.Size[LivenessProbeOutcome],
		)
		total := A.Size(opOutcomes)
		failed := total - passed
		return OperationHealth{
			Operation: spec.Operation,
			RouteArea: spec.RouteArea,
			Total:     total,
			Passed:    passed,
			Failed:    failed,
			Latency:   averageProbeLatency(opOutcomes),
			Status:    operationStatus(failed),
		}
	}
}

func buildLivenessReport(
	state livenessRunState,
	outcomes []LivenessProbeOutcome,
) LivenessReport {
	isFailed := P.Not(isPassedOutcome)
	passed := F.Pipe2(
		outcomes,
		A.Filter(isPassedOutcome),
		A.Size[LivenessProbeOutcome],
	)
	total := A.Size(outcomes)
	failed := total - passed
	summarize := summarizeOperation(outcomes)
	return LivenessReport{
		Endpoint:   state.Endpoint,
		Status:     overallLivenessStatus(passed, failed),
		Total:      total,
		Passed:     passed,
		Failed:     failed,
		Duration:   time.Since(state.StartedAt),
		Operations: F.Pipe1(state.Specs, A.Map(summarize)),
		Failures:   F.Pipe1(outcomes, A.Filter(isFailed)),
	}
}

func evaluateLivenessState(
	state livenessRunState,
) IOE.IOEither[error, LivenessReport] {
	toInput := func(probe LivenessProbe) livenessProbeInput {
		return livenessProbeInput{
			Ctx:     state.Ctx,
			Timeout: state.Timeout,
			Client:  state.Client,
			Probe:   probe,
		}
	}
	inputs := F.Pipe1(state.Probes, A.Map(toInput))
	traverseProbes := IOE.TraverseArray(executeProbe)
	toReport := func(outcomes []LivenessProbeOutcome) LivenessReport {
		return buildLivenessReport(state, outcomes)
	}
	return F.Pipe2(
		inputs,
		traverseProbes,
		IOE.Map[error](toReport),
	)
}

func RunLivenessCheckCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		livenessCLIArgs{Context: ctx, Command: cmd},
		IOE.Of[error, livenessCLIArgs],
		IOE.ChainEitherK(validateLivenessArgs),
		IOE.Map[error](toLivenessRunState),
		IOE.Chain(evaluateLivenessState),
		ioeutils.ToEither[error, LivenessReport],
		E.Fold(
			F.Identity[error],
			func(report LivenessReport) error {
				return writeLivenessMarkdown(os.Stdout, report)
			},
		),
	)
}
