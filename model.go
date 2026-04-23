package main

import "github.com/hasura/go-graphql-client"

type PlasmidType string

func (PlasmidType) GetGraphQLType() string { return "PlasmidType" }

const (
	PlasmidTypeAll         PlasmidType = "ALL"
	PlasmidTypeRegular     PlasmidType = "REGULAR"
	PlasmidTypeGoldenBraid PlasmidType = "GOLDEN_BRAID"
)

var plasmidTypeMap = map[string]PlasmidType{
	"ALL":          PlasmidTypeAll,
	"REGULAR":      PlasmidTypeRegular,
	"GOLDEN_BRAID": PlasmidTypeGoldenBraid,
}

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
