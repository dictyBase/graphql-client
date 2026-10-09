package main

import "github.com/hasura/go-graphql-client"

// GraphQL variable keys
const (
	gqlVarCursor = "cursor"
	gqlVarFilter = "filter"
	gqlVarInput  = "input"
)

type PlasmidType string

func (PlasmidType) GetGraphQLType() string { return "PlasmidType" }

const (
	PlasmidTypeAll         PlasmidType = "ALL"
	PlasmidTypeRegular     PlasmidType = "REGULAR"
	PlasmidTypeGoldenBraid PlasmidType = "GOLDEN_BRAID"
)

type PlasmidListFilter struct {
	PlasmidType PlasmidType `json:"plasmid_type"`
}

func (PlasmidListFilter) GetGraphQLType() string { return "PlasmidListFilter" }

type PlasmidAttributeFilter struct {
	PlasmidType PlasmidType `json:"plasmid_type"`
	Name        *string     `json:"name,omitempty"`
	Summary     *string     `json:"summary,omitempty"`
}

func (PlasmidAttributeFilter) GetGraphQLType() string { return "PlasmidListFilter" }

type Plasmid struct {
	ID      graphql.ID `graphql:"id"`
	Name    string     `graphql:"name"`
	Summary string     `graphql:"summary"`
	InStock bool       `graphql:"in_stock"`
}

type ListPlasmidsResult struct {
	NextCursor int64
	TotalCount int
	Plasmids   []Plasmid
}

type ListPlasmidsQuery struct {
	ListPlasmids struct {
		NextCursor int64     `graphql:"nextCursor"`
		TotalCount int       `graphql:"totalCount"`
		Plasmids   []Plasmid `graphql:"plasmids"`
	} `graphql:"listPlasmids(cursor: $cursor, limit: $limit, filter: $filter)"`
}

func (q *ListPlasmidsQuery) toResult() ListPlasmidsResult {
	return ListPlasmidsResult{
		NextCursor: q.ListPlasmids.NextCursor,
		TotalCount: q.ListPlasmids.TotalCount,
		Plasmids:   q.ListPlasmids.Plasmids,
	}
}

type ListFilteredPlasmidsQuery struct {
	ListPlasmids struct {
		NextCursor int64     `graphql:"nextCursor"`
		TotalCount int       `graphql:"totalCount"`
		Plasmids   []Plasmid `graphql:"plasmids"`
	} `graphql:"listPlasmids(cursor: $cursor, limit: $limit, filter: $filter)"`
}

func (q *ListFilteredPlasmidsQuery) toResult() ListPlasmidsResult {
	return ListPlasmidsResult{
		NextCursor: q.ListPlasmids.NextCursor,
		TotalCount: q.ListPlasmids.TotalCount,
		Plasmids:   q.ListPlasmids.Plasmids,
	}
}

type StrainType string

func (StrainType) GetGraphQLType() string { return "StrainType" }

const (
	StrainTypeAll       StrainType = "ALL"
	StrainTypeRegular   StrainType = "REGULAR"
	StrainTypeGwdi      StrainType = "GWDI"
	StrainTypeBacterial StrainType = "BACTERIAL"
)

type StrainListFilter struct {
	StrainType StrainType `json:"strain_type"`
}

func (StrainListFilter) GetGraphQLType() string { return "StrainListFilter" }

type StrainAttributeFilter struct {
	StrainType StrainType `json:"strain_type"`
	Label      *string    `json:"label,omitempty"`
	Summary    *string    `json:"summary,omitempty"`
}

func (StrainAttributeFilter) GetGraphQLType() string { return "StrainListFilter" }

type Strain struct {
	ID      graphql.ID `graphql:"id"`
	Label   string     `graphql:"label"`
	Summary string     `graphql:"summary"`
	InStock bool       `graphql:"in_stock"`
}

type ListStrainsResult struct {
	NextCursor int64
	TotalCount int
	Strains    []Strain
}

type ListStrainsQuery struct {
	ListStrains struct {
		NextCursor int64    `graphql:"nextCursor"`
		TotalCount int      `graphql:"totalCount"`
		Strains    []Strain `graphql:"strains"`
	} `graphql:"listStrains(cursor: $cursor, limit: $limit, filter: $filter)"`
}

func (q *ListStrainsQuery) toResult() ListStrainsResult {
	return ListStrainsResult{
		NextCursor: q.ListStrains.NextCursor,
		TotalCount: q.ListStrains.TotalCount,
		Strains:    q.ListStrains.Strains,
	}
}

type ListFilteredStrainsQuery struct {
	ListStrains struct {
		NextCursor int64    `graphql:"nextCursor"`
		TotalCount int      `graphql:"totalCount"`
		Strains    []Strain `graphql:"strains"`
	} `graphql:"listStrains(cursor: $cursor, limit: $limit, filter: $filter)"`
}

func (q *ListFilteredStrainsQuery) toResult() ListStrainsResult {
	return ListStrainsResult{
		NextCursor: q.ListStrains.NextCursor,
		TotalCount: q.ListStrains.TotalCount,
		Strains:    q.ListStrains.Strains,
	}
}

type StatusEnum string

func (StatusEnum) GetGraphQLType() string { return "StatusEnum" }

const (
	StatusInPreparation StatusEnum = "IN_PREPARATION"
	StatusGrowing       StatusEnum = "GROWING"
	StatusCancelled     StatusEnum = "CANCELLED"
	StatusShipped       StatusEnum = "SHIPPED"
)

type UserInfoInput struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Organization  string `json:"organization"`
	FirstAddress  string `json:"first_address"`
	SecondAddress string `json:"second_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Zipcode       string `json:"zipcode"`
	Country       string `json:"country"`
	Phone         string `json:"phone"`
}

func (UserInfoInput) GetGraphQLType() string { return "UserInfoInput" }

type CreateOrderInput struct {
	Courier          string         `json:"courier"`
	CourierAccount   string         `json:"courier_account"`
	Comments         string         `json:"comments"`
	Payment          string         `json:"payment"`
	PurchaseOrderNum string         `json:"purchase_order_num"`
	Status           StatusEnum     `json:"status"`
	Consumer         string         `json:"consumer"`
	Payer            string         `json:"payer"`
	Purchaser        string         `json:"purchaser"`
	Items            []string       `json:"items"`
	ConsumerInfo     *UserInfoInput `json:"consumer_info,omitempty"`
	PayerInfo        *UserInfoInput `json:"payer_info,omitempty"`
}

func (CreateOrderInput) GetGraphQLType() string { return "CreateOrderInput" }

type CreateOrderMutation struct {
	CreateOrder struct {
		ID graphql.ID `graphql:"id"`
	} `graphql:"createOrder(input: $input)"`
}

// fakeConsumerInfo is the fake consumer profile attached to test orders.
func fakeConsumerInfo() *UserInfoInput {
	return &UserInfoInput{
		FirstName:    "Test",
		LastName:     "Consumer",
		Organization: "Test Lab",
		FirstAddress: "123 Test Street",
		City:         "Testville",
		State:        "TS",
		Zipcode:      "00000",
		Country:      "USA",
		Phone:        "555-0100",
	}
}

// fakePayerInfo is the fake payer profile attached to test orders.
func fakePayerInfo() *UserInfoInput {
	return &UserInfoInput{
		FirstName:    "Test",
		LastName:     "Payer",
		Organization: "Test Lab",
		FirstAddress: "123 Test Street",
		City:         "Testville",
		State:        "TS",
		Zipcode:      "00000",
		Country:      "USA",
		Phone:        "555-0200",
	}
}
