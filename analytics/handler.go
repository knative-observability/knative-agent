package analytics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/404bro/knative-agent/clients"
	"github.com/404bro/knative-agent/model"
	"github.com/404bro/knative-agent/servicemap"
)

type params struct {
	from      int64
	to        int64
	name      string
	namespace string
}

func getParams(r *http.Request) (params, error) {
	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		return params{}, fmt.Errorf("invalid param: from")
	}
	to, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	if err != nil {
		return params{}, fmt.Errorf("invalid param: to")
	}
	names, ok := r.URL.Query()["name"]
	if !ok || len(names) < 1 {
		return params{}, fmt.Errorf("invalid param: name")
	}
	name := names[0]
	namespaces, ok := r.URL.Query()["namespace"]
	if !ok || len(namespaces) < 1 {
		return params{}, fmt.Errorf("invalid param: namespace")
	}
	namespace := namespaces[0]
	return params{from, to, name, namespace}, nil
}

func TracesHandler(w http.ResponseWriter, r *http.Request) {
	params, err := getParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	traces, err := clients.DBClient.QueryAnalyticTraces(params.name, params.namespace, time.UnixMicro(params.from), time.UnixMicro(params.to))
	if err != nil {
		fmt.Printf("Failed to query analytic traces: %v", err)
		http.Error(w, "Failed to query analytic traces", http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(traces)
	if err != nil {
		http.Error(w, "Failed to marshal analytic traces", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func GraphHandler(w http.ResponseWriter, r *http.Request) {
	params, err := getParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	graph, err := clients.DBClient.QueryAnalyticGraph(params.name, params.namespace, time.UnixMicro(params.from), time.UnixMicro(params.to))
	if err != nil {
		fmt.Printf("Failed to query analytic graph: %v", err)
		http.Error(w, "Failed to query analytic graph", http.StatusInternalServerError)
		return
	}
	servicemap.GetMetrics(&graph, time.UnixMicro(params.from), time.UnixMicro(params.to))
	result, err := json.Marshal(graph)
	if err != nil {
		http.Error(w, "Failed to marshal analytic graph", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	params, err := getParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	graph, err := clients.DBClient.QueryAnalyticGraph(params.name, params.namespace, time.UnixMicro(params.from), time.UnixMicro(params.to))
	if err != nil {
		fmt.Printf("Failed to query analytic graph: %v", err)
		http.Error(w, "Failed to query analytic graph", http.StatusInternalServerError)
		return
	}
	nns := []model.BasicServiceMapNode{}
	for _, node := range graph.Nodes {
		nns = append(nns, model.BasicServiceMapNode{
			Name:      node.Name,
			Namespace: node.Namespace,
		})
	}
	result, err := json.Marshal(nns)
	if err != nil {
		http.Error(w, "Failed to marshal analytic graph", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
