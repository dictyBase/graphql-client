package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const defaultLimit = 10

const defaultEndpoint = "https://graphql.dictybase.dev/graphql"

// CLI flag names
const (
	flagEndpoint       = "endpoint"
	flagLimit          = "limit"
	flagType           = "type"
	flagName           = "name"
	flagSummary        = "summary"
	flagLabel          = "label"
	flagConsumer       = "consumer"
	flagPayer          = "payer"
	flagPurchaser      = "purchaser"
	flagItems          = "items"
	flagCourier        = "courier"
	flagCourierAccount = "courier-account"
	flagPayment        = "payment"
	flagComments       = "comments"
	flagPONum          = "purchase-order-num"
	flagStatus         = "status"
)

const usageEndpoint = "GraphQL API endpoint URL"

// Fake defaults for the create-order command; only consumer/payer emails
// and stock item ids need real values.
const (
	defaultCourier      = "FedEx"
	defaultCourierAcct  = "FAKE-ACCT-0001"
	defaultPayment      = "credit_card"
	defaultComments     = "automated CLI test order"
	defaultPONum        = "PO-FAKE-0001"
	defaultPurchaser    = "fake-purchaser@example.org"
	usageConsumerEmail  = "Real consumer email address, receives the invoice"
	usagePayerEmail     = "Real payer email address, rendered in the invoice"
	usagePurchaserEmail = "Purchaser email address (unused in email)"
	usageOrderItems     = "Real stock item ids (DBS/DBP accessions), repeat or comma-separate"
)

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

func newCreateOrderCommand() *cli.Command {
	return &cli.Command{
		Name:   "create-order",
		Usage:  "Create a stock order through the GraphQL endpoint",
		Action: RunCreateOrderCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEndpoint,
				Value: defaultEndpoint,
				Usage: usageEndpoint,
			},
			&cli.StringFlag{
				Name:     flagConsumer,
				Usage:    usageConsumerEmail,
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagPayer,
				Usage:    usagePayerEmail,
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     flagItems,
				Usage:    usageOrderItems,
				Required: true,
			},
			&cli.StringFlag{
				Name:  flagPurchaser,
				Value: defaultPurchaser,
				Usage: usagePurchaserEmail,
			},
			&cli.StringFlag{
				Name:  flagCourier,
				Value: defaultCourier,
				Usage: "Courier name",
			},
			&cli.StringFlag{
				Name:  flagCourierAccount,
				Value: defaultCourierAcct,
				Usage: "Courier account number",
			},
			&cli.StringFlag{
				Name:  flagPayment,
				Value: defaultPayment,
				Usage: "Payment method",
			},
			&cli.StringFlag{
				Name:  flagComments,
				Value: defaultComments,
				Usage: "Order comments",
			},
			&cli.StringFlag{
				Name:  flagPONum,
				Value: defaultPONum,
				Usage: "Purchase order number",
			},
			&cli.StringFlag{
				Name:  flagStatus,
				Value: string(StatusInPreparation),
				Usage: "Order status (IN_PREPARATION, GROWING, CANCELLED, SHIPPED)",
			},
		},
	}
}

func newListPlasmidsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-plasmids",
		Usage:  "List plasmids from the GraphQL endpoint",
		Action: RunListPlasmidCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEndpoint,
				Value: defaultEndpoint,
				Usage: usageEndpoint,
			},
			&cli.IntFlag{
				Name:  flagLimit,
				Value: defaultLimit,
				Usage: "Number of plasmid entries to fetch",
			},
			&cli.StringFlag{
				Name:  flagType,
				Value: string(PlasmidTypeAll),
				Usage: "Plasmid type to filter (ALL, REGULAR, GOLDEN_BRAID)",
			},
		},
	}
}

func newListFilteredPlasmidsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-filtered-plasmids",
		Usage:  "List plasmids with attribute-level filtering from the GraphQL endpoint",
		Action: RunListFilteredPlasmidCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEndpoint,
				Value: defaultEndpoint,
				Usage: usageEndpoint,
			},
			&cli.IntFlag{
				Name:  flagLimit,
				Value: defaultLimit,
				Usage: "Number of plasmid entries to fetch",
			},
			&cli.StringFlag{
				Name:  flagType,
				Value: string(PlasmidTypeAll),
				Usage: "Plasmid type to filter (ALL, REGULAR, GOLDEN_BRAID)",
			},
			&cli.StringFlag{
				Name:  flagName,
				Usage: "Filter by plasmid name",
			},
			&cli.StringFlag{
				Name:  flagSummary,
				Usage: "Filter by plasmid summary",
			},
		},
	}
}

func newListStrainsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-strains",
		Usage:  "List strains from the GraphQL endpoint",
		Action: RunListStrainCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEndpoint,
				Value: defaultEndpoint,
				Usage: usageEndpoint,
			},
			&cli.IntFlag{
				Name:  flagLimit,
				Value: defaultLimit,
				Usage: "Number of strain entries to fetch",
			},
			&cli.StringFlag{
				Name:  flagType,
				Value: string(StrainTypeAll),
				Usage: "Strain type to filter (ALL, REGULAR, GWDI, BACTERIAL)",
			},
		},
	}
}

func newListFilteredStrainsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-filtered-strains",
		Usage:  "List strains with attribute-level filtering from the GraphQL endpoint",
		Action: RunListFilteredStrainCLI,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEndpoint,
				Value: defaultEndpoint,
				Usage: usageEndpoint,
			},
			&cli.IntFlag{
				Name:  flagLimit,
				Value: defaultLimit,
				Usage: "Number of strain entries to fetch",
			},
			&cli.StringFlag{
				Name:  flagType,
				Value: string(StrainTypeAll),
				Usage: "Strain type to filter (ALL, REGULAR, GWDI, BACTERIAL)",
			},
			&cli.StringFlag{
				Name:  flagLabel,
				Usage: "Filter by strain label",
			},
			&cli.StringFlag{
				Name:  flagSummary,
				Usage: "Filter by strain summary",
			},
		},
	}
}

func main() {
	if err := newRootCommand().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
