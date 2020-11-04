// Queries Prometheus Server to do a query range for a given start_time
// and end_time for all metrics present in the TSDB of Prometheus.

// All error(s) and warning(s) are directed to stderr
// Only results are directed to stdout
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func main() {
	// Flags
	var addr = flag.String("addr", "", "address")
	var startTime = flag.String("start_time", "", "start time")
	var endTime = flag.String("end_time", "", "end time")
	var stepPeriod = flag.String("step", "10", "step period (in minutes)")
	flag.Parse()

	// Set flags for logger
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// Get an new client
	client, err := api.NewClient(api.Config{Address: *addr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Parse start_time
	st, err := time.Parse(time.RFC3339, *startTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing start_time: %v\n", err)
		os.Exit(1)
	}
	// Parse end_time
	et, err := time.Parse(time.RFC3339, *endTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing end_time: %v\n", err)
		os.Exit(1)
	}

	// Step Period
	sp, err := strconv.ParseInt(*stepPeriod, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing step: %v\n", err)
		os.Exit(1)
	}

	// Wrap the API
	var v1api = v1.NewAPI(client)

	// Don't do anything; for now keep it TODO
	var ctx = context.TODO()

	// Get all metric_names using `__name__` reserved label
	mlvs, warnings, err := v1api.LabelValues(ctx, "__name__", st, et)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching metric name(s): %v\n", err)
	}
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", warnings)
	}

	// Set range
	r := v1.Range{
		Start: st,
		End:   et,
		Step:  time.Duration(sp) * time.Minute,
	}

	var name string
	var result model.Value

	// Loop through metric name(s)
	for _, mlv := range mlvs {
		name = string(mlv)
		// Do a range query
		result, warnings, err = v1api.QueryRange(ctx, name, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while quering range: %v\n", err)
			continue
		}
		if len(warnings) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", warnings)
		}
		// Encode the result to JSON format using the internal Marshaler
		if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "Error while forming JSON: %v\n", err)
		}
	}
}
