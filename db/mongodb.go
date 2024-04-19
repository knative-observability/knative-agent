package db

import (
	"context"
	"fmt"
	"time"

	"github.com/404bro/knative-agent/config"
	"github.com/404bro/knative-agent/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDB struct {
	serviceMapCollection *mongo.Collection
	analyticsCollection  *mongo.Collection
}

func NewMongoDB(client *mongo.Client, dbName string) *MongoDB {
	serviceMapCollection := client.Database(dbName).Collection("servicemap")
	analyticsCollection := client.Database(dbName).Collection("analytics")
	return &MongoDB{
		serviceMapCollection: serviceMapCollection,
		analyticsCollection:  analyticsCollection,
	}
}

func (db *MongoDB) QueryServiceMap(from time.Time, to time.Time) (model.ServiceMap, error) {
	nodeMap := make(map[model.BasicServiceMapNode]bool)
	edgeMap := make(map[model.ServiceMapEdge]bool)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	filter := bson.M{"from": bson.M{"$lt": to, "$gt": from.Add(-time.Duration(config.Interval) * time.Second)}}
	cursor, err := db.serviceMapCollection.Find(ctx, filter)
	if err != nil {
		return model.ServiceMap{}, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var entry model.BasicServiceMap
		err := cursor.Decode(&entry)
		if err != nil {
			return model.ServiceMap{}, err
		}
		for _, node := range entry.Nodes {
			nodeMap[node] = true
		}
		for _, edge := range entry.Edges {
			edgeMap[edge] = true
		}
	}
	result := model.ServiceMap{}
	for node := range nodeMap {
		result.Nodes = append(result.Nodes, model.ServiceMapNode{
			Name:      node.Name,
			Namespace: node.Namespace,
			RPS:       0,
			Latency:   0,
			Success:   0,
		})
	}
	for edge := range edgeMap {
		result.Edges = append(result.Edges, edge)
	}
	return result, nil
}

func (db *MongoDB) InsertServiceMap(from time.Time, serviceMap model.BasicServiceMap) error {
	entry := model.BasicServiceMap{
		From:  from,
		Nodes: make([]model.BasicServiceMapNode, 0),
		Edges: serviceMap.Edges,
	}
	for _, node := range serviceMap.Nodes {
		entry.Nodes = append(entry.Nodes, model.BasicServiceMapNode{
			Name:      node.Name,
			Namespace: node.Namespace,
		})
	}
	fmt.Printf("inserting %d nodes and %d edges\n", len(entry.Nodes), len(entry.Edges))
	_, err := db.serviceMapCollection.InsertOne(context.Background(), entry)
	return err
}

func (db *MongoDB) InsertAnalyticsEntry(entry model.AnalyticsEntry) error {
	_, err := db.analyticsCollection.InsertOne(context.Background(), entry)
	return err
}

func (db *MongoDB) QueryAnalyticTraces(name string, namespace string, from time.Time, to time.Time) ([]string, error) {
	result := []string{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	filter := bson.M{
		"time": bson.M{
			"$lt": to,
			"$gt": from.Add(-time.Duration(config.Interval) * time.Second),
		},
		"serviceMap.nodes": bson.M{
			"$elemMatch": bson.M{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
	cursor, err := db.analyticsCollection.Find(ctx, filter)
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var entry model.AnalyticsEntry
		err := cursor.Decode(&entry)
		if err != nil {
			return []string{}, err
		}
		result = append(result, entry.TraceID)
	}
	return result, nil
}

func (db *MongoDB) QueryAnalyticGraph(name string, namespace string, from time.Time, to time.Time) (model.ServiceMap, error) {
	nodeMap := make(map[model.ServiceMapNode]bool)
	edgeMap := make(map[model.ServiceMapEdge]bool)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	filter := bson.M{
		"time": bson.M{"$lt": to, "$gt": from.Add(-time.Duration(config.Interval) * time.Second)},
		"serviceMap.nodes": bson.M{
			"$elemMatch": bson.M{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
	cursor, err := db.analyticsCollection.Find(ctx, filter)
	if err != nil {
		return model.ServiceMap{}, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var entry model.AnalyticsEntry
		err := cursor.Decode(&entry)
		if err != nil {
			return model.ServiceMap{}, err
		}
		for _, node := range entry.ServiceMap.Nodes {
			nodeMap[node] = true
		}
		for _, edge := range entry.ServiceMap.Edges {
			edgeMap[edge] = true
		}
	}
	result := model.ServiceMap{}
	for node := range nodeMap {
		result.Nodes = append(result.Nodes, node)
	}
	for edge := range edgeMap {
		result.Edges = append(result.Edges, edge)
	}
	return result, nil
}
