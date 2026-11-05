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
		markdown.Code(op.Operation),
		op.RouteArea,
		fmt.Sprintf("%d/%d", op.Passed, op.Total),
		strconv.Itoa(op.Failed),
		formatDurationCompact(op.Latency),
		markdown.Bold(op.Status),
	}
}

func formatFailureBullet(failure LivenessProbeOutcome) string {
	opCode := markdown.Code(failure.Probe.Operation)
	probeLabel := markdown.Code(failure.Probe.Label)
	opHeader := strutils.Space.Concat(
		markdown.Bold(opCode),
		strutils.Parens(probeLabel),
	)
	return strutils.Label.Concat(
		opHeader,
		markdown.Code(failure.ErrorMsg),
	)
}

func buildLivenessSummaryTable(report LivenessReport) markdown.TableSet {
	return markdown.TableSet{
		Header: []string{
			"Endpoint",
			"Status",
			"Probes",
			"Passed",
			"Failed",
			"Duration",
		},
		Rows: [][]string{
			{
				markdown.Code(report.Endpoint),
				markdown.Bold(report.Status),
				strconv.Itoa(report.Total),
				strconv.Itoa(report.Passed),
				strconv.Itoa(report.Failed),
				formatDurationCompact(report.Duration),
			},
		},
		Alignment: []markdown.TableAlignment{
			markdown.AlignLeft,
			markdown.AlignLeft,
			markdown.AlignRight,
			markdown.AlignRight,
			markdown.AlignRight,
			markdown.AlignRight,
		},
	}
}

func buildOperationHealthTable(ops []OperationHealth) markdown.TableSet {
	return markdown.TableSet{
		Header: []string{
			"Operation",
			"Route Area",
			"Passed",
			"Failed",
			"Latency",
			"Status",
		},
		Rows: F.Pipe1(ops, A.Map(formatOperationHealthRow)),
		Alignment: []markdown.TableAlignment{
			markdown.AlignLeft,
			markdown.AlignLeft,
			markdown.AlignRight,
			markdown.AlignRight,
			markdown.AlignRight,
			markdown.AlignLeft,
		},
	}
}

func buildFailureSummaryLines(failures []LivenessProbeOutcome) []string {
	return F.Pipe2(
		failures,
		O.FromPredicate(predarrays.IsNonEmpty[LivenessProbeOutcome]()),
		O.Fold(
			func() []string {
				return []string{"All GraphQL liveness probes passed."}
			},
			A.Map(formatFailureBullet),
		),
	)
}

func writeLivenessMarkdown(w io.Writer, report LivenessReport) error {
	summaryTable := buildLivenessSummaryTable(report)
	opTable := buildOperationHealthTable(report.Operations)
	failureLines := buildFailureSummaryLines(report.Failures)
	return markdown.NewMarkdown(w, markdown.WithBlockSpacing()).
		H1("GraphQL API Liveness Report").
		Table(summaryTable).
		H2("Operation Health").
		Table(opTable).
		H2("Failure Summary").
		BulletList(failureLines...).
		Build()
}
