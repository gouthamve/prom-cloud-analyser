package grafana

type MetricsInGrafana struct {
	MetricsUsed []string           `json:"metricsUsed"`
	Dashboards  []DashboardMetrics `json:"dashboards"`
}

type DashboardMetrics struct {
	Slug        string   `json:"slug"`
	UID         string   `json:"uid,omitempty"`
	Title       string   `json:"title"`
	Metrics     []string `json:"metrics"`
	ParseErrors []error  `json:"parse_errors"`
}
