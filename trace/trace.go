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

func GetMarkedSpans(from time.Time, to time.Time, jaegerURL string) (map[string]*Span, error) {
	spans := make(map[string]*Span)
	services := []string{"broker-ingress.knative-eventing", "activator-service"}
	for _, service := range services {
		retries := 5
		url := fmt.Sprintf("%s/api/traces?end=%d&limit=0&service=%s&start=%d", jaegerURL, to.UnixMicro(), service, from.UnixMicro())
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
			return nil, fmt.Errorf("failed to get traces: %v", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var data Traces
		json.Unmarshal(body, &data)
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
