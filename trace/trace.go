package trace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func GetTraces(url string) (Traces, error) {
	retries := 5
	var resp *http.Response
	var err error
	for i := 0; i < retries; i++ {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("retrying %d times\n", i+1)
	}
	if err != nil {
		return Traces{}, fmt.Errorf("failed to get traces: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Traces{}, err
	}
	var data Traces
	json.Unmarshal(body, &data)
	return data, nil
}

func GetMarkedSpans(from time.Time, to time.Time, jaegerURL string) (map[string]*Span, error) {
	spans := make(map[string]*Span)
	services := []string{"broker-ingress.knative-eventing", "activator-service"}
	for _, service := range services {
		url := fmt.Sprintf("%s/api/traces?end=%d&limit=0&service=%s&start=%d", jaegerURL, to.UnixMicro(), service, from.UnixMicro())
		data, err := GetTraces(url)
		if err != nil {
			return nil, err
		}
		for _, trace := range data.Data {
			for _, span := range trace.Spans {
				if _, ok := spans[span.SpanID]; ok {
					continue
				}
				spanCopy := span
				spans[span.SpanID] = &spanCopy
			}
		}
	}
	markSpans(spans)
	return spans, nil
}

func markSpans(spans map[string]*Span) {
	for _, span := range spans {
		for _, ref := range span.References {
			if ref.RefType == "CHILD_OF" && ref.SpanID != "" {
				span.ParentSpanID = ref.SpanID
			}
		}
		for _, tag := range span.Tags {
			switch tag.Key {
			case "http.url":
				regexPattern := `^https?://([a-zA-Z0-9-]+)\.([a-zA-Z0-9-]+)\.svc\.cluster\.local$`
				regex := regexp.MustCompile(regexPattern)
				matches := regex.FindStringSubmatch(tag.Value)
				if len(matches) > 2 {
					if !strings.HasSuffix(matches[1], "-kn-channel") {
						span.ServiceName = matches[1]
						span.ServiceNamespace = matches[2]
					}
				}
			case "http.host":
				parts := strings.Split(tag.Value, ".")
				span.IsParallel = strings.HasSuffix(parts[0], "-kn-parallel-kn-channel") ||
					strings.HasSuffix(parts[0], "-kne-trigger-kn-channel")

				regexPattern := `^([a-zA-Z0-9-]+)\.([a-zA-Z0-9-]+)\.svc\.cluster\.local$`
				regex := regexp.MustCompile(regexPattern)
				matches := regex.FindStringSubmatch(tag.Value)
				if len(matches) > 2 {
					span.ActivatorServiceName = matches[1]
					span.ActivatorServiceNamespace = matches[2]
				}
			case "error":
				span.Error = true
			}
		}
	}
	for _, span := range spans {
		if spans[span.ParentSpanID] == nil {
			span.ParentSpanID = ""
		}
	}
}

func GetMarkedErrMTraces(from time.Time, to time.Time, jaegerURL string) ([]MTrace, error) {
	url := fmt.Sprintf("%s/api/traces?end=%d&limit=0&service=broker-ingress.knative-eventing&start=%d&tags={\"error\":\"true\"}", jaegerURL, to.UnixMicro(), from.UnixMicro())
	traces, err := GetTraces(url)
	if err != nil {
		return nil, err
	}
	mTraces := []MTrace{}
	for _, trace := range traces.Data {
		mTrace := MTrace{TraceID: trace.TraceID, Spans: make(map[string]*Span)}
		if len(traces.Data) > 0 {
			mTrace.Time = time.UnixMicro(trace.Spans[0].StartTime)
		}
		for _, span := range trace.Spans {
			if _, ok := mTrace.Spans[span.SpanID]; ok {
				continue
			}
			if span.StartTime < mTrace.Time.UnixMicro() {
				mTrace.Time = time.UnixMicro(span.StartTime)
			}
			spanCopy := span
			mTrace.Spans[span.SpanID] = &spanCopy
		}
		markSpans(mTrace.Spans)
		mTraces = append(mTraces, mTrace)
	}
	return mTraces, nil
}
