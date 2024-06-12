package scraper_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/snyk/vervet/v6/internal/scraper"
	"github.com/snyk/vervet/v6/internal/testutil"
)

type mockRoundTripper struct {
	statusCode int
}

func (m mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: m.statusCode}, nil
}

func TestDurationTransport(t *testing.T) {
	c := qt.New(t)

	tcs := map[string]struct {
		req          *http.Request
		resStatus    int
		labelHost    string
		labelMethod  string
		labelStatus  string
		expectSample bool
	}{
		"collect metrics on /openapi": {
			req:          httptest.NewRequest(http.MethodGet, "http://test.test/openapi", http.NoBody),
			resStatus:    200,
			labelStatus:  "2xx",
			labelHost:    "test.test",
			labelMethod:  "GET",
			expectSample: true,
		},
		"collect metrics on /rest/openapi": {
			req:          httptest.NewRequest(http.MethodGet, "http://test.test/rest/openapi", http.NoBody),
			resStatus:    200,
			labelStatus:  "2xx",
			labelHost:    "test.test",
			labelMethod:  "GET",
			expectSample: true,
		},
		"collect metrics on /openapi/2022-01-01": {
			req:          httptest.NewRequest(http.MethodGet, "http://test.test/openapi/2022-01-01", http.NoBody),
			resStatus:    200,
			labelStatus:  "2xx",
			labelHost:    "test.test",
			labelMethod:  "GET",
			expectSample: true,
		},
		"collect metrics on /openapi with 4xx": {
			req:          httptest.NewRequest(http.MethodGet, "http://test.test/openapi/2022-01-01", http.NoBody),
			resStatus:    404,
			labelStatus:  "4xx",
			labelHost:    "test.test",
			labelMethod:  "GET",
			expectSample: true,
		},
		"does not collect metrics on /coolapi": {
			req:          httptest.NewRequest(http.MethodGet, "http://test.test/coolapi", http.NoBody),
			resStatus:    200,
			labelStatus:  "2xx",
			labelHost:    "test.test",
			labelMethod:  "GET",
			expectSample: false,
		},
	}

	metricName := "vu_scraper_service_scrape_http_duration_seconds"
	for name, tc := range tcs {
		c.Run(name, func(c *qt.C) {
			before, err := prometheus.DefaultGatherer.Gather()
			c.Assert(err, qt.IsNil)

			_, err = scraper.DurationTransport(mockRoundTripper{tc.resStatus}).RoundTrip(tc.req)
			c.Assert(err, qt.IsNil)

			after, err := prometheus.DefaultGatherer.Gather()
			c.Assert(err, qt.IsNil)

			labels := map[string]string{
				"host":   tc.labelHost,
				"method": tc.labelMethod,
				"status": tc.labelStatus,
			}
			if tc.expectSample {
				c.Assert(testutil.SampleDelta(metricName, labels, before, after),
					qt.Equals, uint64(1))
			} else {
				c.Assert(testutil.SampleDelta(metricName, labels, before, after),
					qt.Equals, uint64(0))
			}
		})
	}
}
