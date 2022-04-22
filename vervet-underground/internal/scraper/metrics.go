package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metrics = struct {
	runDuration     prometheus.Histogram
	runError        prometheus.Counter
	scrapeDuration  *prometheus.HistogramVec
	scrapeError     *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}{
	runDuration: promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "vu_scraper_run_duration_seconds",
		Help:    "Time spent scraping and collating service specs",
		Buckets: []float64{1, 5, 10, 25, 45, 60, 90},
	}),
	runError: promauto.NewCounter(prometheus.CounterOpts{
		Name: "vu_scraper_run_error_total",
		Help: "Count of errors during a scraper execution",
	}),
	scrapeDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vu_scraper_service_scrape_duration_seconds",
		Help:    "Time spent scraping a service",
		Buckets: []float64{1, 5, 10, 25, 45, 60, 90},
	}, []string{"service"}),
	scrapeError: promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vu_scraper_service_scrape_error_total",
		Help: "Count of errors encountered scraping services",
	}, []string{"service"}),
	requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vu_scraper_service_scrape_http_duration_seconds",
		Help:    "Time spent on a service http call",
		Buckets: []float64{0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"host", "method", "status"}),
}

// durationAllowList is a list of regex matchers of path patterns. This is used in DurationTransport.
// Duration will only be observed if the incoming path matches any of the items in the list.
var durationAllowList = []*regexp.Regexp{
	regexp.MustCompile(`/openapi$`),
	regexp.MustCompile(`/openapi/[1-9][0-9]{3}-[0-1][0-9]-[0-3][0-9](~\w+)?$`),
}

// DurationTransport returns a http.RoundTripper for tracking http calls made from the Scraper.
func DurationTransport(next http.RoundTripper) http.RoundTripper {
	return promhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		start := time.Now()
		resp, err := next.RoundTrip(r)
		if err == nil && pathAllowed(durationAllowList, r.URL.Path) {
			status := fmt.Sprintf("%dxx", resp.StatusCode/100)
			metrics.requestDuration.WithLabelValues(r.URL.Host, r.Method, status).Observe(time.Since(start).Seconds())
		}
		return resp, err
	})
}

func pathAllowed(allowList []*regexp.Regexp, path string) bool {
	for _, check := range allowList {
		if check.MatchString(path) {
			return true
		}
	}
	return false
}
