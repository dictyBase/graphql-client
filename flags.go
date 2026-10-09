package main

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
