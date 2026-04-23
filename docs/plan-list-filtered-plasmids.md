# Plan: `list-filtered-plasmids` Subcommand

## Overview

Add a new `list-filtered-plasmids` subcommand that filters plasmids by **any combination** of attribute filters (name, summary) **in addition to** the existing `plasmid_type` (ALL/REGULAR/GOLDEN_BRAID) category filter. This follows the identical architectural pattern as the existing `list-plasmids` subcommand.

## GraphQL Schema Context

The `PlasmidListFilter` input type in the GraphQL schema (`src/schema/stock.graphql`) is:

```graphql
input PlasmidListFilter {
  name: String
  summary: String
  id: ID
  in_stock: Boolean
  plasmid_type: PlasmidType!
}
```

The `listPlasmids` query accepts this filter:

```graphql
listPlasmids(cursor: Int, limit: Int, filter: PlasmidListFilter): PlasmidListWithCursor
```

The existing `list-plasmids` subcommand only populates the `plasmid_type` field of `PlasmidListFilter`. The new subcommand will allow populating the `name` and `summary` fields of `PlasmidListFilter` (the `id` and `in_stock` fields are not exposed as CLI flags).

---

## New CLI Flags

The `list-filtered-plasmids` subcommand will have these flags:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--limit` | int | `10` | Number of plasmid entries to fetch |
| `--type` | string | `ALL` | Plasmid type filter (ALL, REGULAR, GOLDEN_BRAID) |
| `--name` | string | `""` (empty, not sent) | Filter by plasmid name (substring match) |
| `--summary` | string | `""` (empty, not sent) | Filter by plasmid summary (substring match) |

**Design decision for string flags:** Empty string (`""`) means "do not apply this filter". When a non-empty value is provided, it is sent as the filter value. Using pointer types (`*string`) for optional string filters ensures that empty strings are never sent as filter values — nil pointers are omitted from the GraphQL variables.

---

## File-by-File Implementation Plan

### 1. `model.go` — Add new types

Add a new `PlasmidAttributeFilter` type with pointer-based attribute fields and a new query struct for the filtered variant.

**Why pointer types?** Using `*string` with `omitempty` ensures:
- Unset flags → `nil` → field omitted from the GraphQL variables
- Set flags → non-nil pointer → field included in the GraphQL variables

```go
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

// Existing filter - used by list-plasmids subcommand
type PlasmidListFilter struct {
	PlasmidType PlasmidType `json:"plasmid_type"`
}

func (PlasmidListFilter) GetGraphQLType() string { return "PlasmidListFilter" }

// New filter - used by list-filtered-plasmids subcommand
// Pointer fields ensure nil values are omitted from GraphQL variables
// Only name and summary are exposed as CLI flags
// (id and in_stock are not used by this subcommand)
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

// Existing query struct - used by list-plasmids
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

// New query struct - used by list-filtered-plasmids
// Reuses the same GraphQL query but with the richer filter type
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
```

### 2. `errors.go` — No changes needed

The existing `errInvalidPlasmidType` is sufficient. No new error types are needed since `id` and `in-stock` flags are not being added.

### 3. `cmd.go` — Add new subcommand pipeline

Add new type aliases and the full fp-go pipeline for the filtered subcommand. The existing `list-plasmids` code remains untouched.

```go
package main

import (
	"context"
	"fmt"
	"os"

	R "github.com/IBM/fp-go/v2/record"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	T "github.com/IBM/fp-go/v2/tuple"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/urfave/cli/v3"
)

// ---- existing list-plasmids types (unchanged) ----

type (
	listPlasmidCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	listPlasmidValidatedArgs = T.Tuple3[context.Context, *cli.Command, PlasmidType]
	listPlasmidFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, PlasmidListFilter]
)

// ---- new list-filtered-plasmids types ----

type (
	filteredPlasmidCLIArgs       = T.Tuple2[context.Context, *cli.Command]
	filteredPlasmidValidatedArgs = T.Tuple3[context.Context, *cli.Command, PlasmidType]
	filteredPlasmidFetchArgs     = T.Tuple4[context.Context, *graphql.Client, int, PlasmidAttributeFilter]
)

// ============================================================
// existing list-plasmids functions (unchanged)
// ============================================================

func toListPlasmidValidatedArgs(
	input listPlasmidCLIArgs,
) IOE.IOEither[error, listPlasmidValidatedArgs] {
	return F.Pipe2(
		ParsePlasmidType(input.F2.String("type")),
		IOE.FromEither[error, PlasmidType],
		IOE.Map[error](func(plasmidType PlasmidType) listPlasmidValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, plasmidType)
		}),
	)
}

func mapListPlasmidFetchArgs(input listPlasmidValidatedArgs) listPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String("endpoint"), nil),
		input.F2.Int("limit"),
		PlasmidListFilter{PlasmidType: input.F3},
	)
}

func fetchListPlasmidsIO(input listPlasmidFetchArgs) IOE.IOEither[error, ListPlasmidsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListPlasmidsQuery, error) {
			query := new(ListPlasmidsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				"cursor": 0,
				"limit":  input.F3,
				"filter": input.F4,
			})
		}),
		IOE.MapLeft[*ListPlasmidsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListPlasmidsQuery).toResult),
	)
}

func RunListPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, listPlasmidCLIArgs],
		IOE.Chain(toListPlasmidValidatedArgs),
		IOE.Map[error](mapListPlasmidFetchArgs),
		IOE.Chain(fetchListPlasmidsIO),
		toEither[error, ListPlasmidsResult],
		E.Fold(
			func(err error) error { return err },
			func(result ListPlasmidsResult) error {
				writePlasmidTable(os.Stdout, result.Plasmids)
				writeSummary(os.Stdout, result)
				return nil
			},
		),
	)
}

// ============================================================
// new list-filtered-plasmids functions
// ============================================================

// ParsePlasmidType validates and converts a string to PlasmidType
func ParsePlasmidType(s string) E.Either[error, PlasmidType] {
	return F.Pipe2(
		plasmidTypeMap,
		R.Lookup[PlasmidType](s),
		E.FromOption[PlasmidType](func() error {
			return errInvalidPlasmidType(PlasmidType(s))
		}),
	)
}

// stringPtr returns a *string if the input is non-empty, nil otherwise.
// This ensures empty-string flags are not sent as filter values.
// Uses fp-go Option: FromPredicate + Fold instead of imperative if/else.
func stringPtr(s string) *string {
	return F.Pipe2(
		s,
		O.FromPredicate(P.IsNonZero[string]()),
		O.Fold(
			func() *string { return nil },
			func(s string) *string { return &s },
		),
	)
}

// buildPlasmidAttributeFilter constructs the filter from CLI flags.
// Only non-nil fields will be included in the GraphQL variables.
func buildPlasmidAttributeFilter(
	plasmidType PlasmidType,
	cmd *cli.Command,
) PlasmidAttributeFilter {
	return PlasmidAttributeFilter{
		PlasmidType: plasmidType,
		Name:        stringPtr(cmd.String("name")),
		Summary:     stringPtr(cmd.String("summary")),
	}
}

func toFilteredPlasmidValidatedArgs(
	input filteredPlasmidCLIArgs,
) IOE.IOEither[error, filteredPlasmidValidatedArgs] {
	return F.Pipe2(
		ParsePlasmidType(input.F2.String("type")),
		IOE.FromEither[error, PlasmidType],
		IOE.Map[error](func(plasmidType PlasmidType) filteredPlasmidValidatedArgs {
			return T.MakeTuple3(input.F1, input.F2, plasmidType)
		}),
	)
}

func mapFilteredPlasmidFetchArgs(input filteredPlasmidValidatedArgs) filteredPlasmidFetchArgs {
	return T.MakeTuple4(
		input.F1,
		graphql.NewClient(input.F2.String("endpoint"), nil),
		input.F2.Int("limit"),
		buildPlasmidAttributeFilter(input.F3, input.F2),
	)
}

func fetchListFilteredPlasmidsIO(
	input filteredPlasmidFetchArgs,
) IOE.IOEither[error, ListPlasmidsResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*ListFilteredPlasmidsQuery, error) {
			query := new(ListFilteredPlasmidsQuery)
			return query, input.F2.Query(input.F1, query, map[string]any{
				"cursor": 0,
				"limit":  input.F3,
				"filter": input.F4,
			})
		}),
		IOE.MapLeft[*ListFilteredPlasmidsQuery](func(err error) error {
			return fmt.Errorf("graphql query failed: %w", err)
		}),
		IOE.Map[error]((*ListFilteredPlasmidsQuery).toResult),
	)
}

func RunListFilteredPlasmidCLI(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		T.MakeTuple2(ctx, cmd),
		IOE.Of[error, filteredPlasmidCLIArgs],
		IOE.Chain(toFilteredPlasmidValidatedArgs),
		IOE.Map[error](mapFilteredPlasmidFetchArgs),
		IOE.Chain(fetchListFilteredPlasmidsIO),
		toEither[error, ListPlasmidsResult],
		E.Fold(
			func(err error) error { return err },
			func(result ListPlasmidsResult) error {
				writePlasmidTable(os.Stdout, result.Plasmids)
				writeSummary(os.Stdout, result)
				return nil
			},
		),
	)
}
```

### 4. `main.go` — Register the new subcommand

Add the `list-filtered-plasmids` subcommand alongside the existing `list-plasmids`.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const defaultLimit = 10

const defaultEndpoint = "https://graphql.dictybase.dev/graphql"

func main() {
	cmd := &cli.Command{
		Name:  "gql-plasmid",
		Usage: "CLI to query plasmid data from a GraphQL endpoint",
		Commands: []*cli.Command{
			{
				Name:   "list-plasmids",
				Usage:  "List plasmids from the GraphQL endpoint",
				Action: RunListPlasmidCLI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "endpoint",
						Value: defaultEndpoint,
						Usage: "GraphQL API endpoint URL",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: defaultLimit,
						Usage: "Number of plasmid entries to fetch",
					},
					&cli.StringFlag{
						Name:  "type",
						Value: string(PlasmidTypeAll),
						Usage: "Plasmid type to filter (ALL, REGULAR, GOLDEN_BRAID)",
					},
				},
			},
			{
				Name:   "list-filtered-plasmids",
				Usage:  "List plasmids with attribute-level filtering from the GraphQL endpoint",
				Action: RunListFilteredPlasmidCLI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "endpoint",
						Value: defaultEndpoint,
						Usage: "GraphQL API endpoint URL",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: defaultLimit,
						Usage: "Number of plasmid entries to fetch",
					},
					&cli.StringFlag{
						Name:  "type",
						Value: string(PlasmidTypeAll),
						Usage: "Plasmid type to filter (ALL, REGULAR, GOLDEN_BRAID)",
					},
					&cli.StringFlag{
						Name:  "name",
						Usage: "Filter by plasmid name",
					},
					&cli.StringFlag{
						Name:  "summary",
						Usage: "Filter by plasmid summary",
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

### 5. `display.go` — No changes needed

The `writePlasmidTable` and `writeSummary` functions work on `[]Plasmid` and `ListPlasmidsResult` respectively, which are the same types returned by both subcommands. No changes required.

### 6. `parse_fp_test.go` — Add tests for new helpers

Add tests for `stringPtr` and `buildPlasmidAttributeFilter` alongside the existing `ParsePlasmidType` tests.

```go
package main

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// ---- existing tests (unchanged) ----

func TestParsePlasmidTypeFP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PlasmidType
		wantErr bool
	}{
		{
			name:  "ALL type",
			input: "ALL",
			want:  PlasmidTypeAll,
		},
		{
			name:  "REGULAR type",
			input: "REGULAR",
			want:  PlasmidTypeRegular,
		},
		{
			name:  "GOLDEN_BRAID type",
			input: "GOLDEN_BRAID",
			want:  PlasmidTypeGoldenBraid,
		},
		{
			name:    "empty string is invalid",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "INVALID",
			wantErr: true,
		},
		{
			name:    "lowercase is invalid",
			input:   "all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePlasmidType(tt.input)
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

// ---- new tests ----

func TestStringPtr(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		result := stringPtr("")
		require.Nil(t, result)
	})

	t.Run("non-empty string returns pointer", func(t *testing.T) {
		result := stringPtr("pDM123")
		require.NotNil(t, result)
		require.Equal(t, "pDM123", *result)
	})
}

func TestBuildPlasmidAttributeFilter(t *testing.T) {
	t.Run("only plasmid type, no attribute filters", func(t *testing.T) {
		cmd := buildTestCommand(t, nil)
		filter := buildPlasmidAttributeFilter(PlasmidTypeAll, cmd)
		require.Equal(t, PlasmidTypeAll, filter.PlasmidType)
		require.Nil(t, filter.Name)
		require.Nil(t, filter.Summary)
	})

	t.Run("all attribute filters set", func(t *testing.T) {
		cmd := buildTestCommand(t, map[string]string{
			"name":    "pDM123",
			"summary": "expression vector",
		})
		filter := buildPlasmidAttributeFilter(PlasmidTypeRegular, cmd)
		require.Equal(t, PlasmidTypeRegular, filter.PlasmidType)
		require.NotNil(t, filter.Name)
		require.Equal(t, "pDM123", *filter.Name)
		require.NotNil(t, filter.Summary)
		require.Equal(t, "expression vector", *filter.Summary)
	})

	t.Run("partial attribute filters", func(t *testing.T) {
		cmd := buildTestCommand(t, map[string]string{
			"name": "pDM",
		})
		filter := buildPlasmidAttributeFilter(PlasmidTypeGoldenBraid, cmd)
		require.Equal(t, PlasmidTypeGoldenBraid, filter.PlasmidType)
		require.NotNil(t, filter.Name)
		require.Equal(t, "pDM", *filter.Name)
		require.Nil(t, filter.Summary)
	})
}

// buildTestCommand creates a *cli.Command with flags set from the given map.
// This is needed because cli.Command flags must be parsed through the CLI
// framework for IsSet() to work correctly.
func buildTestCommand(t *testing.T, flags map[string]string) *cli.Command {
	t.Helper()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
			&cli.StringFlag{Name: "summary"},
		},
	}
	for k, v := range flags {
		switch k {
		case "name":
			cmd.Set("name", v)
		case "summary":
			cmd.Set("summary", v)
		}
	}
	return cmd
}
```

---

## Architectural Alignment

The new subcommand mirrors the existing `list-plasmids` exactly:

| Aspect | `list-plasmids` | `list-filtered-plasmids` |
|--------|----------------|--------------------------|
| CLI args type | `listPlasmidCLIArgs` | `filteredPlasmidCLIArgs` |
| Validated args type | `listPlasmidValidatedArgs` | `filteredPlasmidValidatedArgs` |
| Fetch args type | `listPlasmidFetchArgs` | `filteredPlasmidFetchArgs` |
| Filter struct | `PlasmidListFilter` | `PlasmidAttributeFilter` |
| Query struct | `ListPlasmidsQuery` | `ListFilteredPlasmidsQuery` |
| Validation fn | `toListPlasmidValidatedArgs` | `toFilteredPlasmidValidatedArgs` |
| Mapping fn | `mapListPlasmidFetchArgs` | `mapFilteredPlasmidFetchArgs` |
| Fetch fn | `fetchListPlasmidsIO` | `fetchListFilteredPlasmidsIO` |
| Run fn | `RunListPlasmidCLI` | `RunListFilteredPlasmidCLI` |
| Pipeline depth | `Pipe6` | `Pipe6` |
| Display | `writePlasmidTable` + `writeSummary` | `writePlasmidTable` + `writeSummary` |

Both subcommands use the same fp-go pipeline pattern:
1. Wrap CLI args in `IOE.Of`
2. Validate plasmid type → `IOE.Chain(toValidatedArgs)`
3. Map to fetch args → `IOE.Map(mapFetchArgs)`
4. Execute GraphQL query → `IOE.Chain(fetchIO)`
5. Unwrap IO → `toEither`
6. Fold error/success

---

## Usage Examples

```bash
# List all plasmids with default filters (same as list-plasmids --type ALL)
gql-plasmid list-filtered-plasmids

# List golden braid plasmids with name containing "pDM"
gql-plasmid list-filtered-plasmids --type GOLDEN_BRAID --name pDM

# List regular plasmids with a specific summary substring
gql-plasmid list-filtered-plasmids --type REGULAR --summary "expression"

# Combine name and summary filters
gql-plasmid list-filtered-plasmids --type REGULAR --name pDM --summary vector --limit 20
```

---

## Files Changed Summary

| File | Action | Description |
|------|--------|-------------|
| `model.go` | Modify | Add `PlasmidAttributeFilter`, `ListFilteredPlasmidsQuery` |
| `errors.go` | No change | Existing `errInvalidPlasmidType` is sufficient |
| `cmd.go` | Modify | Add `filteredPlasmid*` types, `stringPtr`, `buildPlasmidAttributeFilter`, `toFilteredPlasmidValidatedArgs`, `mapFilteredPlasmidFetchArgs`, `fetchListFilteredPlasmidsIO`, `RunListFilteredPlasmidCLI` |
| `main.go` | Modify | Add `list-filtered-plasmids` subcommand with additional flags |
| `display.go` | No change | Reuses existing `writePlasmidTable` + `writeSummary` |
| `parse_fp_test.go` | Modify | Add `TestStringPtr`, `TestBuildPlasmidAttributeFilter`, `buildTestCommand` helper |

---

## Edge Cases Handled

1. **Empty string flags** → `stringPtr("")` returns `nil` → field omitted from GraphQL filter
2. **Invalid plasmid type** → `ParsePlasmidType` returns `Left(errInvalidPlasmidType)` → pipeline short-circuits → error printed
3. **No attribute filters** → all pointer fields are `nil` → behaves identically to `list-plasmids` for the same type value
4. **GraphQL errors** → `IOE.TryCatchError` catches → `MapLeft` wraps with context → error printed
