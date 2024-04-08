package servicemap

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/404bro/knative-agent/clients"
	"github.com/404bro/knative-agent/model"
	"github.com/404bro/knative-agent/trace"
	pmodel "github.com/prometheus/common/model"
	"google.golang.org/appengine/log"
)

func UpdateServiceMap(spans map[string]*trace.Span, from time.Time, to time.Time) error {
	edges := getEdges(spans)
	nodes, err := getNodes()
	if err != nil {
		fmt.Printf("error get nodes: %v\n", err)
		return err
	}
	err = clients.DBClient.InsertServiceMap(from, model.BasicServiceMap{Nodes: nodes, Edges: edges})
	if err != nil {
		log.Errorf(context.Background(), "failed to insert service map: %v", err)
	}
	return err
}

func getNodes() ([]model.BasicServiceMapNode, error) {
	serviceList, err := clients.KnsClient.ListServices(context.Background())
	if err != nil {
		return nil, err
	}
	nodes := []model.BasicServiceMapNode{}
	for _, service := range serviceList.Items {
		nodes = append(nodes, model.BasicServiceMapNode{Name: service.Name, Namespace: service.Namespace})
	}
	return nodes, nil
}

func getEdges(spans map[string]*trace.Span) []model.ServiceMapEdge {
	children := make(map[string][]string)
	roots := make(map[string]bool)
	for _, span := range spans {
		if span.ParentSpanID == "" {
			roots[span.SpanID] = true
		} else {
			children[span.ParentSpanID] = append(children[span.ParentSpanID], span.SpanID)
		}
	}

	for key, childrenList := range children {
		sort.Slice(childrenList, func(i, j int) bool {
			return spans[childrenList[i]].StartTime < spans[childrenList[j]].StartTime
		})
		children[key] = childrenList
	}

	edges := make(map[model.ServiceMapEdge]bool)

	isVisible := func(spanId string) bool {
		name := spans[spanId].ServiceName
		return name != ""
	}
	isParallel := func(spanId string) bool {
		return spans[spanId].IsParallel
	}
	addEdge := func(src string, dst string) {
		edges[model.ServiceMapEdge{SrcName: spans[src].ServiceName, SrcNamespace: spans[src].ServiceNamespace,
			DstName: spans[dst].ServiceName, DstNamespace: spans[dst].ServiceNamespace}] = true
	}

	var dfs func(string, string) string
	dfs = func(cur string, pre string) string {
		if isVisible(cur) {
			if pre != "" {
				addEdge(pre, cur)
			}
			pre = cur
		}
		for _, child := range children[cur] {
			r := dfs(child, pre)
			if !isParallel(cur) {
				pre = r
			}
		}
		return pre
	}

	for root, ok := range roots {
		if ok {
			dfs(root, "")
		}
	}

	fmt.Printf("edge size: %d\n", len(edges))
	result := []model.ServiceMapEdge{}
	for edge := range edges {
		result = append(result, edge)
	}
	return result
}

func getMetrics(serviceMap *model.ServiceMap, from time.Time, to time.Time) error {
	for i := range serviceMap.Nodes {
		node := &serviceMap.Nodes[i]

		promQueryStr := fmt.Sprintf("scalar(sum(delta(revision_app_request_count{namespace=\"%s\", service_name=\"%s\"}[%ds]))) / %d",
			node.Namespace, node.Name, int(to.Sub(from).Seconds()), int(to.Sub(from).Seconds()))
		result, _, err := clients.PromClient.Query(context.Background(), promQueryStr, to)
		if err != nil {
			return err
		}
		node.RPS = float64(result.(*pmodel.Scalar).Value)
		if math.IsNaN(node.RPS) {
			node.RPS = 0
		}

		promQueryStr = fmt.Sprintf("scalar((sum(delta(revision_app_request_latencies_sum{namespace=\"%s\", service_name=\"%s\"} [%ds]))) "+
			"/ (sum(delta(revision_app_request_latencies_count{namespace=\"%s\", service_name=\"%s\"} [%ds]))))",
			node.Namespace, node.Name, int(to.Sub(from).Seconds()), node.Namespace, node.Name, int(to.Sub(from).Seconds()))
		result, _, err = clients.PromClient.Query(context.Background(), promQueryStr, to)
		if err != nil {
			return err
		}
		node.Latency = float64(result.(*pmodel.Scalar).Value)
		if math.IsNaN(node.Latency) {
			node.Latency = 0
		}

		promQueryStr = fmt.Sprintf("scalar((sum(delta(revision_app_request_count{namespace=\"%s\", service_name=\"%s\", response_code_class=\"2xx\"} [%ds]))))",
			node.Namespace, node.Name, int(to.Sub(from).Seconds()))
		result, _, err = clients.PromClient.Query(context.Background(), promQueryStr, to)
		if err != nil {
			return err
		}
		success := float64(result.(*pmodel.Scalar).Value)
		if math.IsNaN(success) {
			success = 0
		}

		promQueryStr = fmt.Sprintf("scalar((sum(delta(revision_app_request_count{namespace=\"%s\", service_name=\"%s\", response_code_class=\"5xx\"} [%ds]))))",
			node.Namespace, node.Name, int(to.Sub(from).Seconds()))
		result, _, err = clients.PromClient.Query(context.Background(), promQueryStr, to)
		if err != nil {
			return err
		}
		fail := float64(result.(*pmodel.Scalar).Value)
		if math.IsNaN(fail) {
			fail = 0
		}

		if success+fail == 0 {
			node.Success = 0
		} else {
			node.Success = success / (success + fail)
		}
	}
	return nil
}
