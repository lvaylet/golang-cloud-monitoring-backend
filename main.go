package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
)

// readTimeSeriesAlign reads the last 20 minutes of the given metric and aligns
// everything on 10 minute intervals.
func readTimeSeriesAlign(w io.Writer, projectID string) error {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("NewMetricClient: %v", err)
	}
	defer client.Close()
	startTime := time.Now().UTC().Add(time.Minute * -20).Unix()
	endTime := time.Now().UTC().Unix()
	req := &monitoringpb.ListTimeSeriesRequest{
		// See https://pkg.go.dev/cloud.google.com/go/monitoring/apiv3/v2/monitoringpb#ListTimeSeriesRequest.
		Name: "projects/" + projectID,
		// Filter: `metric.type="compute.googleapis.com/instance/cpu/utilization"`,
		Filter: `metric.type="run.googleapis.com/request_count"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime,
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime,
			},
		},
		Aggregation: &monitoringpb.Aggregation{
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_MEAN,
			AlignmentPeriod: &duration.Duration{
				Seconds: 600,
			},
		},
	}
	iter := client.ListTimeSeries(ctx, req)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read time series value: %v", err)
		}
		fmt.Fprintln(w, resp.GetMetric().GetLabels()["instance_name"])
		fmt.Fprintf(w, "\tNow: %.4f\n", resp.GetPoints()[0].GetValue().GetDoubleValue())
		if len(resp.GetPoints()) > 1 {
			fmt.Fprintf(w, "\t10 minutes ago: %.4f\n", resp.GetPoints()[1].GetValue().GetDoubleValue())
		}
	}
	fmt.Fprintln(w, "Done")
	return nil
}

// readTimeSeriesValue reads the TimeSeries for the value specified by metric type in a time window from the last 20 minutes.
func readTimeSeriesValue(w io.Writer, projectID, metricType string) error {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("NewMetricClient: %v", err)
	}
	defer client.Close()
	startTime := time.Now().UTC().Add(time.Minute * -20).Unix()
	endTime := time.Now().UTC().Unix()

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + projectID,
		Filter: fmt.Sprintf("metric.type=\"%s\"", metricType),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime,
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime,
			},
		},
	}
	iter := client.ListTimeSeries(ctx, req)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read time series value: %v ", err)
		}
		fmt.Fprintf(w, "%+v\n", resp)
	}
	fmt.Fprintln(w, "Done")
	return nil
}

func main() {
	var projectID = "slo-generator-demo"
	readTimeSeriesAlign(os.Stdout, projectID)
	readTimeSeriesValue(os.Stdout, projectID, "run.googleapis.com/request_count")
}
