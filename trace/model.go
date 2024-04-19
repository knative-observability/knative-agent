package trace

import "time"

type Traces struct {
	Data []Trace `json:"data"`
}

type Trace struct {
	TraceID string `json:"traceID"`
	Spans   []Span `json:"spans"`
}

type MTrace struct {
	TraceID string
	Time    time.Time
	Spans   map[string]*Span
}

type Span struct {
	TraceID       string      `json:"traceID"`
	SpanID        string      `json:"spanID"`
	OperationName string      `json:"operationName"`
	References    []Reference `json:"references"`
	StartTime     int64       `json:"startTime"`
	Duration      int64       `json:"duration"`
	Tags          []Tag       `json:"tags"`

	ParentSpanID              string `json:"parentSpanID,omitempty"`
	ServiceName               string `json:"serviceName,omitempty"`
	ServiceNamespace          string `json:"serviceNamespace,omitempty"`
	ActivatorServiceName      string `json:"activatorServiceName,omitempty"`
	ActivatorServiceNamespace string `json:"activatorServiceNamespace,omitempty"`
	IsParallel                bool   `json:"isParallel,omitempty"`
	Error                     bool   `json:"error,omitempty"`
}

type Reference struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

type Tag struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Services struct {
	Data []string `json:"data"`
}
