# Prometheus Cloud Analyser

This repo contains a set of tools that can be run against your Grafana, Prometheus and Cortex clusters to analyse which metrics are actually being used and which metrics can be dropped.

## grafana-analyse 

This command will be run against your Grafana instance and it will download the dashboards and pick the Prometheus metrics that are used in the queries. The output is a JSON file.

#### Configuration

| Env Variables     | Flag      | Description                                                                                                   |
| ----------------- | --------- | ------------------------------------------------------------------------------------------------------------- |
| GRAFANA_ADDRESS    | `address` | Address of the Grafana instance.                                                              |
| GRAFANA_API_KEY    | `key`     | The API Key for the Grafana instance. Create a key using the following instructions: https://grafana.com/docs/grafana/latest/http_api/auth/ |
| __ | `output`      | The output file path. metrics-in-grafana.json by default.  |

#### Running the command

```
prom-cloud-analyser grafana-analyse --address=<grafana-address> --key=<API-Key>
```

## prometheus-analyse 

This command will be run against your Prometheus / GrafanaCloud instance and it will use the output from `grafana-analyse` show you how many series in the Prometheus server are actually being used in dashboards. The output is a JSON file

#### Configuration

| Env Variables     | Flag      | Description                                                                                                   |
| ----------------- | --------- | ------------------------------------------------------------------------------------------------------------- |
| PROMETHEUS_ADDRESS    | `address` | Address of the Prometheus  instance.                                                              |
| PROMETHEUS_USER    | `username`   |  If you're using GrafanaCloud this is your instance ID. |
| PROMETHEUS_PASSWORD    | `password`   |  If you're using GrafanaCloud this is your API Key. |
| __ | `grafana-metrics-file`      | The input file path. metrics-in-grafana.json by default.  |
| __ | `output`      | The output file path. prometheus-metrics.json by default.  |

#### Running the command

```
prom-cloud-analyser prometheus-analyse --address=https://prometheus-us-central1.grafana.net/api/prom --username=<1234> --password=<API-Key> --log.level=debug
```

### License

Licensed Apache 2.0, see [LICENSE](LICENSE).