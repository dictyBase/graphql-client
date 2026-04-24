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
