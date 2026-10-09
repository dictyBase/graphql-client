package main

import (
	"slices"

	"github.com/urfave/cli/v3"
)

// endpointFlag returns the shared --endpoint flag definition.
func endpointFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  flagEndpoint,
		Value: defaultEndpoint,
		Usage: usageEndpoint,
	}
}

// limitFlag returns the --limit flag with a per-command usage string.
func limitFlag(usage string) *cli.IntFlag {
	return &cli.IntFlag{
		Name:  flagLimit,
		Value: defaultLimit,
		Usage: usage,
	}
}

// orderContactFlags returns the order party flags: who receives and pays
// the invoice, plus the optional purchaser and the requested stock items.
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

func newCreateOrderCommand() *cli.Command {
	return &cli.Command{
		Name:   "create-order",
		Usage:  "Create a stock order through the GraphQL endpoint",
		Action: RunCreateOrderCLI,
		Flags: slices.Concat(
			orderContactFlags(),
			orderShipmentFlags(),
			orderRecordFlags(),
		),
	}
}

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
