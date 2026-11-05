package main

import (
	"context"
	"time"

	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/nao1215/markdown"
	"github.com/urfave/cli/v3"
)

var (
	// Lenses on LivenessOperationSpec.
	specOperationLens = L.MakeLens(
		func(s LivenessOperationSpec) string { return s.Operation },
		func(s LivenessOperationSpec, v string) LivenessOperationSpec {
			s.Operation = v
			return s
		},
	)
	specRouteAreaLens = L.MakeLens(
		func(s LivenessOperationSpec) string { return s.RouteArea },
		func(s LivenessOperationSpec, v string) LivenessOperationSpec {
			s.RouteArea = v
			return s
		},
	)

	// Lenses on LivenessProbe.
	probeOperationLens = L.MakeLens(
		func(p LivenessProbe) string { return p.Operation },
		func(p LivenessProbe, v string) LivenessProbe {
			p.Operation = v
			return p
		},
	)
	probeRouteAreaLens = L.MakeLens(
		func(p LivenessProbe) string { return p.RouteArea },
		func(p LivenessProbe, v string) LivenessProbe {
			p.RouteArea = v
			return p
		},
	)
	probeRootFieldLens = L.MakeLens(
		func(p LivenessProbe) string { return p.RootField },
		func(p LivenessProbe, v string) LivenessProbe {
			p.RootField = v
			return p
		},
	)
	probeLabelLens = L.MakeLens(
		func(p LivenessProbe) string { return p.Label },
		func(p LivenessProbe, v string) LivenessProbe {
			p.Label = v
			return p
		},
	)
	probeQueryLens = L.MakeLens(
		func(p LivenessProbe) string { return p.Query },
		func(p LivenessProbe, v string) LivenessProbe {
			p.Query = v
			return p
		},
	)
	probeVariablesLens = L.MakeLens(
		func(p LivenessProbe) map[string]any { return p.Variables },
		func(p LivenessProbe, v map[string]any) LivenessProbe {
			p.Variables = v
			return p
		},
	)

	// Lenses on LivenessProbeOutcome and composed lenses into LivenessProbe.
	outcomeProbeLens = L.MakeLens(
		func(o LivenessProbeOutcome) LivenessProbe { return o.Probe },
		func(o LivenessProbeOutcome, v LivenessProbe) LivenessProbeOutcome {
			o.Probe = v
			return o
		},
	)
	outcomePassedLens = L.MakeLens(
		func(o LivenessProbeOutcome) bool { return o.Passed },
		func(o LivenessProbeOutcome, v bool) LivenessProbeOutcome {
			o.Passed = v
			return o
		},
	)
	outcomeLatencyLens = L.MakeLens(
		func(o LivenessProbeOutcome) time.Duration { return o.Latency },
		func(o LivenessProbeOutcome, v time.Duration) LivenessProbeOutcome {
			o.Latency = v
			return o
		},
	)
	outcomeErrorMsgLens = L.MakeLens(
		func(o LivenessProbeOutcome) string { return o.ErrorMsg },
		func(o LivenessProbeOutcome, v string) LivenessProbeOutcome {
			o.ErrorMsg = v
			return o
		},
	)
	outcomeOperationLens = F.Pipe1(
		outcomeProbeLens,
		L.Compose[LivenessProbeOutcome](probeOperationLens),
	)
	outcomeLabelLens = F.Pipe1(
		outcomeProbeLens,
		L.Compose[LivenessProbeOutcome](probeLabelLens),
	)

	// Lenses on OperationHealth.
	opHealthOperationLens = L.MakeLens(
		func(h OperationHealth) string { return h.Operation },
		func(h OperationHealth, v string) OperationHealth {
			h.Operation = v
			return h
		},
	)
	opHealthRouteAreaLens = L.MakeLens(
		func(h OperationHealth) string { return h.RouteArea },
		func(h OperationHealth, v string) OperationHealth {
			h.RouteArea = v
			return h
		},
	)
	opHealthTotalLens = L.MakeLens(
		func(h OperationHealth) int { return h.Total },
		func(h OperationHealth, v int) OperationHealth {
			h.Total = v
			return h
		},
	)
	opHealthPassedLens = L.MakeLens(
		func(h OperationHealth) int { return h.Passed },
		func(h OperationHealth, v int) OperationHealth {
			h.Passed = v
			return h
		},
	)
	opHealthFailedLens = L.MakeLens(
		func(h OperationHealth) int { return h.Failed },
		func(h OperationHealth, v int) OperationHealth {
			h.Failed = v
			return h
		},
	)
	opHealthLatencyLens = L.MakeLens(
		func(h OperationHealth) time.Duration { return h.Latency },
		func(h OperationHealth, v time.Duration) OperationHealth {
			h.Latency = v
			return h
		},
	)
	opHealthStatusLens = L.MakeLens(
		func(h OperationHealth) string { return h.Status },
		func(h OperationHealth, v string) OperationHealth {
			h.Status = v
			return h
		},
	)

	// Lenses on LivenessReport.
	reportEndpointLens = L.MakeLens(
		func(r LivenessReport) string { return r.Endpoint },
		func(r LivenessReport, v string) LivenessReport {
			r.Endpoint = v
			return r
		},
	)
	reportStatusLens = L.MakeLens(
		func(r LivenessReport) string { return r.Status },
		func(r LivenessReport, v string) LivenessReport {
			r.Status = v
			return r
		},
	)
	reportTotalLens = L.MakeLens(
		func(r LivenessReport) int { return r.Total },
		func(r LivenessReport, v int) LivenessReport {
			r.Total = v
			return r
		},
	)
	reportPassedLens = L.MakeLens(
		func(r LivenessReport) int { return r.Passed },
		func(r LivenessReport, v int) LivenessReport {
			r.Passed = v
			return r
		},
	)
	reportFailedLens = L.MakeLens(
		func(r LivenessReport) int { return r.Failed },
		func(r LivenessReport, v int) LivenessReport {
			r.Failed = v
			return r
		},
	)
	reportDurationLens = L.MakeLens(
		func(r LivenessReport) time.Duration { return r.Duration },
		func(r LivenessReport, v time.Duration) LivenessReport {
			r.Duration = v
			return r
		},
	)
	reportOperationsLens = L.MakeLens(
		func(r LivenessReport) []OperationHealth { return r.Operations },
		func(r LivenessReport, v []OperationHealth) LivenessReport {
			r.Operations = v
			return r
		},
	)
	reportFailuresLens = L.MakeLens(
		func(r LivenessReport) []LivenessProbeOutcome { return r.Failures },
		func(r LivenessReport, v []LivenessProbeOutcome) LivenessReport {
			r.Failures = v
			return r
		},
	)

	// Lenses on livenessCLIArgs.
	cliArgsCtxLens = L.MakeLens(
		func(a livenessCLIArgs) context.Context { return a.Context },
		func(a livenessCLIArgs, v context.Context) livenessCLIArgs {
			a.Context = v
			return a
		},
	)
	cliArgsCmdLens = L.MakeLens(
		func(a livenessCLIArgs) *cli.Command { return a.Command },
		func(a livenessCLIArgs, v *cli.Command) livenessCLIArgs {
			a.Command = v
			return a
		},
	)

	// Lenses on livenessRunState.
	stateCtxLens = L.MakeLens(
		func(s livenessRunState) context.Context { return s.Ctx },
		func(s livenessRunState, v context.Context) livenessRunState {
			s.Ctx = v
			return s
		},
	)
	stateEndpointLens = L.MakeLens(
		func(s livenessRunState) string { return s.Endpoint },
		func(s livenessRunState, v string) livenessRunState {
			s.Endpoint = v
			return s
		},
	)
	stateTimeoutLens = L.MakeLens(
		func(s livenessRunState) time.Duration { return s.Timeout },
		func(s livenessRunState, v time.Duration) livenessRunState {
			s.Timeout = v
			return s
		},
	)
	stateClientLens = L.MakeLens(
		func(s livenessRunState) *graphql.Client { return s.Client },
		func(s livenessRunState, v *graphql.Client) livenessRunState {
			s.Client = v
			return s
		},
	)
	stateProbesLens = L.MakeLens(
		func(s livenessRunState) []LivenessProbe { return s.Probes },
		func(s livenessRunState, v []LivenessProbe) livenessRunState {
			s.Probes = v
			return s
		},
	)
	stateSpecsLens = L.MakeLens(
		func(s livenessRunState) []LivenessOperationSpec { return s.Specs },
		func(s livenessRunState, v []LivenessOperationSpec) livenessRunState {
			s.Specs = v
			return s
		},
	)
	stateStartedAtLens = L.MakeLens(
		func(s livenessRunState) time.Time { return s.StartedAt },
		func(s livenessRunState, v time.Time) livenessRunState {
			s.StartedAt = v
			return s
		},
	)

	// Lenses on livenessProbeInput and composed lenses into LivenessProbe.
	probeInputCtxLens = L.MakeLens(
		func(i livenessProbeInput) context.Context { return i.Ctx },
		func(i livenessProbeInput, v context.Context) livenessProbeInput {
			i.Ctx = v
			return i
		},
	)
	probeInputTimeoutLens = L.MakeLens(
		func(i livenessProbeInput) time.Duration { return i.Timeout },
		func(i livenessProbeInput, v time.Duration) livenessProbeInput {
			i.Timeout = v
			return i
		},
	)
	probeInputClientLens = L.MakeLens(
		func(i livenessProbeInput) *graphql.Client { return i.Client },
		func(i livenessProbeInput, v *graphql.Client) livenessProbeInput {
			i.Client = v
			return i
		},
	)
	probeInputProbeLens = L.MakeLens(
		func(i livenessProbeInput) LivenessProbe { return i.Probe },
		func(i livenessProbeInput, v LivenessProbe) livenessProbeInput {
			i.Probe = v
			return i
		},
	)
	probeInputOperationLens = F.Pipe1(
		probeInputProbeLens,
		L.Compose[livenessProbeInput](probeOperationLens),
	)
	probeInputRootFieldLens = F.Pipe1(
		probeInputProbeLens,
		L.Compose[livenessProbeInput](probeRootFieldLens),
	)
	probeInputQueryLens = F.Pipe1(
		probeInputProbeLens,
		L.Compose[livenessProbeInput](probeQueryLens),
	)
	probeInputVariablesLens = F.Pipe1(
		probeInputProbeLens,
		L.Compose[livenessProbeInput](probeVariablesLens),
	)

	// Lenses on markdown.TableSet.
	tableHeaderLens = L.MakeLens(
		func(t markdown.TableSet) []string { return t.Header },
		func(t markdown.TableSet, v []string) markdown.TableSet {
			t.Header = v
			return t
		},
	)
	tableRowsLens = L.MakeLens(
		func(t markdown.TableSet) [][]string { return t.Rows },
		func(t markdown.TableSet, v [][]string) markdown.TableSet {
			t.Rows = v
			return t
		},
	)
	tableAlignmentLens = L.MakeLens(
		func(t markdown.TableSet) []markdown.TableAlignment {
			return t.Alignment
		},
		func(
			t markdown.TableSet,
			v []markdown.TableAlignment,
		) markdown.TableSet {
			t.Alignment = v
			return t
		},
	)
)
