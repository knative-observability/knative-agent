package config

import (
	"os"
	"strconv"
)

var Port string
var Interval int64 // seconds
var JaegerURL string
var PrometheusURL string
var MongoDBURL string

func LoadConfig() {
	Port = "9091"
	Interval = 1
	JaegerURL = "http://simplest-query.default.svc.cluster.local:16686"
	PrometheusURL = "http://prometheus-kube-prometheus-prometheus.default.svc.cluster.local:9090"
	MongoDBURL = "mongodb://mongodb:27017"

	if os.Getenv("PORT") != "" {
		Port = os.Getenv("PORT")
	}
	if os.Getenv("INTERVAL") != "" {
		interval, err := strconv.ParseInt(os.Getenv("INTERVAL"), 10, 64)
		if err != nil {
			panic(err)
		}
		Interval = interval
	}
	if os.Getenv("JAEGER_URL") != "" {
		JaegerURL = os.Getenv("JAEGER_URL")
	}
	if os.Getenv("PROMETHEUS_URL") != "" {
		PrometheusURL = os.Getenv("PROMETHEUS_URL")
	}
}
