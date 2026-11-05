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
	stateOutcomesLens = L.MakeLens(
		func(s livenessRunState) []LivenessProbeOutcome { return s.Outcomes },
		func(
			s livenessRunState,
			v []LivenessProbeOutcome,
		) livenessRunState {
			s.Outcomes = v
			return s
		},
	)
	stateReportLens = L.MakeLens(
		func(s livenessRunState) LivenessReport { return s.Report },
		func(s livenessRunState, v LivenessReport) livenessRunState {
			s.Report = v
			return s
		},
	)
	stateReportEndpointLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportEndpointLens),
	)
	stateReportStatusLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportStatusLens),
	)
	stateReportTotalLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportTotalLens),
	)
	stateReportPassedLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportPassedLens),
	)
	stateReportFailedLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportFailedLens),
	)
	stateReportDurationLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportDurationLens),
	)
	stateReportOperationsLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportOperationsLens),
	)
	stateReportFailuresLens = F.Pipe1(
		stateReportLens,
		L.Compose[livenessRunState](reportFailuresLens),
	)

	// Lenses on opHealthBuildState and composed lenses into Spec and Health.
	opBuildSpecLens = L.MakeLens(
		func(s opHealthBuildState) LivenessOperationSpec { return s.spec },
		func(
			s opHealthBuildState,
			v LivenessOperationSpec,
		) opHealthBuildState {
			s.spec = v
			return s
		},
	)
	opBuildOutcomesLens = L.MakeLens(
		func(s opHealthBuildState) []LivenessProbeOutcome {
			return s.outcomes
		},
		func(
			s opHealthBuildState,
			v []LivenessProbeOutcome,
		) opHealthBuildState {
			s.outcomes = v
			return s
		},
	)
	opBuildHealthLens = L.MakeLens(
		func(s opHealthBuildState) OperationHealth { return s.health },
		func(s opHealthBuildState, v OperationHealth) opHealthBuildState {
			s.health = v
			return s
		},
	)
	opBuildSpecOpLens = F.Pipe1(
		opBuildSpecLens,
		L.Compose[opHealthBuildState](specOperationLens),
	)
	opBuildSpecRouteLens = F.Pipe1(
		opBuildSpecLens,
		L.Compose[opHealthBuildState](specRouteAreaLens),
	)
	opBuildOperationLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthOperationLens),
	)
	opBuildRouteAreaLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthRouteAreaLens),
	)
	opBuildTotalLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthTotalLens),
	)
	opBuildPassedLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthPassedLens),
	)
	opBuildFailedLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthFailedLens),
	)
	opBuildLatencyLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthLatencyLens),
	)
	opBuildStatusLens = F.Pipe1(
		opBuildHealthLens,
		L.Compose[opHealthBuildState](opHealthStatusLens),
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
	probeInputStartedAtLens = L.MakeLens(
		func(i livenessProbeInput) time.Time { return i.StartedAt },
		func(i livenessProbeInput, v time.Time) livenessProbeInput {
			i.StartedAt = v
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

	// Lenses on PlasmidListFilter and PlasmidAttributeFilter.
	plasmidFilterTypeLens = L.MakeLens(
		func(f PlasmidListFilter) PlasmidType { return f.PlasmidType },
		func(f PlasmidListFilter, v PlasmidType) PlasmidListFilter {
			f.PlasmidType = v
			return f
		},
	)
	plasmidAttrTypeLens = L.MakeLens(
		func(f PlasmidAttributeFilter) PlasmidType { return f.PlasmidType },
		func(f PlasmidAttributeFilter, v PlasmidType) PlasmidAttributeFilter {
			f.PlasmidType = v
			return f
		},
	)
	plasmidAttrNameLens = L.MakeLens(
		func(f PlasmidAttributeFilter) *string { return f.Name },
		func(f PlasmidAttributeFilter, v *string) PlasmidAttributeFilter {
			f.Name = v
			return f
		},
	)
	plasmidAttrSummaryLens = L.MakeLens(
		func(f PlasmidAttributeFilter) *string { return f.Summary },
		func(f PlasmidAttributeFilter, v *string) PlasmidAttributeFilter {
			f.Summary = v
			return f
		},
	)

	// Lenses on StrainListFilter and StrainAttributeFilter.
	strainFilterTypeLens = L.MakeLens(
		func(f StrainListFilter) StrainType { return f.StrainType },
		func(f StrainListFilter, v StrainType) StrainListFilter {
			f.StrainType = v
			return f
		},
	)
	strainAttrTypeLens = L.MakeLens(
		func(f StrainAttributeFilter) StrainType { return f.StrainType },
		func(f StrainAttributeFilter, v StrainType) StrainAttributeFilter {
			f.StrainType = v
			return f
		},
	)
	strainAttrLabelLens = L.MakeLens(
		func(f StrainAttributeFilter) *string { return f.Label },
		func(f StrainAttributeFilter, v *string) StrainAttributeFilter {
			f.Label = v
			return f
		},
	)
	strainAttrSummaryLens = L.MakeLens(
		func(f StrainAttributeFilter) *string { return f.Summary },
		func(f StrainAttributeFilter, v *string) StrainAttributeFilter {
			f.Summary = v
			return f
		},
	)

	// Lenses on searchStockCLIArgs and searchStockFetchArgs.
	searchCLICtxLens = L.MakeLens(
		func(a searchStockCLIArgs) context.Context { return a.Context },
		func(a searchStockCLIArgs, v context.Context) searchStockCLIArgs {
			a.Context = v
			return a
		},
	)
	searchCLICmdLens = L.MakeLens(
		func(a searchStockCLIArgs) *cli.Command { return a.Command },
		func(a searchStockCLIArgs, v *cli.Command) searchStockCLIArgs {
			a.Command = v
			return a
		},
	)
	searchFetchCtxLens = L.MakeLens(
		func(a searchStockFetchArgs) context.Context { return a.Ctx },
		func(a searchStockFetchArgs, v context.Context) searchStockFetchArgs {
			a.Ctx = v
			return a
		},
	)
	searchFetchClientLens = L.MakeLens(
		func(a searchStockFetchArgs) *graphql.Client { return a.Client },
		func(a searchStockFetchArgs, v *graphql.Client) searchStockFetchArgs {
			a.Client = v
			return a
		},
	)
	searchFetchQueryLens = L.MakeLens(
		func(a searchStockFetchArgs) string { return a.Query },
		func(a searchStockFetchArgs, v string) searchStockFetchArgs {
			a.Query = v
			return a
		},
	)
	searchFetchLimitLens = L.MakeLens(
		func(a searchStockFetchArgs) int { return a.Limit },
		func(a searchStockFetchArgs, v int) searchStockFetchArgs {
			a.Limit = v
			return a
		},
	)
	searchFetchEntityLens = L.MakeLens(
		func(a searchStockFetchArgs) StockEntityType { return a.Entity },
		func(a searchStockFetchArgs, v StockEntityType) searchStockFetchArgs {
			a.Entity = v
			return a
		},
	)

	// Lenses on createOrderCLIArgs, CreateOrderInput, and CreateOrderInfo.
	orderCLICtxLens = L.MakeLens(
		func(a createOrderCLIArgs) context.Context { return a.Context },
		func(a createOrderCLIArgs, v context.Context) createOrderCLIArgs {
			a.Context = v
			return a
		},
	)
	orderCLICmdLens = L.MakeLens(
		func(a createOrderCLIArgs) *cli.Command { return a.Command },
		func(a createOrderCLIArgs, v *cli.Command) createOrderCLIArgs {
			a.Command = v
			return a
		},
	)
	orderCourierLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.Courier },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.Courier = v
			return i
		},
	)
	orderCourierAcctLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.CourierAccount },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.CourierAccount = v
			return i
		},
	)
	orderCommentsLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.Comments },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.Comments = v
			return i
		},
	)
	orderPaymentLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.Payment },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.Payment = v
			return i
		},
	)
	orderPONumLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.PurchaseOrderNum },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.PurchaseOrderNum = v
			return i
		},
	)
	orderStatusLens = L.MakeLens(
		func(i CreateOrderInput) StatusEnum { return i.Status },
		func(i CreateOrderInput, v StatusEnum) CreateOrderInput {
			i.Status = v
			return i
		},
	)
	orderConsumerLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.Consumer },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.Consumer = v
			return i
		},
	)
	orderPayerLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.Payer },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.Payer = v
			return i
		},
	)
	orderPurchaserLens = L.MakeLens(
		func(i CreateOrderInput) string { return i.Purchaser },
		func(i CreateOrderInput, v string) CreateOrderInput {
			i.Purchaser = v
			return i
		},
	)
	orderItemsLens = L.MakeLens(
		func(i CreateOrderInput) []string { return i.Items },
		func(i CreateOrderInput, v []string) CreateOrderInput {
			i.Items = v
			return i
		},
	)
	orderConsumerInfoLens = L.MakeLens(
		func(i CreateOrderInput) *UserInfoInput { return i.ConsumerInfo },
		func(i CreateOrderInput, v *UserInfoInput) CreateOrderInput {
			i.ConsumerInfo = v
			return i
		},
	)
	orderPayerInfoLens = L.MakeLens(
		func(i CreateOrderInput) *UserInfoInput { return i.PayerInfo },
		func(i CreateOrderInput, v *UserInfoInput) CreateOrderInput {
			i.PayerInfo = v
			return i
		},
	)
	orderInfoEndpointLens = L.MakeLens(
		func(i CreateOrderInfo) string { return i.Endpoint },
		func(i CreateOrderInfo, v string) CreateOrderInfo {
			i.Endpoint = v
			return i
		},
	)
	orderInfoCtxLens = L.MakeLens(
		func(i CreateOrderInfo) context.Context { return i.Ctx },
		func(i CreateOrderInfo, v context.Context) CreateOrderInfo {
			i.Ctx = v
			return i
		},
	)
	orderInfoInputLens = L.MakeLens(
		func(i CreateOrderInfo) CreateOrderInput { return i.Input },
		func(i CreateOrderInfo, v CreateOrderInput) CreateOrderInfo {
			i.Input = v
			return i
		},
	)
)
