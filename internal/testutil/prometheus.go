package testutil

import (
	promclient "github.com/prometheus/client_model/go"
)

// SampleDelta returns the delta of before and after metric sample count for the given name and labels.
func SampleDelta(name string, labels map[string]string, before, after []*promclient.MetricFamily) uint64 {
	sampleBefore := GetSampleCount(name, labels, before)
	sampleAfter := GetSampleCount(name, labels, after)
	return sampleAfter - sampleBefore
}

// GetSampleCount returns the sample count of the given metric name and labels.
func GetSampleCount(name string, labels map[string]string, mfs []*promclient.MetricFamily) uint64 {
	for _, mf := range mfs {
		if mf.GetName() == name {
			for _, metric := range mf.Metric {
				// label counts don't match, no need to match labels.
				if len(metric.Label) != len(labels) {
					continue
				}

				// try to match all labels
				var labelsMatch int
				for _, label := range metric.Label {
					if labels[label.GetName()] == label.GetValue() {
						labelsMatch++
					}
				}
				if labelsMatch == len(labels) {
					return metric.Histogram.GetSampleCount()
				}
			}
		}
	}
	return 0
}
