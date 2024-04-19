package model

import "time"

type AnalyticsEntry struct {
	Time       time.Time  `json:"time" bson:"time"`
	TraceID    string     `json:"traceID" bson:"traceID"`
	ServiceMap ServiceMap `json:"serviceMap" bson:"serviceMap"`
}
