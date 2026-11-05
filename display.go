package main

import (
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	"github.com/nao1215/markdown"

	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictyBase/fp-go-loom/strutils"
)

const tabPadding = 3

var (
	isNonEmptyFailures = predarrays.IsNonEmpty[LivenessProbeOutcome]()
	defaultHealthyLine = F.Constant(
		[]string{"All GraphQL liveness probes passed."},
	)
	livenessSummaryHeaders = []string{
		"Endpoint",
		"Status",
		"Probes",
		"Passed",
		"Failed",
		"Duration",
	}
	livenessSummaryAlignments = []markdown.TableAlignment{
		markdown.AlignLeft,
		markdown.AlignLeft,
		markdown.AlignRight,
		markdown.AlignRight,
		markdown.AlignRight,
		markdown.AlignRight,
	}
	operationHealthHeaders = []string{
		"Operation",
		"Route Area",
		"Passed",
		"Failed",
		"Latency",
		"Status",
	}
	operationHealthAlignments = []markdown.TableAlignment{
		markdown.AlignLeft,
		markdown.AlignLeft,
		markdown.AlignRight,
		markdown.AlignRight,
		markdown.AlignRight,
		markdown.AlignLeft,
	}
	mapOperationHealthRows = A.Map(formatOperationHealthRow)
	failureHeader          = F.Pipe1(
		[]func(LivenessProbeOutcome) string{
			failureOperationText,
			failureLabelText,
		},
		strutils.ConcatFields[LivenessProbeOutcome](strutils.Space),
	)
	formatFailureBullet = F.Pipe1(
		[]func(LivenessProbeOutcome) string{
			failureHeader,
			failureErrorText,
		},
		strutils.ConcatFields[LivenessProbeOutcome](strutils.Label),
	)
)

func writePlasmidTable(w io.Writer, plasmids []Plasmid) {
	tw := tabwriter.NewWriter(w, 0, 0, tabPadding, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tIN STOCK\tSUMMARY")
	fmt.Fprintln(tw, "--\t----\t--------\t-------")
	for _, p := range plasmids {
		fmt.Fprintf(tw, "%s\t%s\t%t\t%s\n", p.ID, p.Name, p.InStock, p.Summary)
	}
	tw.Flush()
}

func writeStrainTable(w io.Writer, strains []Strain) {
	tw := tabwriter.NewWriter(w, 0, 0, tabPadding, ' ', 0)
	fmt.Fprintln(tw, "ID\tLABEL\tIN STOCK\tSUMMARY")
	fmt.Fprintln(tw, "--\t-----\t--------\t-------")
	for _, s := range strains {
		fmt.Fprintf(tw, "%s\t%s\t%t\t%s\n", s.ID, s.Label, s.InStock, s.Summary)
	}
	tw.Flush()
}

func writeSummary(w io.Writer, result ListPlasmidsResult) {
	if len(result.Plasmids) == 0 {
		fmt.Fprintln(w, "\nNo plasmids found.")
		return
	}
	fmt.Fprintf(w, "\nTotal: %d | Next cursor: %d\n", result.TotalCount, result.NextCursor)
}

func writeStrainSummary(w io.Writer, result ListStrainsResult) {
	if len(result.Strains) == 0 {
		fmt.Fprintln(w, "\nNo strains found.")
		return
	}
	fmt.Fprintf(w, "\nTotal: %d | Next cursor: %d\n", result.TotalCount, result.NextCursor)
}

func writeStockSearchTable(w io.Writer, results []StockSearchResult) {
	tw := tabwriter.NewWriter(w, 0, 0, tabPadding, ' ', 0)
	fmt.Fprintln(tw, "ID\tENTITY\tFIELD\tDISPLAY TEXT\tIN STOCK\tSCORE\tLABEL\tSUMMARY")
	fmt.Fprintln(tw, "--\t------\t-----\t------------\t--------\t-----\t-----\t-------")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%t\t%.4f\t%s\t%s\n",
			r.ID, r.Entity, r.Field, r.DisplayText, r.InStock, r.Score, r.StrainLabel, r.Summary)
	}
	tw.Flush()
}

func writeStockSearchSummary(w io.Writer, result SearchStocksResult) {
	if len(result.Results) == 0 {
		fmt.Fprintln(w, "\nNo stocks found.")
		return
	}
	fmt.Fprintf(w, "\nTotal: %d\n", result.TotalCount)
}

func writeCreatedOrder(w io.Writer, id string) {
	fmt.Fprintf(w, "Order created: %s\n", id)
}

func formatDurationCompact(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}

func formatOperationHealthRow(op OperationHealth) []string {
	return []string{
		markdown.Code(opHealthOperationLens.Get(op)),
		opHealthRouteAreaLens.Get(op),
		fmt.Sprintf("%d/%d", opHealthPassedLens.Get(op), opHealthTotalLens.Get(op)),
		strconv.Itoa(opHealthFailedLens.Get(op)),
		formatDurationCompact(opHealthLatencyLens.Get(op)),
		markdown.Bold(opHealthStatusLens.Get(op)),
	}
}

func failureOperationText(failure LivenessProbeOutcome) string {
	return F.Pipe3(
		failure,
		outcomeOperationLens.Get,
		markdown.Code,
		markdown.Bold,
	)
}

func failureLabelText(failure LivenessProbeOutcome) string {
	return F.Pipe3(
		failure,
		outcomeLabelLens.Get,
		markdown.Code,
		strutils.Parens,
	)
}

func failureErrorText(failure LivenessProbeOutcome) string {
	return F.Pipe2(
		failure,
		outcomeErrorMsgLens.Get,
		markdown.Code,
	)
}

func formatLivenessSummaryRows(report LivenessReport) [][]string {
	return [][]string{
		{
			markdown.Code(reportEndpointLens.Get(report)),
			markdown.Bold(reportStatusLens.Get(report)),
			strconv.Itoa(reportTotalLens.Get(report)),
			strconv.Itoa(reportPassedLens.Get(report)),
			strconv.Itoa(reportFailedLens.Get(report)),
			formatDurationCompact(reportDurationLens.Get(report)),
		},
	}
}

func buildLivenessSummaryTable(report LivenessReport) markdown.TableSet {
	return F.Pipe3(
		markdown.TableSet{},
		tableHeaderLens.Set(livenessSummaryHeaders),
		tableRowsLens.Set(formatLivenessSummaryRows(report)),
		tableAlignmentLens.Set(livenessSummaryAlignments),
	)
}

func buildOperationHealthTable(ops []OperationHealth) markdown.TableSet {
	return F.Pipe3(
		markdown.TableSet{},
		tableHeaderLens.Set(operationHealthHeaders),
		tableRowsLens.Set(F.Pipe1(ops, mapOperationHealthRows)),
		tableAlignmentLens.Set(operationHealthAlignments),
	)
}

func buildFailureSummaryLines(failures []LivenessProbeOutcome) []string {
	return F.Pipe2(
		failures,
		O.FromPredicate(isNonEmptyFailures),
		O.Fold(
			defaultHealthyLine,
			A.Map(formatFailureBullet),
		),
	)
}

func writeLivenessMarkdown(w io.Writer, report LivenessReport) error {
	summaryTable := buildLivenessSummaryTable(report)
	opTable := buildOperationHealthTable(reportOperationsLens.Get(report))
	failureLines := buildFailureSummaryLines(reportFailuresLens.Get(report))
	return markdown.NewMarkdown(w, markdown.WithBlockSpacing()).
		H1("GraphQL API Liveness Report").
		Table(summaryTable).
		H2("Operation Health").
		Table(opTable).
		H2("Failure Summary").
		BulletList(failureLines...).
		Build()
}
