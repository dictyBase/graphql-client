package main

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"

	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predstrings "github.com/dictyBase/fp-go-loom/predicate/strings"
)

type (
	createOrderCLIArgs struct {
		Context context.Context
		Command *cli.Command
	}
	createOrderValidatedArgs struct {
		Context context.Context
		Command *cli.Command
		Input   CreateOrderInput
	}
	createOrderFetchArgs struct {
		Context context.Context
		Client  *graphql.Client
		Input   CreateOrderInput
	}
)

// createOrderOptions carries the raw flag values of the create-order command.
type createOrderOptions struct {
	Consumer         string
	Payer            string
	Purchaser        string
	Items            []string
	Courier          string
	CourierAccount   string
	Payment          string
	Comments         string
	PurchaseOrderNum string
	Status           string
}

func orderOptionsFromCommand(cmd *cli.Command) createOrderOptions {
	return createOrderOptions{
		Consumer:         cmd.String(flagConsumer),
		Payer:            cmd.String(flagPayer),
		Purchaser:        cmd.String(flagPurchaser),
		Items:            cmd.StringSlice(flagItems),
		Courier:          cmd.String(flagCourier),
		CourierAccount:   cmd.String(flagCourierAccount),
		Payment:          cmd.String(flagPayment),
		Comments:         cmd.String(flagComments),
		PurchaseOrderNum: cmd.String(flagPONum),
		Status:           cmd.String(flagStatus),
	}
}

func ParseStatus(s string) E.Either[error, StatusEnum] {
	switch st := StatusEnum(s); st {
	case StatusInPreparation, StatusGrowing, StatusCancelled, StatusShipped:
		return E.Right[error](st)
	default:
		return E.Left[StatusEnum](errInvalidStatus(s))
	}
}

func validateConsumerEmail(opts createOrderOptions) E.Either[error, createOrderOptions] {
	return F.Pipe2(
		opts.Consumer,
		E.FromPredicate(
			predstrings.HasAtSign,
			func(string) error { return errInvalidEmail("consumer") },
		),
		E.Map[error](func(string) createOrderOptions { return opts }),
	)
}

func validatePayerEmail(opts createOrderOptions) E.Either[error, createOrderOptions] {
	return F.Pipe2(
		opts.Payer,
		E.FromPredicate(
			predstrings.HasAtSign,
			func(string) error { return errInvalidEmail("payer") },
		),
		E.Map[error](func(string) createOrderOptions { return opts }),
	)
}

func validatePurchaserEmail(opts createOrderOptions) E.Either[error, createOrderOptions] {
	return F.Pipe2(
		opts.Purchaser,
		E.FromPredicate(
			predstrings.HasAtSign,
			func(string) error { return errInvalidEmail("purchaser") },
		),
		E.Map[error](func(string) createOrderOptions { return opts }),
	)
}

func validateOrderItems(opts createOrderOptions) E.Either[error, createOrderOptions] {
	return F.Pipe2(
		opts.Items,
		E.FromPredicate(
			predarrays.IsNonEmpty[string](),
			func([]string) error { return errNoOrderItems },
		),
		E.Map[error](func([]string) createOrderOptions { return opts }),
	)
}

func toCreateOrderInput(opts createOrderOptions) E.Either[error, CreateOrderInput] {
	return F.Pipe2(
		opts.Status,
		ParseStatus,
		E.Map[error](func(status StatusEnum) CreateOrderInput {
			return CreateOrderInput{
				Courier:          opts.Courier,
				CourierAccount:   opts.CourierAccount,
				Comments:         opts.Comments,
				Payment:          opts.Payment,
				PurchaseOrderNum: opts.PurchaseOrderNum,
				Status:           status,
				Consumer:         opts.Consumer,
				Payer:            opts.Payer,
				Purchaser:        opts.Purchaser,
				Items:            opts.Items,
				ConsumerInfo:     fakeConsumerInfo(),
				PayerInfo:        fakePayerInfo(),
			}
		}),
	)
}

func buildCreateOrderInput(opts createOrderOptions) E.Either[error, CreateOrderInput] {
	return F.Pipe5(
		opts,
		validateConsumerEmail,
		E.Chain(validatePayerEmail),
		E.Chain(validatePurchaserEmail),
		E.Chain(validateOrderItems),
		E.Chain(toCreateOrderInput),
	)
}

func toCreateOrderValidatedArgs(
	input createOrderCLIArgs,
) IOE.IOEither[error, createOrderValidatedArgs] {
	return F.Pipe3(
		orderOptionsFromCommand(input.Command),
		buildCreateOrderInput,
		IOE.FromEither[error, CreateOrderInput],
		IOE.Map[error](func(orderInput CreateOrderInput) createOrderValidatedArgs {
			return createOrderValidatedArgs{
				Context: input.Context,
				Command: input.Command,
				Input:   orderInput,
			}
		}),
	)
}

func mapCreateOrderFetchArgs(input createOrderValidatedArgs) createOrderFetchArgs {
	return createOrderFetchArgs{
		Context: input.Context,
		Client:  clientFromCommand(input.Command),
		Input:   input.Input,
	}
}

func submitCreateOrder(input createOrderFetchArgs) IOE.IOEither[error, string] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*CreateOrderMutation, error) {
			mutation := new(CreateOrderMutation)
			vars := map[string]any{gqlVarInput: input.Input}
			return mutation, input.Client.Mutate(input.Context, mutation, vars)
		}),
		IOE.MapLeft[*CreateOrderMutation](func(err error) error {
			return fmt.Errorf("graphql mutation failed: %w", err)
		}),
		IOE.Map[error](func(mutation *CreateOrderMutation) string {
			return string(mutation.CreateOrder.ID)
		}),
	)
}

func RunCreateOrderCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		createOrderCLIArgs{Context: ctx, Command: cmd},
		IOE.Of[error, createOrderCLIArgs],
		IOE.Chain(toCreateOrderValidatedArgs),
		IOE.Map[error](mapCreateOrderFetchArgs),
		IOE.Chain(submitCreateOrder),
		ioeutils.ToEither[error, string],
		E.Fold(
			func(err error) error { return err },
			func(orderID string) error {
				writeCreatedOrder(os.Stdout, orderID)
				return nil
			},
		),
	)
}
