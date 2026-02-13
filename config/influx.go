package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxClient struct {
	Client influxdb2.Client
}

func NewInfluxClient() *InfluxClient {
	influxUrl := os.Getenv("INFLUX_URI")
	token := os.Getenv("INFLUX_TOKEN")
	if influxUrl == "" || token == "" {
		panic("INFLUX_URL and INFLUX_TOKEN environment variables must be set")
	}
	var client influxdb2.Client
	client = influxdb2.NewClient(
		influxUrl,
		token,
	)
	return &InfluxClient{
		Client: client,
	}
}

func (client InfluxClient) ExecuteQuery(experimentId string, bucket string, deviceAddress string, measurement string, field string) ([]string, []float64, error) {
	// build base flux query; append _field filter only when field is non-empty
	flux := fmt.Sprintf(`
		from(bucket: "%s")
		  |> range(start: -%ds)
		  |> filter(fn: (r) => r["experimentId"] == "%s")
		  |> filter(fn: (r) => r["deviceAddress"] == "%s")
		  |> filter(fn: (r) => r["_measurement"] == "%s")
		`,
		bucket,
		5,
		experimentId,
		deviceAddress,
		measurement,
	)
	if strings.TrimSpace(field) != "" {
		// append field filter
		flux += fmt.Sprintf("\n\t\t  |> filter(fn: (r) => r[\"_field\"] == \"%s\")\n\t\t", field)
	}
	log.Printf("Query: %s", flux)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	queryAPI := client.Client.QueryAPI("003e6c7c0dc0eb8b")
	result, err := queryAPI.Query(ctx, flux)
	if err != nil {
		return nil, nil, err
	}
	var times []string
	var values []float64

	for result.Next() {
		rec := result.Record()
		// _time -> time, _value -> numeric value
		t := rec.Time().Format(time.RFC3339)
		// careful with type assertion: _value often float64
		switch v := rec.Value().(type) {
		case float64:
			times = append(times, t)
			values = append(values, v)
		case int64:
			times = append(times, t)
			values = append(values, float64(v))
		default:
			// skip non-numeric or handle as needed
		}
	}
	if result.Err() != nil {
		return nil, nil, result.Err()
	}
	return times, values, nil
}
