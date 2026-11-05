package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	loomcheck "github.com/dictyBase/fp-go-loom/loomcheck"
	pipelinecheck "github.com/dictyBase/fp-go-loom/pipelinecheck"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func livenessTestCommand() *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEndpoint,
				Value: "https://liveness.unit.test/graphql",
			},
			&cli.DurationFlag{
				Name:  flagTimeout,
				Value: 5 * time.Second,
			},
		},
	}
}

func livenessTestArgs() livenessCLIArgs {
	return F.Pipe2(
		livenessCLIArgs{},
		cliArgsCtxLens.Set(context.Background()),
		cliArgsCmdLens.Set(livenessTestCommand()),
	)
}

func TestDefaultLivenessProbesCoverAllTwelveAuditOperations(t *testing.T) {
	specs := defaultLivenessOperationSpecs()
	require.Len(t, specs, 12)

	probes := defaultLivenessProbes()
	require.Len(t, probes, 30)

	seenOps := make(map[string]int)
	for _, probe := range probes {
		require.NotEmpty(t, probe.Operation)
		require.NotEmpty(t, probe.RouteArea)
		require.NotEmpty(t, probe.RootField)
		require.NotEmpty(t, probe.Label)
		require.NotEmpty(t, probe.Query)
		require.NotEmpty(t, probe.Variables)
		seenOps[probe.Operation]++
	}

	for _, spec := range specs {
		require.Positive(
			t,
			seenOps[spec.Operation],
			"expected at least one probe for operation %s",
			spec.Operation,
		)
	}
}

func TestValidateLivenessArgs(t *testing.T) {
	t.Run("valid flags pass", func(t *testing.T) {
		args := livenessTestArgs()
		val, err := E.Unwrap(validateLivenessArgs(args))
		require.NoError(t, err)
		require.Equal(t, args, val)
	})

	t.Run("empty endpoint fails", func(t *testing.T) {
		args := livenessTestArgs()
		require.NoError(t, args.Command.Set(flagEndpoint, ""))
		_, err := E.Unwrap(validateLivenessArgs(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "endpoint")
	})

	t.Run("non-positive timeout fails", func(t *testing.T) {
		args := livenessTestArgs()
		require.NoError(t, args.Command.Set(flagTimeout, "0s"))
		_, err := E.Unwrap(validateLivenessArgs(args))
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout")
	})
}

func TestValidateProbePayload(t *testing.T) {
	validate := F.Bind1st(validateProbePayload, "strain")

	t.Run("empty payload fails", func(t *testing.T) {
		_, err := E.Unwrap(validate(nil))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty graphql response payload")
	})

	t.Run("invalid JSON fails", func(t *testing.T) {
		_, err := E.Unwrap(validate([]byte("{not-json")))
		require.Error(t, err)
	})

	t.Run("null root field fails", func(t *testing.T) {
		_, err := E.Unwrap(validate([]byte(`{"strain":null}`)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty or null data for strain")
	})

	t.Run("missing root field fails", func(t *testing.T) {
		_, err := E.Unwrap(validate([]byte(`{"other":{}}`)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty or null data for strain")
	})

	t.Run("non-null root field passes", func(t *testing.T) {
		raw, err := E.Unwrap(validate([]byte(`{"strain":{"id":"DBS0350026"}}`)))
		require.NoError(t, err)
		require.Contains(t, raw, "DBS0350026")
	})
}

func TestConciseGraphQLError(t *testing.T) {
	t.Run("extracts first message from graphql.Errors", func(t *testing.T) {
		errs := graphql.Errors{
			{Message: "ListPhenotypes is not yet implemented"},
			{Message: "secondary error"},
		}
		require.Equal(
			t,
			"ListPhenotypes is not yet implemented",
			conciseGraphQLError(errs),
		)
	})

	t.Run("falls back to Error() for standard errors", func(t *testing.T) {
		err := errors.New("connection refused")
		require.Equal(t, "connection refused", conciseGraphQLError(err))
	})
}

func TestEvaluateLivenessAndWriteMarkdown(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			var payload struct {
				OperationName string         `json:"operationName"`
				Variables     map[string]any `json:"variables"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))

			if payload.OperationName == opListPhenotypes {
				_, _ = w.Write([]byte(`{
					"data": null,
					"errors": [{"message": "ListPhenotypes is not yet implemented"}]
				}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"data": {
					"listStrains": {"totalCount": 1, "nextCursor": 0, "strains": [{"id": "DBS1"}]}
				}
			}`))
		}),
	)
	defer srv.Close()

	state := F.Pipe7(
		livenessRunState{},
		stateCtxLens.Set(context.Background()),
		stateEndpointLens.Set(srv.URL),
		stateTimeoutLens.Set(5*time.Second),
		stateClientLens.Set(graphql.NewClient(srv.URL, srv.Client())),
		stateProbesLens.Set([]LivenessProbe{
			newStrainListProbe("strain_type=ALL", StrainTypeAll),
			newPhenotypeAutocompleteProbe(
				opListPhenotypes,
				"listPhenotypes",
				gqlDocListPhenotypes,
				sampleSearchGrowth,
			),
		}),
		stateSpecsLens.Set([]LivenessOperationSpec{
			newOperationSpec(opStrainList, routeCatalogAndStrains),
			newOperationSpec(opListPhenotypes, routeEditablePhenotype),
		}),
		stateStartedAtLens.Set(time.Now()),
	)

	report, err := E.Unwrap(evaluateLivenessState(state)())
	require.NoError(t, err)
	require.Equal(t, livenessStatusDegraded, report.Status)
	require.Equal(t, 2, report.Total)
	require.Equal(t, 1, report.Passed)
	require.Equal(t, 1, report.Failed)
	require.Len(t, report.Operations, 2)
	require.Len(t, report.Failures, 1)
	require.Equal(
		t,
		"ListPhenotypes is not yet implemented",
		report.Failures[0].ErrorMsg,
	)

	var buf bytes.Buffer
	require.NoError(t, writeLivenessMarkdown(&buf, report))
	mdOut := buf.String()
	require.Contains(t, mdOut, "# GraphQL API Liveness Report")
	require.Contains(t, mdOut, "**DEGRADED**")
	require.Contains(t, mdOut, "## Operation Health")
	require.Contains(t, mdOut, "`StrainList`")
	require.Contains(t, mdOut, "**PASS**")
	require.Contains(t, mdOut, "`ListPhenotypes`")
	require.Contains(t, mdOut, "**FAIL**")
	require.Contains(t, mdOut, "## Failure Summary")
	require.Contains(t, mdOut, "`ListPhenotypes is not yet implemented`")
}

func TestWriteLivenessMarkdownHealthyReport(t *testing.T) {
	opRow := F.Pipe7(
		OperationHealth{},
		opHealthOperationLens.Set(opPublication),
		opHealthRouteAreaLens.Set(routeReferencePreview),
		opHealthTotalLens.Set(1),
		opHealthPassedLens.Set(1),
		opHealthFailedLens.Set(0),
		opHealthLatencyLens.Set(42*time.Millisecond),
		opHealthStatusLens.Set(probeStatusPass),
	)
	report := F.Pipe7(
		LivenessReport{},
		reportEndpointLens.Set("https://graphql.dictybase.dev/graphql"),
		reportStatusLens.Set(livenessStatusHealthy),
		reportTotalLens.Set(1),
		reportPassedLens.Set(1),
		reportFailedLens.Set(0),
		reportDurationLens.Set(42*time.Millisecond),
		reportOperationsLens.Set([]OperationHealth{opRow}),
	)

	var buf bytes.Buffer
	require.NoError(t, writeLivenessMarkdown(&buf, report))
	mdOut := buf.String()
	require.Contains(t, mdOut, "**HEALTHY**")
	require.Contains(t, mdOut, "- All GraphQL liveness probes passed.")
}

func TestOverallLivenessStatus(t *testing.T) {
	require.Equal(t, livenessStatusHealthy, overallLivenessStatus(10, 0))
	require.Equal(t, livenessStatusDegraded, overallLivenessStatus(5, 2))
	require.Equal(t, livenessStatusDown, overallLivenessStatus(0, 5))
}

func TestNoHandRolledPredicates(t *testing.T) {
	loomcheck.Require(t, loomcheck.Config{Roots: []string{"."}})
}

func TestPipelineCheckLivenessEntrypoint(t *testing.T) {
	cfg := pipelinecheck.Config{
		Roots: []string{"."},
		IsEntrypoint: func(name string) bool {
			return strings.HasPrefix(name, "RunLiveness")
		},
		RequirePointFreeSeed:      true,
		RequirePointFreeBranching: true,
	}
	pipelinecheck.Require(t, cfg)
	pipelinecheck.RequireNoNonRawTryCatchCallback(t, cfg)
	pipelinecheck.RequireNoIfErrInTryCatch(t, cfg)
	pipelinecheck.RequireBareIf(t, cfg)
	pipelinecheck.RequireSafePrintf(t, cfg)
	pipelinecheck.RequireNoRedundantFold(t, cfg)
	pipelinecheck.RequireNoHandRolledFileJoinInArrayMap(t, cfg)
	pipelinecheck.RequireNoDirectApply(t, cfg)
	pipelinecheck.RequireNoPopulatedStruct(t, cfg)
}
