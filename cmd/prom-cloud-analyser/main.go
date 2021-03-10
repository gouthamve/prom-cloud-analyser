package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/gouthamve/prom-cloud-analyser/pkg/commands"
)

var (
	logConfig                commands.LoggerConfig
	grafanaAnalyseCommand    commands.GrafanaAnalyseCommand
	prometheusAnalyseCommand commands.PrometheusAnalyseCommand
)

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("analyse-metrics", "A command-line tool to analyse metric use.")
	logConfig.Register(app)
	grafanaAnalyseCommand.Register(app)
	prometheusAnalyseCommand.Register(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
