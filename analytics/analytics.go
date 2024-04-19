package analytics

import (
	"fmt"
	"time"

	"github.com/404bro/knative-agent/clients"
	"github.com/404bro/knative-agent/model"
	"github.com/404bro/knative-agent/servicemap"
	"github.com/404bro/knative-agent/trace"
)

func UpdateAnalytics(mTraces []trace.MTrace, from time.Time, to time.Time) {
	for _, trace := range mTraces {
		fmt.Printf("Updating analytics for trace %s\n", trace.TraceID)
		serviceMap, err := servicemap.GetServiceMap(trace.Spans, from, to)
		if err != nil {
			fmt.Printf("Failed to get service map: %v\n", err)
			return
		}
		entry := model.AnalyticsEntry{Time: trace.Time, TraceID: trace.TraceID, ServiceMap: serviceMap}
		err = clients.DBClient.InsertAnalyticsEntry(entry)
		if err != nil {
			fmt.Printf("Failed to insert analytics entry: %v\n", err)
			return
		}
	}
}
