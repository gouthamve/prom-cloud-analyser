package commands

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/gouthamve/prom-cloud-analyser/pkg/grafana"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type PrometheusAnalyseCommand struct {
	address     string
	username    string
	password    string
	readTimeout time.Duration

	grafanaMetricsFile string
	outputFile         string
}

func (cmd *PrometheusAnalyseCommand) Register(app *kingpin.Application) {
	prometheusAnalyseCmd := app.Command("prometheus-analyse", "Analyse and output the metrics used in Grafana Dashboards.").Action(cmd.run)

	prometheusAnalyseCmd.Flag("address", "Address of the Prometheus instance, alternatively set $PROMETHEUS_ADDRESS.").
		Envar("PROMETHEUS_ADDRESS").
		Required().
		StringVar(&cmd.address)
	prometheusAnalyseCmd.Flag("username", "Username to use when contacting Prometheus, alternatively set $PROMETHEUS_USER.").
		Envar("PROMETHEUS_USER").
		Default("").
		StringVar(&cmd.username)
	prometheusAnalyseCmd.Flag("password", "Password to use when contacting Prometheus, alternatively set $PROMETHEUS_PASSWORD.").
		Envar("PROMETHEUS_PASSWORD").
		Default("").
		StringVar(&cmd.password)
	prometheusAnalyseCmd.Flag("read-timeout", "timeout for read requests").
		Default("30s").
		DurationVar(&cmd.readTimeout)

	prometheusAnalyseCmd.Flag("grafana-metrics-file", "The path for the input file containing the metrics from grafana-analyse command").
		Default("metrics-in-grafana.json").
		StringVar(&cmd.grafanaMetricsFile)
	prometheusAnalyseCmd.Flag("output", "The path for the output file").
		Default("prometheus-metrics.json").
		StringVar(&cmd.outputFile)
}

func (cmd *PrometheusAnalyseCommand) run(k *kingpin.ParseContext) error {
	metrics := map[string]int{}
	grafanaMetrics := grafana.MetricsInGrafana{}

	byt, err := ioutil.ReadFile(cmd.grafanaMetricsFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(byt, &grafanaMetrics); err != nil {
		return err
	}

	rt := api.DefaultRoundTripper
	if cmd.username != "" {
		rt = config.NewBasicAuthRoundTripper(cmd.username, config.Secret(cmd.password), "", api.DefaultRoundTripper)
	}
	promClient, err := api.NewClient(api.Config{
		Address: cmd.address,

		RoundTripper: rt,
	})

	v1api := v1.NewAPI(promClient)
	for _, metric := range grafanaMetrics.MetricsUsed {
		ctx, cancel := context.WithTimeout(context.Background(), cmd.readTimeout)
		defer cancel()

		query := "count(" + metric + ")"
		result, _, err := v1api.Query(ctx, query, time.Now())
		if err != nil {
			return errors.Wrap(err, "error querying "+query)
		}

		vec := result.(model.Vector)
		if len(vec) == 0 {
			metrics[metric] += 0
			log.Debugln(metric, 0)

			continue
		}

		metrics[metric] += int(vec[0].Value)
		log.Debugln(metric, vec[0].Value)
	}

	output := MetricsInPrometheus{}
	for metric, count := range metrics {
		output.TotalActiveSeries += count
		output.MetricCounts = append(output.MetricCounts, MetricCount{metric, count})
	}
	sort.Slice(output.MetricCounts, func(i, j int) bool {
		return output.MetricCounts[i].Count > output.MetricCounts[j].Count
	})

	out, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	if ioutil.WriteFile(cmd.outputFile, out, os.FileMode(int(0666))); err != nil {
		return err
	}

	return nil
}

type MetricsInPrometheus struct {
	TotalActiveSeries int           `json:"total_active_series"`
	MetricCounts      []MetricCount `json:"metric_counts"`
}

type MetricCount struct {
	Metric string `string:"metric"`
	Count  int    `int:"count"`
}
