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

// Exit error code(s)
const (
	ErrorClient          = 1
	ErrorParseStartTime  = 2
	ErrorParseEndTime    = 3
	ErrorParseStepPeriod = 4
	ErrorFetchMetricName = 5
)

func main() {
	// Defaults (Query ranges for a day)
	now := time.Now()
	now3339 := now.Format(time.RFC3339)
	now3339DayBack := now.Add(-24 * time.Hour).Format(time.RFC3339)

	// Flags
	var (
		addr       = flag.String("addr", "", "address")
		startTime  = flag.String("start_time", now3339DayBack, "start time (current_time - 24h)")
		endTime    = flag.String("end_time", now3339, "end time (current_time)")
		stepPeriod = flag.String("step", "10", "step period (in minutes)")
	)
	flag.Parse()

	// Set flags for logger
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// Get an new client
	client, err := api.NewClient(api.Config{Address: *addr})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating client: ", err)
		os.Exit(ErrorClient)
	}

	// Parse start_time
	st, err := time.Parse(time.RFC3339, *startTime)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing start_time: ", err)
		os.Exit(ErrorParseStartTime)
	}
	// Parse end_time
	et, err := time.Parse(time.RFC3339, *endTime)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing end_time: ", err)
		os.Exit(ErrorParseEndTime)
	}

	// Step Period
	sp, err := strconv.ParseInt(*stepPeriod, 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing step: ", err)
		os.Exit(ErrorParseStepPeriod)
	}

	// Wrap the API
	v1api := v1.NewAPI(client)

	// Get all metric_names using `__name__` reserved label
	mlvs, warnings, err := v1api.LabelValues(context.TODO(), "__name__", st, et)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error fetching metric name(s): ", err)
		os.Exit(ErrorFetchMetricName)
	}
	if len(warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Warning: ", warnings)
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
		result, warnings, err = v1api.QueryRange(context.TODO(), name, r)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error while quering range: ", err)
			continue
		}
		if len(warnings) > 0 {
			fmt.Fprintln(os.Stderr, "Warning: ", warnings)
		}
		// Encode the result to JSON format using the internal Marshaler
		if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, "Error while forming JSON: ", err)
		}
	}
}
