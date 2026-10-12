# gql-stocks

A CLI for querying plasmid and strain stock data from a dictyBase GraphQL endpoint.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
  - [list-plasmids](#list-plasmids)
  - [list-filtered-plasmids](#list-filtered-plasmids)
  - [list-strains](#list-strains)
  - [list-filtered-strains](#list-filtered-strains)
  - [search-stocks](#search-stocks)
  - [create-order](#create-order)
- [Output](#output)
- [Development](#development)
- [Architecture](#architecture)

## Installation

### Clone and build

Requires [Go 1.25+](https://go.dev/dl/).

```bash
git clone https://github.com/dictybase-docker/graphql-client.git
cd graphql-client
go build -o gql-stocks .
```

### Install with `go install`

```bash
go install github.com/dictybase-docker/graphql-client@latest
```

This places the `graphql-client` binary in your `$GOPATH/bin`.

## Quick Start

```bash
# List the first 10 plasmids
./gql-stocks list-plasmids

# List the first 10 strains
./gql-stocks list-strains

# Filter plasmids by name
./gql-stocks list-filtered-plasmids --name pDM

# Filter strains by label
./gql-stocks list-filtered-strains --label DBS0352420

# Search stocks by a text query
./gql-stocks search-stocks --query pDM

# Create a stock order
./gql-stocks create-order --consumer jane@example.org --payer jane@example.org --items DBS0352420
```

## Commands

### `list-plasmids`

List plasmids from the GraphQL endpoint, filtered by type.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--limit` | int | `10` | Number of entries to fetch |
| `--type` | string | `ALL` | Plasmid type: `ALL`, `REGULAR`, `GOLDEN_BRAID` |

```bash
./gql-stocks list-plasmids --type REGULAR --limit 25
```

### `list-filtered-plasmids`

List plasmids with attribute-level filtering (name, summary) in addition to type.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--limit` | int | `10` | Number of entries to fetch |
| `--type` | string | `ALL` | Plasmid type: `ALL`, `REGULAR`, `GOLDEN_BRAID` |
| `--name` | string | — | Filter by plasmid name (substring match) |
| `--summary` | string | — | Filter by plasmid summary (substring match) |

```bash
./gql-stocks list-filtered-plasmids --type GOLDEN_BRAID --name pDM
./gql-stocks list-filtered-plasmids --type REGULAR --summary "expression vector" --limit 20
```

Omitted string flags are not sent to the API — only non-empty values are included in the GraphQL filter.

### `list-strains`

List strains from the GraphQL endpoint, filtered by type.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--limit` | int | `10` | Number of entries to fetch |
| `--type` | string | `ALL` | Strain type: `ALL`, `REGULAR`, `GWDI`, `BACTERIAL` |

```bash
./gql-stocks list-strains --type GWDI --limit 15
```

### `list-filtered-strains`

List strains with attribute-level filtering (label, summary) in addition to type.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--limit` | int | `10` | Number of entries to fetch |
| `--type` | string | `ALL` | Strain type: `ALL`, `REGULAR`, `GWDI`, `BACTERIAL` |
| `--label` | string | — | Filter by strain label (substring match) |
| `--summary` | string | — | Filter by strain summary (substring match) |

```bash
./gql-stocks list-filtered-strains --type BACTERIAL --label DBS
./gql-stocks list-filtered-strains --type REGULAR --summary "axenic" --limit 20
```

### `search-stocks`

Search strains and plasmids with one text query. The server searches many stock
fields at the same time and sorts the results by relevance score. Use this
command when you know a fragment of a name, a label, a gene, or a summary, but
you do not know the stock accession.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--query` | string | — | Search query, at least 2 characters (required) |
| `--limit` | int | `10` | Number of search results to fetch. The server accepts `0` to `100` |
| `--entity` | string | `ALL` | Stock entity to search: `ALL`, `STRAIN`, `PLASMID` |

```bash
./gql-stocks search-stocks --query pDM
./gql-stocks search-stocks --query axenic --entity STRAIN --limit 20
./gql-stocks search-stocks --query GFP --entity PLASMID --limit 4
```

#### How the query matches

The query is one string. The command does not accept per-field flags. The
server applies the string to every searchable field and returns the best field
per stock.

- The query must have 2 characters or more. A shorter query fails before any
  network call with `invalid stock search query: must be at least 2 characters`.
- The match ignores letter case. `pDM`, `PDM`, and `pdm` return the same stocks.
- A prefix match ranks first. `pDM` matches the plasmid name `pDM309` and also
  the name `pdmpA::bsr`.
- A fuzzy match ranks after every prefix match. The query `309` matches the
  strain label `V30309` and the plasmid name `pDM309` with a much lower score.
- The `FIELD` column reports the field that produced the match. Therefore one
  query can return a `NAME` hit, a `LABEL` hit, and a `SUMMARY` hit together.

The server schema defines 10 match fields. The table lists each value of the
`StockSearchFieldEnum` type.

| `FIELD` value | Meaning | Seen in live probes |
|---------------|---------|---------------------|
| `STOCK_ID` | Stock accession, such as `DBS0236853` | yes |
| `LABEL` | Strain label, such as `rasC-/AX2` | yes |
| `NAME` | Plasmid name, such as `pDM309` | yes |
| `SUMMARY` | Free-text stock summary | yes |
| `SPECIES` | Strain species, such as `Polysphondylium pallidum` | yes |
| `GENES` | Associated gene name, such as `mhcA` | yes |
| `NAMES` | Alternate strain names | not seen |
| `PLASMID` | Plasmid reference of a strain | not seen |
| `DBXREFS` | External database cross-references | not seen |
| `DEPOSITOR` | Depositor of the stock | not seen |

#### Entity filter

The `--entity` flag restricts the search to one stock type. The CLI validates
the value before the network call. An unknown value fails with
`invalid stock entity type: SPECIES, must be one of: ALL, STRAIN, PLASMID`.

| `--entity` | Result set | Use it when |
|------------|------------|-------------|
| `ALL` | Strains and plasmids in one ranked list | You do not know the stock type |
| `STRAIN` | Strains only, with the `LABEL` column filled | You want cell lines |
| `PLASMID` | Plasmids only, with an empty `LABEL` column | You want vectors |

```bash
# Both stock types, strongest matches first
./gql-stocks search-stocks --query AX4 --limit 8

# Strains only
./gql-stocks search-stocks --query rasC --entity STRAIN --limit 5

# Plasmids only
./gql-stocks search-stocks --query pDM --entity PLASMID --limit 5
```

#### Read the output

The command prints one row per match and a total line. The columns come from
the `StockSearchResult` type of the API.

| Column | Content |
|--------|---------|
| `ID` | Stock accession. `DBS` marks a strain and `DBP` marks a plasmid. Pass this value to `create-order --items` |
| `ENTITY` | `STRAIN` or `PLASMID` |
| `FIELD` | The field that matched the query |
| `DISPLAY TEXT` | The text of the matched field |
| `IN STOCK` | `true` when the stock center holds inventory for the stock |
| `SCORE` | Relevance score with 4 decimals. A higher score ranks higher |
| `LABEL` | Strain label. The column stays empty for plasmids |
| `SUMMARY` | Stock summary. The text can repeat `DISPLAY TEXT` for a `SUMMARY` match |

The score is relative, not a percentage. A prefix match returns the top score
`1000.0000`. A fuzzy match returns a lower value, and the exact number depends
on the field and the corpus. Compare scores inside one result set only.

The `Total:` line reports the number of rows in the response, not the number of
stocks in the database. The search API has no cursor. To see more matches,
raise `--limit` up to `100`.

Strain search, with a label prefix match for every row:

```bash
./gql-stocks search-stocks --query rasC --entity STRAIN --limit 5
```

```
ID           ENTITY   FIELD   DISPLAY TEXT   IN STOCK   SCORE       LABEL         SUMMARY
--           ------   -----   ------------   --------   -----       -----         -------
DBS0236853   STRAIN   LABEL   rasC-/AX2      true       1000.0000   rasC-/AX2     rasC null in AX2
DBS0236854   STRAIN   LABEL   rasC-/JH10     true       1000.0000   rasC-/JH10    rasC null in JH10
DBS0236855   STRAIN   LABEL   rasC-/dmpA-    true       1000.0000   rasC-/dmpA-   rasC-/dmpA- double null mutant...
DBS0236856   STRAIN   LABEL   rasC-/pikD1    true       1000.0000   rasC-/pikD1   REMI of rasC-, disrupting the pikD gene...
DBS0236857   STRAIN   LABEL   rasC-/rasD-    false      1000.0000   rasC-/rasD-   rasC/rasD double null mutant

Total: 5
```

Mixed search, with prefix matches first and summary matches after them:

```bash
./gql-stocks search-stocks --query AX4 --limit 6
```

```
ID           ENTITY   FIELD     DISPLAY TEXT       IN STOCK   SCORE       LABEL             SUMMARY
--           ------   -----     ------------       --------   -----       -----             -------
DBS0237637   STRAIN   LABEL     AX4                true       1000.0000   AX4               Adam Kuspa lab strain; axenic...
DBS0307938   STRAIN   LABEL     AX4/[tagB]:lacZ    true       1000.0000   AX4/[tagB]:lacZ   lacZ taggedneomycin resistant
DBS0349730   STRAIN   LABEL     AX4/ura-           true       1000.0000   AX4/ura-          pyr56-null strain with Uracil auxotrophy...
DBS0351471   STRAIN   LABEL     AX4                true       1000.0000   AX4               AX4 strain from the Thompson lab...
DBS0236498   STRAIN   SUMMARY   lagD- in AX4       false      507.7521    lagD-             lagD- in AX4
DBS0235916   STRAIN   SUMMARY   dnmA null in AX4   false      507.5391    dnmA-             dnmA null in AX4

Total: 6
```

A query with no match prints the header and a message:

```
ID   ENTITY   FIELD   DISPLAY TEXT   IN STOCK   SCORE   LABEL   SUMMARY
--   ------   -----   ------------   --------   -----   -----   -------

No stocks found.
```

#### Errors

| Condition | Message | Checked by |
|-----------|---------|------------|
| Query shorter than 2 characters | `invalid stock search query: must be at least 2 characters` | CLI, before the request |
| Entity outside the enum | `invalid stock entity type: <value>, must be one of: ALL, STRAIN, PLASMID` | CLI, before the request |
| `--limit` above `100` or below `0` | `graphql query failed: ... validation error: - data.attributes.limit: value must be greater than or equal to 0 and less than or equal to 100` | Server |

The command exits with status 1 on an error and writes the message to standard
error. A `--limit` of `0` is valid and the server then applies its own default
of 50 results.

### `create-order`

Send a stock order to the GraphQL endpoint. Only `--consumer`, `--payer`, and `--items` need real values. The other flags have defaults that you can keep.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--endpoint` | string | `https://graphql.dictybase.dev/graphql` | GraphQL API endpoint URL |
| `--consumer` | string | — | Consumer email address, receives the invoice (required) |
| `--payer` | string | — | Payer email address, shown in the invoice (required) |
| `--items` | string | — | Stock item ids (DBS/DBP accessions). Repeat the flag or use commas (required) |
| `--purchaser` | string | `fake-purchaser@example.org` | Purchaser email address |
| `--courier` | string | `FedEx` | Courier name |
| `--courier-account` | string | `FAKE-ACCT-0001` | Courier account number |
| `--payment` | string | `credit_card` | Payment method |
| `--comments` | string | `automated CLI test order` | Order comments |
| `--purchase-order-num` | string | `PO-FAKE-0001` | Purchase order number |
| `--status` | string | `IN_PREPARATION` | Order status: `IN_PREPARATION`, `GROWING`, `CANCELLED`, `SHIPPED` |

```bash
./gql-stocks create-order \
  --consumer jane@example.org \
  --payer jane@example.org \
  --items DBS0352420 --items DBP1234
```

On success, the command prints the new order ID.

## Output

Results are printed as a tab-aligned table followed by a summary line.

**Plasmids:**

```
ID      NAME        IN STOCK  SUMMARY
--      ----        --------  -------
DBP123  pDM304      true      expression vector for D. discoideum

Total: 142 | Next cursor: 10
```

**Strains:**

```
ID        LABEL         IN STOCK  SUMMARY
--        -----         --------  -------
DBS03524  DBS0352420    true      axenic strain

Total: 87 | Next cursor: 10
```

**Stock search:**

```
ID           ENTITY    FIELD   DISPLAY TEXT   IN STOCK   SCORE       LABEL   SUMMARY
--           ------    -----   ------------   --------   -----       -----   -------
DBP0000511   PLASMID   NAME    pdmpA::bsr     true       1000.0000           Remi rescued plasmid...

Total: 5
```

**Orders:**

```
Order created: 17342e12-acd0-4f2a-9456-123e51c0
```

## Development

### Prerequisites

- Go 1.25+

### Test

```bash
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...
```

### Lint

```bash
golangci-lint run ./...
```

### Format

```bash
golangci-lint fmt
```

## Architecture

The CLI follows a functional pipeline pattern using [fp-go](https://github.com/IBM/fp-go). Each command flows through the same stages:

1. **Wrap** — CLI args wrapped in `IOEither`
2. **Validate** — Parse and validate type enums
3. **Map** — Build GraphQL client and filter from validated args
4. **Fetch** — Execute the GraphQL query
5. **Unwrap** — Convert `IOEither` to `Either`
6. **Fold** — Render table on success, print error on failure

### Key Files

| File | Purpose |
|------|---------|
| `main.go` | CLI command definitions and flag registration |
| `cmd.go` | Functional pipelines for each subcommand |
| `model.go` | GraphQL query structs, filter types, and result types |
| `display.go` | Tab-aligned table rendering |
| `errors.go` | Typed error constructors |
