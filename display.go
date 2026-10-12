package main

import (
	"fmt"
	"io"
	"text/tabwriter"
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
