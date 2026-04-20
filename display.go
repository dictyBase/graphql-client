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

func writeSummary(w io.Writer, result ListPlasmidsResult) {
	if len(result.Plasmids) == 0 {
		fmt.Fprintln(w, "\nNo plasmids found.")
		return
	}
	fmt.Fprintf(w, "\nTotal: %d | Next cursor: %d\n", result.TotalCount, result.NextCursor)
}
