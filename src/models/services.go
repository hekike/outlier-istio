package models

import (
	"context"
	"fmt"
	"time"

	"github.com/hekike/outlier-istio/src/utils"
	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

const reqsFmt = "sum(rate(istio_requests_total{reporter=\"destination\"}[%s])) by (source_workload, destination_workload, source_app, destination_app)"

/*Service struct.*/
type Service struct {
	Name       string   `json:"name"`
	Downstream []string `json:"downstream"`
}

func (srv *Service) addDownstream(item string) []string {
	srv.Downstream = utils.UniqueStrings(append(srv.Downstream, item))
	return srv.Downstream
}

/*GetServices returns the services.*/
func GetServices(addr string) (map[string]Service, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)
	query := fmt.Sprintf(reqsFmt, "60s")

	val, err := api.Query(context.Background(), query, time.Now())
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Vector)

	services := make(map[string]Service)
	for _, sample := range matrix {
		metrics := sample.Metric
		sourceApp := string(metrics["source_app"])

		var srv Service
		if v, found := services[sourceApp]; found {
			srv = v
		} else {
			srv = Service{Name: sourceApp}
		}

		detinationApp := string(metrics["destination_app"])
		srv.addDownstream(detinationApp)

		services[sourceApp] = srv
	}

	return services, nil
}
