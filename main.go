package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// newRootCommand assembles the full command tree.
func newRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "graphql-client",
		Usage: "CLI to query stock data from a GraphQL endpoint",
		Commands: []*cli.Command{
			newListPlasmidsCommand(),
			newListFilteredPlasmidsCommand(),
			newListStrainsCommand(),
			newListFilteredStrainsCommand(),
			newCreateOrderCommand(),
		},
	}
}

func main() {
	app := newRootCommand()

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
