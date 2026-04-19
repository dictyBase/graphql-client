# Refined Project Prompt: Plasmid GraphQL CLI

Create a Go-based CLI application using `urfave/cli/v3` and `github.com/hasura/go-graphql-client` that lists plasmids from a GraphQL endpoint.

## Core Requirements
- **CLI Framework:** Use `urfave/cli/v3`.
- **GraphQL Client:** Use `github.com/hasura/go-graphql-client` (modeled after the `pubclient` example).
- **Subcommand:** Implement a `list-plasmids` subcommand.
- **Flags:**
  - `--endpoint`: The GraphQL API URL (default: `https://graphql.dictybase.dev/graphql`).
  - `--limit` (int): Number of entries to fetch (maps to `limit` argument in query).

## GraphQL Integration
- **Endpoint:** Use the `listPlasmids` query (filtering as specified by `PlasmidListFilter`).
- **Schema Reference:** [dicty-graphql-schema](https://github.com/dictyBase/dicty-graphql-schema/).
- **Required Fields for Output:**
  - `id`
  - `name`
  - `summary`
  - `in_stock`

## Implementation Guidance
- Use the `pubclient/main.go` file as a structural template for the GraphQL client initialization and query execution.
- Ensure proper error handling for network requests and GraphQL errors.
- Output the plasmid list in a clear, readable format to the console.
