package model

import "time"

type ServiceMap struct {
	Nodes []ServiceMapNode `json:"nodes" bson:"nodes"`
	Edges []ServiceMapEdge `json:"edges" bson:"edges"`
}

type ServiceMapNode struct {
	Name      string  `json:"name" bson:"name"`
	Namespace string  `json:"namespace" bson:"namespace"`
	RPS       float64 `json:"rps" bson:"rps"`
	Latency   float64 `json:"latency" bson:"latency"`
	Success   float64 `json:"success" bson:"success"`
}

type ServiceMapEdge struct {
	SrcName      string `json:"srcName" bson:"srcName"`
	SrcNamespace string `json:"srcNamespace" bson:"srcNamespace"`
	DstName      string `json:"dstName" bson:"dstName"`
	DstNamespace string `json:"dstNamespace" bson:"dstNamespace"`
}

type BasicServiceMapNode struct {
	Name      string `bson:"name"`
	Namespace string `bson:"namespace"`
}

type BasicServiceMap struct {
	From  time.Time             `bson:"from"`
	Nodes []BasicServiceMapNode `bson:"nodes"`
	Edges []ServiceMapEdge      `bson:"edges"`
}
