package clients

import (
	"context"

	"github.com/404bro/knative-agent/config"
	"github.com/404bro/knative-agent/db"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"go.mongodb.org/mongo-driver/mongo"
	mongoopt "go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/client-go/rest"
	knsclient "knative.dev/client/pkg/serving/v1"
	knsclientv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

var KnsClient knsclient.KnServingClient
var PromClient promv1.API
var DBClient db.Database

func ConnectClients() error {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	knsClientv1, err := knsclientv1.NewForConfig(k8sConfig)
	if err != nil {
		return err
	}
	KnsClient = knsclient.NewKnServingClient(knsClientv1, "")

	promClient, err := promapi.NewClient(promapi.Config{
		Address: config.PrometheusURL,
	})
	if err != nil {
		return err
	}
	PromClient = promv1.NewAPI(promClient)

	mongoOptions := mongoopt.Client().ApplyURI(config.MongoDBURL)
	mongoClient, err := mongo.Connect(context.Background(), mongoOptions)
	if err != nil {
		return err
	}
	DBClient = db.NewMongoDB(mongoClient, "knative")
	return nil
}
