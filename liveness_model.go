package main

import "time"

const (
	livenessCatalogLimit   = 12
	livenessSearchLimit    = 50
	livenessPhenotypeLimit = 50
	defaultLivenessTimeout = 30 * time.Second
)

const (
	opStrainList                = "StrainList"
	opPlasmidListFilter         = "PlasmidListFilter"
	opSearchStocks              = "SearchStocks"
	opListStockSuggestions      = "ListStockSuggestions"
	opPublication               = "Publication"
	opContentBySlug             = "ContentBySlug"
	opStrain                    = "Strain"
	opPlasmid                   = "Plasmid"
	opListStrainsWithPhenotype  = "ListStrainsWithPhenotype"
	opListPhenotypes            = "ListPhenotypes"
	opListPhenotypeEnvironments = "ListPhenotypeEnvironments"
	opListPhenotypeAssays       = "ListPhenotypeAssays"
)

const (
	routeCatalogAndStrains  = "/strains, /catalog"
	routeCatalogAndPlasmids = "/plasmids, /catalog"
	routeCatalogSearch      = "/catalog?q=..."
	routeNavbarSearch       = "Navbar search"
	routeReferencePreview   = "Reference preview"
	routeInfoAndCommunity   = "Info & community pages"
	routeStrainDetail       = "/strains/:id"
	routePlasmidDetail      = "/plasmids/:id"
	routePhenotypeList      = "/phenotypes/:name"
	routeEditablePhenotype  = "Editable phenotype form"
)

const (
	livenessStatusHealthy  = "HEALTHY"
	livenessStatusDegraded = "DEGRADED"
	livenessStatusDown     = "DOWN"
	probeStatusPass        = "PASS"
	probeStatusFail        = "FAIL"
	rootFieldSearchStocks  = "searchStocks"
	sampleQueryAct         = "act"
	sampleSearchGrowth     = "growth"
)

const (
	gqlDocContentBySlug = `query ContentBySlug($slug: String!) {
  contentBySlug(slug: $slug) {
    id
    content
    name
    slug
    created_at
    updated_at
    created_by {
      id
      email
      first_name
      last_name
    }
    updated_by {
      id
      email
      first_name
      last_name
    }
  }
}`

	gqlDocListPhenotypeAssays = `query ListPhenotypeAssays($search: String!) {
  listPhenotypeAssays(search: $search)
}`

	gqlDocListPhenotypeEnvironments = `query ListPhenotypeEnvironments($search: String!) {
  listPhenotypeEnvironments(search: $search)
}`

	gqlDocListPhenotypes = `query ListPhenotypes($search: String!) {
  listPhenotypes(search: $search)
}`

	gqlDocListStockSuggestions = `query ListStockSuggestions($query: String!, $limit: Int, $entity: StockEntityType) {
  listStockSuggestions(query: $query, limit: $limit, entity: $entity) {
    suggestions {
      id
      entity
      field
      display_text
      score
    }
    limit
    totalCount
  }
}`

	gqlDocListStrainsWithPhenotype = `query ListStrainsWithPhenotype(
  $cursor: Int!
  $limit: Int!
  $type: String!
  $annotation: String!
) {
  listStrainsWithAnnotation(
    cursor: $cursor
    limit: $limit
    type: $type
    annotation: $annotation
  ) {
    totalCount
    nextCursor
    strains {
      id
      label
      genes {
        name
      }
      publications {
        id
        pub_date
        title
        journal
        volume
        pages
        authors {
          last_name
        }
      }
    }
  }
}`

	gqlDocPlasmid = `query Plasmid($id: ID!) {
  plasmid(id: $id) {
    id
    name
    summary
    depositor {
      first_name
      last_name
    }
    publications {
      id
      pub_date
      title
      journal
      volume
      pages
      doi
      authors {
        last_name
      }
    }
    dbxrefs
    genes {
      name
    }
    image_map
    sequence
    keywords
    genbank_accession
    in_stock
  }
}`

	gqlDocPlasmidListFilter = `query PlasmidListFilter($cursor: Int!, $limit: Int!, $filter: PlasmidListFilter) {
  listPlasmids(cursor: $cursor, limit: $limit, filter: $filter) {
    nextCursor
    totalCount
    plasmids {
      id
      name
      summary
      in_stock
    }
  }
}`

	gqlDocPublication = `query Publication($id: ID!) {
  publication(id: $id) {
    id
    doi
    title
    abstract
    journal
    pub_date
    pages
    issue
    volume
    authors {
      initials
      last_name
    }
  }
}`

	gqlDocSearchStocks = `query SearchStocks($query: String!, $limit: Int, $entity: StockEntityType) {
  searchStocks(query: $query, limit: $limit, entity: $entity) {
    results {
      id
      entity
      field
      display_text
      score
      summary
      stock_label
      in_stock
      stock_type
    }
    limit
    totalCount
  }
}`

	gqlDocStrain = `query Strain($id: ID!) {
  strain(id: $id) {
    id
    label
    summary
    species
    parent {
      id
      label
    }
    depositor {
      first_name
      last_name
    }
    plasmid
    dbxrefs
    publications {
      id
      pub_date
      title
      journal
      volume
      pages
      doi
      authors {
        last_name
      }
    }
    genes {
      name
    }
    in_stock
    systematic_name
    genotypes
    mutagenesis_method
    genetic_modification
    names
    characteristics
    phenotypes {
      phenotype
      note
      assay
      environment
      publication {
        id
        pub_date
        title
        journal
        volume
        pages
        authors {
          last_name
        }
      }
    }
  }
}`

	gqlDocStrainList = `query StrainList($cursor: Int!, $limit: Int!, $filter: StrainListFilter) {
  listStrains(cursor: $cursor, limit: $limit, filter: $filter) {
    nextCursor
    totalCount
    strains {
      id
      label
      summary
      in_stock
    }
  }
}`
)

// LivenessOperationSpec defines one audited GraphQL read operation and its
// associated route area.
type LivenessOperationSpec struct {
	Operation string
	RouteArea string
}

// LivenessProbe describes a single GraphQL read query and its input variables.
type LivenessProbe struct {
	Operation string
	RouteArea string
	RootField string
	Label     string
	Query     string
	Variables map[string]any
}

// LivenessProbeOutcome holds the result of executing one LivenessProbe.
type LivenessProbeOutcome struct {
	Probe    LivenessProbe
	Passed   bool
	Latency  time.Duration
	ErrorMsg string
}

// OperationHealth summarizes probe outcomes for one GraphQL operation.
type OperationHealth struct {
	Operation string
	RouteArea string
	Total     int
	Passed    int
	Failed    int
	Latency   time.Duration
	Status    string
}

// LivenessReport aggregates all operation health rows and probe failures for
// Markdown rendering.
type LivenessReport struct {
	Endpoint   string
	Status     string
	Total      int
	Passed     int
	Failed     int
	Duration   time.Duration
	Operations []OperationHealth
	Failures   []LivenessProbeOutcome
}

func defaultLivenessOperationSpecs() []LivenessOperationSpec {
	return []LivenessOperationSpec{
		{Operation: opStrainList, RouteArea: routeCatalogAndStrains},
		{Operation: opPlasmidListFilter, RouteArea: routeCatalogAndPlasmids},
		{Operation: opSearchStocks, RouteArea: routeCatalogSearch},
		{Operation: opListStockSuggestions, RouteArea: routeNavbarSearch},
		{Operation: opPublication, RouteArea: routeReferencePreview},
		{Operation: opContentBySlug, RouteArea: routeInfoAndCommunity},
		{Operation: opStrain, RouteArea: routeStrainDetail},
		{Operation: opPlasmid, RouteArea: routePlasmidDetail},
		{Operation: opListStrainsWithPhenotype, RouteArea: routePhenotypeList},
		{Operation: opListPhenotypes, RouteArea: routeEditablePhenotype},
		{Operation: opListPhenotypeEnvironments, RouteArea: routeEditablePhenotype},
		{Operation: opListPhenotypeAssays, RouteArea: routeEditablePhenotype},
	}
}

func catalogLivenessProbes() []LivenessProbe {
	return []LivenessProbe{
		newStrainListProbe("strain_type=ALL", StrainTypeAll),
		newStrainListProbe("strain_type=REGULAR", StrainTypeRegular),
		newStrainListProbe("strain_type=GWDI", StrainTypeGwdi),
		newStrainListProbe("strain_type=BACTERIAL", StrainTypeBacterial),
		newPlasmidListProbe("plasmid_type=ALL", PlasmidTypeAll),
		newPlasmidListProbe("plasmid_type=REGULAR", PlasmidTypeRegular),
		newPlasmidListProbe("plasmid_type=GOLDEN_BRAID", PlasmidTypeGoldenBraid),
	}
}

func searchAndPublicationLivenessProbes() []LivenessProbe {
	return []LivenessProbe{
		{
			Operation: opSearchStocks,
			RouteArea: routeCatalogSearch,
			RootField: rootFieldSearchStocks,
			Label:     `query="act"`,
			Query:     gqlDocSearchStocks,
			Variables: map[string]any{
				flagQuery: sampleQueryAct,
				flagLimit: livenessSearchLimit,
			},
		},
		{
			Operation: opSearchStocks,
			RouteArea: routeCatalogSearch,
			RootField: rootFieldSearchStocks,
			Label:     `query="act", entity=STRAIN`,
			Query:     gqlDocSearchStocks,
			Variables: map[string]any{
				flagQuery:  sampleQueryAct,
				flagLimit:  livenessSearchLimit,
				flagEntity: string(StockEntityStrain),
			},
		},
		{
			Operation: opSearchStocks,
			RouteArea: routeCatalogSearch,
			RootField: rootFieldSearchStocks,
			Label:     `query="act", entity=PLASMID`,
			Query:     gqlDocSearchStocks,
			Variables: map[string]any{
				flagQuery:  sampleQueryAct,
				flagLimit:  livenessSearchLimit,
				flagEntity: string(StockEntityPlasmid),
			},
		},
		newSuggestionProbe(sampleQueryAct),
		newSuggestionProbe("gfp"),
		{
			Operation: opPublication,
			RouteArea: routeReferencePreview,
			RootField: "publication",
			Label:     `id="12857872"`,
			Query:     gqlDocPublication,
			Variables: map[string]any{"id": "12857872"},
		},
	}
}

func contentAndDetailLivenessProbes() []LivenessProbe {
	return []LivenessProbe{
		newContentSlugProbe("stockcenter-order"),
		newContentSlugProbe("stockcenter-deposit"),
		newContentSlugProbe("stockcenter-payment"),
		newContentSlugProbe("stockcenter-faq"),
		newContentSlugProbe("frontpage-gallery"),
		newContentSlugProbe("frontpage-show"),
		newContentSlugProbe("frontpage-DSC_SOP_OIM_Aug22.pdf"),
		newStrainDetailProbe("DBS0350026"),
		newStrainDetailProbe("DBS0305923"),
		newPlasmidDetailProbe("DBP0391806"),
		newPlasmidDetailProbe("DBP0001102"),
	}
}

func phenotypeLivenessProbes() []LivenessProbe {
	return []LivenessProbe{
		newPhenotypeStrainsProbe("wild type"),
		newPhenotypeStrainsProbe("aberrant aggregation"),
		newPhenotypeStrainsProbe("decreased growth rate"),
		newPhenotypeAutocompleteProbe(
			opListPhenotypes,
			"listPhenotypes",
			gqlDocListPhenotypes,
			sampleSearchGrowth,
		),
		newPhenotypeAutocompleteProbe(
			opListPhenotypeEnvironments,
			"listPhenotypeEnvironments",
			gqlDocListPhenotypeEnvironments,
			"agar",
		),
		newPhenotypeAutocompleteProbe(
			opListPhenotypeAssays,
			"listPhenotypeAssays",
			gqlDocListPhenotypeAssays,
			sampleSearchGrowth,
		),
	}
}

func newPhenotypeAutocompleteProbe(
	op, rootField, queryDoc, search string,
) LivenessProbe {
	return LivenessProbe{
		Operation: op,
		RouteArea: routeEditablePhenotype,
		RootField: rootField,
		Label:     `search="` + search + `"`,
		Query:     queryDoc,
		Variables: map[string]any{"search": search},
	}
}

func newStrainListProbe(label string, strainType StrainType) LivenessProbe {
	return LivenessProbe{
		Operation: opStrainList,
		RouteArea: routeCatalogAndStrains,
		RootField: "listStrains",
		Label:     label,
		Query:     gqlDocStrainList,
		Variables: map[string]any{
			gqlVarCursor: 0,
			flagLimit:    livenessCatalogLimit,
			gqlVarFilter: map[string]any{"strain_type": string(strainType)},
		},
	}
}

func newPlasmidListProbe(label string, plasmidType PlasmidType) LivenessProbe {
	return LivenessProbe{
		Operation: opPlasmidListFilter,
		RouteArea: routeCatalogAndPlasmids,
		RootField: "listPlasmids",
		Label:     label,
		Query:     gqlDocPlasmidListFilter,
		Variables: map[string]any{
			gqlVarCursor: 0,
			flagLimit:    livenessCatalogLimit,
			gqlVarFilter: map[string]any{"plasmid_type": string(plasmidType)},
		},
	}
}

func newSuggestionProbe(query string) LivenessProbe {
	return LivenessProbe{
		Operation: opListStockSuggestions,
		RouteArea: routeNavbarSearch,
		RootField: "listStockSuggestions",
		Label:     `query="` + query + `"`,
		Query:     gqlDocListStockSuggestions,
		Variables: map[string]any{
			"query": query,
			"limit": livenessCatalogLimit,
		},
	}
}

func newContentSlugProbe(slug string) LivenessProbe {
	return LivenessProbe{
		Operation: opContentBySlug,
		RouteArea: routeInfoAndCommunity,
		RootField: "contentBySlug",
		Label:     `slug="` + slug + `"`,
		Query:     gqlDocContentBySlug,
		Variables: map[string]any{"slug": slug},
	}
}

func newStrainDetailProbe(id string) LivenessProbe {
	return LivenessProbe{
		Operation: opStrain,
		RouteArea: routeStrainDetail,
		RootField: "strain",
		Label:     `id="` + id + `"`,
		Query:     gqlDocStrain,
		Variables: map[string]any{"id": id},
	}
}

func newPlasmidDetailProbe(id string) LivenessProbe {
	return LivenessProbe{
		Operation: opPlasmid,
		RouteArea: routePlasmidDetail,
		RootField: "plasmid",
		Label:     `id="` + id + `"`,
		Query:     gqlDocPlasmid,
		Variables: map[string]any{"id": id},
	}
}

func newPhenotypeStrainsProbe(annotation string) LivenessProbe {
	return LivenessProbe{
		Operation: opListStrainsWithPhenotype,
		RouteArea: routePhenotypeList,
		RootField: "listStrainsWithAnnotation",
		Label:     `annotation="` + annotation + `"`,
		Query:     gqlDocListStrainsWithPhenotype,
		Variables: map[string]any{
			gqlVarCursor: 0,
			flagLimit:    livenessPhenotypeLimit,
			"type":       "phenotype",
			"annotation": annotation,
		},
	}
}
