package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gouthamve/prom-cloud-analyser/pkg/grafana"
	"github.com/grafana-tools/sdk"
	"github.com/prometheus/prometheus/promql/parser"
	"gopkg.in/alecthomas/kingpin.v2"
)

type GrafanaAnalyseCommand struct {
	address     string
	apiKey      string
	readTimeout time.Duration

	outputFile string
}

func (cmd *GrafanaAnalyseCommand) Register(app *kingpin.Application) {
	remoteReadCmd := app.Command("grafana-analyse", "Analyse and output the metrics used in Grafana Dashboards.").Action(cmd.run)

	remoteReadCmd.Flag("address", "Address of the Grafana instance, alternatively set $GRAFANA_ADDRESS.").
		Envar("GRAFANA_ADDRESS").
		Required().
		StringVar(&cmd.address)
	remoteReadCmd.Flag("key", "Api key to use when contacting Grafana, alternatively set $GRAFANA_API_KEY.").
		Envar("GRAFANA_API_KEY").
		Default("").
		StringVar(&cmd.apiKey)
	remoteReadCmd.Flag("read-timeout", "timeout for read requests").
		Default("300s").
		DurationVar(&cmd.readTimeout)
	remoteReadCmd.Flag("output", "The path for the output file").
		Default("metrics-in-grafana.json").
		StringVar(&cmd.outputFile)
}

func (cmd *GrafanaAnalyseCommand) run(k *kingpin.ParseContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), cmd.readTimeout)
	defer cancel()

	c := sdk.NewClient(cmd.address, cmd.apiKey, sdk.DefaultHTTPClient)
	boardLinks, err := c.SearchDashboards(ctx, "", false)
	if err != nil {
		return err
	}

	output := grafana.MetricsInGrafana{}
	overallMetrics := map[string]struct{}{}

	for _, link := range boardLinks {
		board, _, err := c.GetDashboardByUID(ctx, link.UID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s %s\n", err, link.UID, link.Title)
			continue
		}

		metrics, errs := parseMetricsInBoard(board)

		metricsInBoard := make([]string, 0, len(metrics))
		for metric := range metrics {
			if metric == "" {
				continue
			}

			metricsInBoard = append(metricsInBoard, metric)
			overallMetrics[metric] = struct{}{}
		}
		sort.Strings(metricsInBoard)

		output.Dashboards = append(output.Dashboards, grafana.DashboardMetrics{
			Slug:        board.Slug,
			UID:         board.UID,
			Title:       board.Title,
			Metrics:     metricsInBoard,
			ParseErrors: errs,
		})
	}

	metricsUsed := make([]string, 0, len(overallMetrics))
	for metric := range overallMetrics {
		metricsUsed = append(metricsUsed, metric)
	}
	sort.Strings(metricsUsed)

	output.MetricsUsed = metricsUsed
	out, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	if ioutil.WriteFile(cmd.outputFile, out, os.FileMode(int(0666))); err != nil {
		return err
	}

	return nil
}

func parseMetricsInBoard(board sdk.Board) (map[string]struct{}, []error) {
	metrics := map[string]struct{}{}
	parseErrors := make([]error, 0)

	// Iterate through all the panels and collect metrics
	for _, panel := range board.Panels {
		metricsFromPanel(*panel, metrics, parseErrors)
	}

	// Iterate through all the rows and collect metrics
	for _, row := range board.Rows {
		for _, panel := range row.Panels {
			metricsFromPanel(panel, metrics, parseErrors)
		}
	}

	return metrics, parseErrors
}

func metricsFromPanel(panel sdk.Panel, metrics map[string]struct{}, parseErrors []error) {
	if panel.GetTargets() == nil {
		return
	}

	for _, target := range *panel.GetTargets() {
		// Prometheus has this set.
		if target.Expr == "" {
			continue
		}

		query := target.Expr
		query = strings.ReplaceAll(query, `$__interval`, "5m")
		query = strings.ReplaceAll(query, `$interval`, "5m")
		query = strings.ReplaceAll(query, `$resolution`, "5s")
		expr, err := parser.ParseExpr(query)
		if err != nil {
			parseErrors = append(parseErrors, err)
			continue
		}

		parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
			if n, ok := node.(*parser.VectorSelector); ok {
				metrics[n.Name] = struct{}{}
			}

			return nil
		})
	}
}
