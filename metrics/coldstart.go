package metrics

import (
	"fmt"
	"time"

	"github.com/404bro/knative-agent/trace"
	"github.com/prometheus/client_golang/prometheus"
)

func UpdateColdstartMetrics(spans map[string]*trace.Span, cv *prometheus.CounterVec, from time.Time, to time.Time) {
	for _, span := range spans {
		if span.OperationName == "throttler_try" {
			parentSpan := spans[span.ParentSpanID]
			if parentSpan == nil {
				fmt.Printf("Parent span not found for span %s, trace %s\n", span.SpanID, span.TraceID)
				continue
			}
			serviceName := parentSpan.ActivatorServiceName
			serviceNamespace := parentSpan.ActivatorServiceNamespace
			if span.Duration > 1e3 {
				cv.WithLabelValues(serviceName, serviceNamespace, "cold").Inc()
			} else {
				cv.WithLabelValues(serviceName, serviceNamespace, "hot").Inc()
			}
		}
	}
}
