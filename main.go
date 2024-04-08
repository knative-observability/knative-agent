package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/404bro/knative-agent/clients"
	"github.com/404bro/knative-agent/config"
	"github.com/404bro/knative-agent/metrics"
	"github.com/404bro/knative-agent/servicemap"
	"github.com/404bro/knative-agent/trace"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config.LoadConfig()
	err := clients.ConnectClients()
	if err != nil {
		log.Fatalf("Failed to connect clients: %v", err)
	}
	log.Printf("Starting knative agent on port %s\n", config.Port)
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.Handle("/map", http.HandlerFunc(servicemap.Handler))
	server := http.Server{Addr: ":" + config.Port, Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	from := time.Now().Add(-time.Duration(config.Interval)*time.Second - time.Second)
	to := time.Now().Add(-time.Duration(config.Interval)*time.Second - time.Second)
	for {
		go func() {
			from = to
			to = time.Now().Add(-time.Second)
			spans, err := trace.GetMarkedSpans(from, to, config.JaegerURL)
			if err == nil {
				fmt.Printf("[%d %d] Got %d spans\n", from.UnixMilli(), to.UnixMilli(), len(spans))
				go metrics.UpdateColdstartMetrics(spans, m.ServiceColdstartCount, from, to)
				go servicemap.UpdateServiceMap(spans, from, to)
			} else {
				fmt.Printf("Failed to get marked spans: %v\n", err)
			}
		}()
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
