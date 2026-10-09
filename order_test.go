package main

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// orderTestCommand builds a create-order command whose flags carry the
// valid default values.
func orderTestCommand() *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagConsumer, Value: "shipper@example.org"},
			&cli.StringFlag{Name: flagPayer, Value: "payer@example.org"},
			&cli.StringFlag{Name: flagPurchaser, Value: "fake-purchaser@example.org"},
			&cli.StringSliceFlag{Name: flagItems, Value: []string{"DBS0235559"}},
			&cli.StringFlag{Name: flagCourier, Value: "FedEx"},
			&cli.StringFlag{Name: flagCourierAccount, Value: "FAKE-ACCT-0001"},
			&cli.StringFlag{Name: flagPayment, Value: "credit_card"},
			&cli.StringFlag{Name: flagComments, Value: "automated CLI test order"},
			&cli.StringFlag{Name: flagPONum, Value: "PO-FAKE-0001"},
			&cli.StringFlag{Name: flagStatus, Value: string(StatusInPreparation)},
			&cli.StringFlag{Name: flagEndpoint, Value: "https://unit.test/graphql"},
		},
	}
}

func orderTestArgs(t *testing.T) createOrderCLIArgs {
	t.Helper()

	return createOrderCLIArgs{Context: context.Background(), Command: orderTestCommand()}
}

func TestParseStatusFP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StatusEnum
		wantErr bool
	}{
		{
			name:  "IN_PREPARATION status",
			input: "IN_PREPARATION",
			want:  StatusInPreparation,
		},
		{
			name:  "GROWING status",
			input: "GROWING",
			want:  StatusGrowing,
		},
		{
			name:  "CANCELLED status",
			input: "CANCELLED",
			want:  StatusCancelled,
		},
		{
			name:  "SHIPPED status",
			input: "SHIPPED",
			want:  StatusShipped,
		},
		{
			name:    "empty status is invalid",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid status",
			input:   "LOST",
			wantErr: true,
		},
		{
			name:    "lowercase status is invalid",
			input:   "shipped",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseStatus(tt.input)
			val, err := E.Unwrap(result)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, val)
		})
	}
}

func TestToCreateOrderInput(t *testing.T) {
	args := orderTestArgs(t)
	cmd := args.Command
	result := toCreateOrderInput(args)
	val, err := E.Unwrap(result)
	require.NoError(t, err)
	require.Equal(t, cmd.String(flagConsumer), val.Consumer)
	require.Equal(t, cmd.String(flagPayer), val.Payer)
	require.Equal(t, cmd.String(flagPurchaser), val.Purchaser)
	require.Equal(t, cmd.StringSlice(flagItems), val.Items)
	require.Equal(t, cmd.String(flagCourier), val.Courier)
	require.Equal(t, cmd.String(flagCourierAccount), val.CourierAccount)
	require.Equal(t, cmd.String(flagPayment), val.Payment)
	require.Equal(t, cmd.String(flagComments), val.Comments)
	require.Equal(t, cmd.String(flagPONum), val.PurchaseOrderNum)
	require.Equal(t, StatusInPreparation, val.Status)
	require.NotNil(t, val.ConsumerInfo, "should attach fake consumer info")
	require.NotNil(t, val.PayerInfo, "should attach fake payer info")
}

func TestValidateEmail(t *testing.T) {
	const (
		consumerField  = "consumer"
		payerField     = "payer"
		purchaserField = "purchaser"
	)
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr string
	}{
		{
			name:    "invalid consumer email",
			field:   consumerField,
			value:   "not-an-email",
			wantErr: consumerField,
		},
		{
			name:    "empty payer email",
			field:   payerField,
			value:   "",
			wantErr: payerField,
		},
		{
			name:    "invalid purchaser email",
			field:   purchaserField,
			value:   "buyer",
			wantErr: purchaserField,
		},
		{
			name:  "valid email passes",
			field: consumerField,
			value: "shipper@example.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validateEmail(tt.field)
			val, err := E.Unwrap(validate(tt.value))
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.value, val)
		})
	}
}

func TestValidateOrderEmails(t *testing.T) {
	t.Run("invalid flag fails the whole tuple", func(t *testing.T) {
		args := orderTestArgs(t)
		require.NoError(t, args.Command.Set(flagPayer, "payer"))
		_, err := E.Unwrap(validateOrderEmails(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "payer")
	})

	t.Run("all three emails fail", func(t *testing.T) {
		args := orderTestArgs(t)
		require.NoError(t, args.Command.Set(flagConsumer, "c"))
		_, err := E.Unwrap(validateOrderEmails(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "consumer")
	})

	t.Run("valid flags pass through the state", func(t *testing.T) {
		args := orderTestArgs(t)
		result, err := E.Unwrap(validateOrderEmails(args))
		require.NoError(t, err)
		require.Equal(t, args, result)
	})
}

func TestValidateOrderItems(t *testing.T) {
	t.Run("empty item list fails", func(t *testing.T) {
		args := createOrderCLIArgs{
			Context: context.Background(),
			Command: &cli.Command{
				Flags: []cli.Flag{&cli.StringSliceFlag{Name: flagItems}},
			},
		}
		_, err := E.Unwrap(validateOrderItems(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "item")
	})

	t.Run("non-empty item list passes", func(t *testing.T) {
		args := orderTestArgs(t)
		_, err := E.Unwrap(validateOrderItems(args))
		require.NoError(t, err)
	})
}

func TestToCreateOrderInputInvalidStatus(t *testing.T) {
	args := orderTestArgs(t)
	require.NoError(t, args.Command.Set(flagStatus, "LOST"))
	_, err := E.Unwrap(toCreateOrderInput(args))
	require.Error(t, err)
	require.Contains(t, err.Error(), "status")
}
