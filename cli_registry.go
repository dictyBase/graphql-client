package main

import (
	A "github.com/IBM/fp-go/v2/array"
	"github.com/urfave/cli/v3"
)

// endpointFlag returns the shared --endpoint flag definition.
// fp-go:allow-struct-literal urfave/cli flag descriptor
func endpointFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  flagEndpoint,
		Value: defaultEndpoint,
		Usage: usageEndpoint,
	}
}

// limitFlag returns the --limit flag with a per-command usage string.
// fp-go:allow-struct-literal urfave/cli flag descriptor
func limitFlag(usage string) *cli.IntFlag {
	return &cli.IntFlag{
		Name:  flagLimit,
		Value: defaultLimit,
		Usage: usage,
	}
}

// orderContactFlags returns the order party flags: who receives and pays
// the invoice, plus the optional purchaser and the requested stock items.
// fp-go:allow-struct-literal urfave/cli flag descriptors
func orderContactFlags() []cli.Flag {
	return []cli.Flag{
		endpointFlag(),
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
	}
}

// orderShipmentFlags returns the delivery flags of the create-order command.
// fp-go:allow-struct-literal urfave/cli flag descriptors
func orderShipmentFlags() []cli.Flag {
	return []cli.Flag{
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
	}
}

// orderRecordFlags returns the bookkeeping flags of the create-order command.
// fp-go:allow-struct-literal urfave/cli flag descriptors
func orderRecordFlags() []cli.Flag {
	return []cli.Flag{
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
	}
}

// fp-go:allow-struct-literal urfave/cli command and flag descriptors
func newSearchStocksCommand() *cli.Command {
	return &cli.Command{
		Name:   "search-stocks",
		Usage:  "Search stocks (strains and plasmids) from the GraphQL endpoint",
		Action: RunSearchStockCLI,
		Flags: []cli.Flag{
			endpointFlag(),
			&cli.StringFlag{
				Name:     flagQuery,
				Usage:    usageStockQuery,
				Required: true,
			},
			limitFlag("Number of search results to fetch"),
			&cli.StringFlag{
				Name:  flagEntity,
				Value: string(StockEntityAll),
				Usage: "Stock entity to search (ALL, STRAIN, PLASMID)",
			},
		},
	}
}

// fp-go:allow-struct-literal urfave/cli command descriptor
func newCreateOrderCommand() *cli.Command {
	return &cli.Command{
		Name:   "create-order",
		Usage:  "Create a stock order through the GraphQL endpoint",
		Action: RunCreateOrderCLI,
		Flags: A.ArrayConcatAll(
			orderContactFlags(),
			orderShipmentFlags(),
			orderRecordFlags(),
		),
	}
}

// fp-go:allow-struct-literal urfave/cli command and flag descriptors
func newLivenessCheckCommand() *cli.Command {
	return &cli.Command{
		Name:   "liveness-check",
		Usage:  "Run a liveness check on GraphQL API operations and write a Markdown report",
		Action: RunLivenessCheckCLI,
		Flags: []cli.Flag{
			endpointFlag(),
			&cli.DurationFlag{
				Name:  flagTimeout,
				Value: defaultLivenessTimeout,
				Usage: "Per-request timeout for each GraphQL liveness probe",
			},
			&cli.StringFlag{
				Name:  flagOutput,
				Value: defaultLivenessReportPath,
				Usage: "Path of the Markdown report file to write",
			},
		},
	}
}

// fp-go:allow-struct-literal urfave/cli command and flag descriptors
func newListPlasmidsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-plasmids",
		Usage:  "List plasmids from the GraphQL endpoint",
		Action: RunListPlasmidCLI,
		Flags: []cli.Flag{
			endpointFlag(),
			limitFlag("Number of plasmid entries to fetch"),
			&cli.StringFlag{
				Name:  flagType,
				Value: string(PlasmidTypeAll),
				Usage: "Plasmid type to filter (ALL, REGULAR, GOLDEN_BRAID)",
			},
		},
	}
}

// fp-go:allow-struct-literal urfave/cli command and flag descriptors
func newListFilteredPlasmidsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-filtered-plasmids",
		Usage:  "List plasmids with attribute-level filtering from the GraphQL endpoint",
		Action: RunListFilteredPlasmidCLI,
		Flags: []cli.Flag{
			endpointFlag(),
			limitFlag("Number of plasmid entries to fetch"),
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

// fp-go:allow-struct-literal urfave/cli command and flag descriptors
func newListStrainsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-strains",
		Usage:  "List strains from the GraphQL endpoint",
		Action: RunListStrainCLI,
		Flags: []cli.Flag{
			endpointFlag(),
			limitFlag("Number of strain entries to fetch"),
			&cli.StringFlag{
				Name:  flagType,
				Value: string(StrainTypeAll),
				Usage: "Strain type to filter (ALL, REGULAR, GWDI, BACTERIAL)",
			},
		},
	}
}

// fp-go:allow-struct-literal urfave/cli command and flag descriptors
func newListFilteredStrainsCommand() *cli.Command {
	return &cli.Command{
		Name:   "list-filtered-strains",
		Usage:  "List strains with attribute-level filtering from the GraphQL endpoint",
		Action: RunListFilteredStrainCLI,
		Flags: []cli.Flag{
			endpointFlag(),
			limitFlag("Number of strain entries to fetch"),
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
