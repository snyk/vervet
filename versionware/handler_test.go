package versionware_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/versionware"
)

func ExampleHandler() {
	h := versionware.NewHandler([]versionware.VersionHandler{{
		Version: vervet.MustParseVersion("2021-10-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("oct")); err != nil {
				panic(err)
			}
		}),
	}, {
		Version: vervet.MustParseVersion("2021-11-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("nov")); err != nil {
				panic(err)
			}
		}),
	}, {
		Version: vervet.MustParseVersion("2021-09-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("sept")); err != nil {
				panic(err)
			}
		}),
	}}...)

	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := s.Client().Get(s.URL + "?version=2021-10-31")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(respBody))
	// Output: oct
}

func TestHandler(t *testing.T) {
	c := qt.New(t)
	h := versionware.NewHandler([]versionware.VersionHandler{{
		Version: vervet.MustParseVersion("2021-08-01~beta"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("aug beta"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2021-10-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("oct"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2021-11-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("nov"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2021-09-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("sept"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2024-10-15"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("on pivot date"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2024-10-20"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("after pivot date"))
			c.Assert(err, qt.IsNil)
		}),
	}}...)
	tests := []struct {
		requested, resolved string
		contents            string
		status              int
	}{{
		"2021-08-31", "", "Not Found\n", 404,
	}, {
		"", "", "Bad Request\n", 400,
	}, {
		"bad wolf", "", "400 Bad Request", 400,
	}, {
		"2021-09-16", "2021-09-01", "sept", 200,
	}, {
		"2021-10-01", "2021-10-01", "oct", 200,
	}, {
		"2021-10-31", "2021-10-01", "oct", 200,
	}, {
		"2021-11-05", "2021-11-01", "nov", 200,
	}, {
		"2023-02-05", "2021-11-01", "nov", 200,
	}, {
		"2021-08-01", "", "Not Found\n", 404,
	}, {
		"2021-08-01~beta", "2021-08-01~beta", "aug beta", 200,
	}, {
		"2021-09-01~beta", "2021-09-01", "sept", 200,
	}, {
		"2024-10-14", "2021-11-01", "nov", 200,
	}, {
		"2024-10-14~beta", "2021-11-01", "nov", 200,
	}, {
		"2024-10-14~experimental", "2021-11-01", "nov", 200,
	}, {
		"2024-10-15", "2024-10-15", "on pivot date", 200,
	}, {
		"2024-10-16", "2024-10-15", "on pivot date", 200,
	}, {
		"2024-10-20", "2024-10-20", "after pivot date", 200,
	}, {
		"2024-10-20~beta", "", "Not Found\n", 404,
	}, {
		"2024-10-20~experimental", "", "Not Found\n", 404,
	}}
	for i, test := range tests {
		c.Run(fmt.Sprintf("%d requested %s resolved %s", i, test.requested, test.resolved), func(c *qt.C) {
			s := httptest.NewServer(h)
			c.Cleanup(s.Close)
			req, err := http.NewRequest("GET", s.URL+"?version="+test.requested, nil)
			c.Assert(err, qt.IsNil)
			resp, err := s.Client().Do(req)
			c.Assert(err, qt.IsNil)
			defer resp.Body.Close()
			c.Assert(resp.StatusCode, qt.Equals, test.status)
			contents, err := io.ReadAll(resp.Body)
			c.Assert(err, qt.IsNil)
			c.Assert(string(contents), qt.Equals, test.contents)
		})
	}
}

func TestHandler_BetaEndpoints(t *testing.T) {
	c := qt.New(t)
	h := versionware.NewHandler([]versionware.VersionHandler{{
		Version: vervet.MustParseVersion("2021-08-01~beta"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("aug beta"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2021-10-01~beta"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("oct beta"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2021-11-01~experimental"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("nov experimental"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2024-10-15~beta"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("beta on pivot date"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2024-10-20~beta"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("beta after pivot date"))
			c.Assert(err, qt.IsNil)
		}),
	}, {
		Version: vervet.MustParseVersion("2024-10-25"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("ga after pivot date"))
			c.Assert(err, qt.IsNil)
		}),
	}}...)
	tests := []struct {
		requested, resolved string
		contents            string
		status              int
	}{{
		"2021-07-31~beta", "", "Not Found\n", 404,
	}, {
		"2021-09-16~beta", "2021-08-01~beta", "aug beta", 200,
	}, {
		"2021-10-01~beta", "2021-10-01~beta", "oct beta", 200,
	}, {
		"2021-11-05~experimental", "2021-11-01~experimental", "nov experimental", 200,
	}, {
		"2024-10-15", "2024-10-15~beta", "beta on pivot date", 200,
	}, {
		"2024-10-15~beta", "", "Not Found\n", 404,
	}, {
		"2024-10-15~experimental", "", "Not Found\n", 404,
	}, {
		"2024-10-16", "2024-10-15~beta", "beta on pivot date", 200,
	}, {
		"2024-10-20", "2024-10-20~beta", "beta after pivot date", 200,
	}, {
		"2024-10-21", "2024-10-20~beta", "beta after pivot date", 200,
	}, {
		"2024-10-26", "2024-10-25", "ga after pivot date", 200,
	}}
	for i, test := range tests {
		c.Run(fmt.Sprintf("%d requested %s resolved %s", i, test.requested, test.resolved), func(c *qt.C) {
			s := httptest.NewServer(h)
			c.Cleanup(s.Close)
			req, err := http.NewRequest("GET", s.URL+"?version="+test.requested, nil)
			c.Assert(err, qt.IsNil)
			resp, err := s.Client().Do(req)
			c.Assert(err, qt.IsNil)
			defer resp.Body.Close()
			c.Assert(resp.StatusCode, qt.Equals, test.status)
			contents, err := io.ReadAll(resp.Body)
			c.Assert(err, qt.IsNil)
			c.Assert(string(contents), qt.Equals, test.contents)
		})
	}
}
