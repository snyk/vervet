package versionware_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v3"
	"github.com/snyk/vervet/v3/versionware"
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
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(respBody))
	// Output: oct
}

func TestHandler(t *testing.T) {
	c := qt.New(t)
	h := versionware.NewHandler([]versionware.VersionHandler{{
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
			contents, err := ioutil.ReadAll(resp.Body)
			c.Assert(err, qt.IsNil)
			c.Assert(string(contents), qt.Equals, test.contents)
		})
	}
}
