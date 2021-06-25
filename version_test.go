package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	. "github.com/snyk/vervet"
)

func TestParseVersion(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		vs  string
		v   Version
		err string
	}{{
		vs: "2021-01-01",
		v:  Version("2021-01-01"),
	}, {
		vs: "beta",
		v:  VersionBeta,
	}, {}, {
		vs: "experimental",
		v:  VersionExperimental,
	}, {
		vs:  "unknown",
		err: `invalid version "unknown"`,
	}}
	for i := range tests {
		v, err := ParseVersion(tests[i].vs)
		if tests[i].err != "" {
			c.Assert(err, qt.ErrorMatches, tests[i].err)
		} else {
			c.Assert(v, qt.Equals, tests[i].v)
		}
	}
}

func TestVersionOrder(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		l, r string
		cmp  int
	}{{
		// Compare some date versions
		l: "2020-12-31", r: "2021-01-01", cmp: -1,
	}, {
		l: "2021-01-01", r: "2020-12-31", cmp: 1,
	}, {
		l: "2021-01-01", r: "2021-01-01", cmp: 0,
	}, {
		l: "9999-12-31", r: "beta", cmp: -1,
	}, {
		// Compare date versions and special tags
		l: "beta", r: "9999-12-31", cmp: 1,
	}, {
		l: "9999-12-31", r: "experimental", cmp: -1,
	}, {
		l: "experimental", r: "9999-12-31", cmp: 1,
		// Compare special tags
	}, {
		l: "beta", r: "experimental", cmp: -1,
	}, {
		l: "experimental", r: "beta", cmp: 1,
	}, {
		l: "beta", r: "beta", cmp: 0,
	}, {
		l: "experimental", r: "experimental", cmp: 0,
	}}
	for i := range tests {
		lv, err := ParseVersion(tests[i].l)
		c.Assert(err, qt.IsNil)
		rv, err := ParseVersion(tests[i].r)
		c.Assert(err, qt.IsNil)
		c.Assert(lv.Compare(rv), qt.Equals, tests[i].cmp)
	}
}
