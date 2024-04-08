package db

import (
	"time"

	"github.com/404bro/knative-agent/model"
)

type Database interface {
	QueryServiceMap(from time.Time, to time.Time) (model.ServiceMap, error)
	InsertServiceMap(from time.Time, serviceMap model.BasicServiceMap) error
}
