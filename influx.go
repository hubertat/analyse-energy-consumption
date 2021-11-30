package main

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Influx struct {
	Host         string
	Bucket       string
	Organization string
	Token        string

	Tags map[string]string `json:",omitempty"`
}

func (inf *Influx) readQuery(query string) (*api.QueryTableResult, error) {
	client := influxdb2.NewClient(inf.Host, inf.Token)
	queryApi := client.QueryAPI(inf.Organization)

	return queryApi.Query(context.Background(), query)
}

func (inf *Influx) energyByHourQuery(start, stop time.Time) (query string, err error) {
	query = fmt.Sprintf(`from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "energy")
	|> filter(fn: (r) => r["_field"] == "energy_import")
	|> filter(fn: (r) => r["ac"] == "active")
	|> difference()
	|> aggregateWindow(every: 1h, fn: sum, createEmpty: false)`, inf.Bucket, start.Format(time.RFC3339), stop.Format(time.RFC3339))

	return
}
