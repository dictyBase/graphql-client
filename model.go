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

type StrainType string

func (StrainType) GetGraphQLType() string { return "StrainType" }

const (
	StrainTypeAll       StrainType = "ALL"
	StrainTypeRegular   StrainType = "REGULAR"
	StrainTypeGwdi      StrainType = "GWDI"
	StrainTypeBacterial StrainType = "BACTERIAL"
)

var strainTypeMap = map[string]StrainType{
	"ALL":       StrainTypeAll,
	"REGULAR":   StrainTypeRegular,
	"GWDI":      StrainTypeGwdi,
	"BACTERIAL": StrainTypeBacterial,
}

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
