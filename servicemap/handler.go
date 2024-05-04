package servicemap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/404bro/knative-agent/clients"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid param: from", http.StatusBadRequest)
		return
	}
	to, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid param: to", http.StatusBadRequest)
		return
	}
	fmt.Printf("from: %s, to: %s\n", time.UnixMicro(from).String(), time.UnixMicro(to).String())

	graph, err := clients.DBClient.QueryServiceMap(time.UnixMicro(from), time.UnixMicro(to))
	if err != nil {
		fmt.Printf("Failed to query service map: %v", err)
		http.Error(w, "Failed to query service map", http.StatusInternalServerError)
		return
	}
	GetMetrics(&graph, time.UnixMicro(from), time.UnixMicro(to))
	result, err := json.Marshal(graph)
	if err != nil {
		http.Error(w, "Failed to marshal service map", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	nodes, err := getNodes()
	if err != nil {
		http.Error(w, "Failed to get nodes", http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(nodes)
	if err != nil {
		http.Error(w, "Failed to marshal nodes", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
