package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	. "github.com/snyk/vervet"
)

func TestParseVersion(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		vs   string
		d    string
		stab Stability
		err  string
	}{{
		vs:   "2021-01-01",
		d:    "2021-01-01",
		stab: StabilityGA,
	}, {
		vs:   "2021-02-02~beta",
		d:    "2021-02-02",
		stab: StabilityBeta,
	}, {
		vs:   "2021-03-03~experimental",
		d:    "2021-03-03",
		stab: StabilityExperimental,
	}, {
		vs:  "unknown",
		err: `invalid version "unknown"`,
	}}
	for i := range tests {
		c.Logf("parse version %q", tests[i].vs)
		v, err := ParseVersion(tests[i].vs)
		if tests[i].err != "" {
			c.Assert(err, qt.ErrorMatches, tests[i].err)
		} else {
			c.Assert(v.Date.Format("2006-01-02"), qt.Equals, tests[i].d)
			c.Assert(v.Stability, qt.Equals, tests[i].stab)
		}
	}
}

func mustParseVersion(s string) *Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
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
		l: "9999-12-31", r: "9999-12-31~beta", cmp: -1,
	}, {
		// Compare date versions and special tags
		l: "9999-12-31~beta", r: "9999-12-31", cmp: 1,
	}, {
		l: "9999-12-31", r: "9999-12-31~experimental", cmp: -1,
	}, {
		l: "9999-12-31~experimental", r: "9999-12-31", cmp: 1,
		// Compare special tags
	}, {
		l: "2021-08-01~beta", r: "2021-08-01~experimental", cmp: -1,
	}, {
		l: "2021-08-01~experimental", r: "2021-08-01~beta", cmp: 1,
	}, {
		l: "2021-08-01~beta", r: "2021-08-01~beta", cmp: 0,
	}, {
		l: "2021-08-01~experimental", r: "2021-08-01~experimental", cmp: 0,
	}, {
		l: "2021-08-01~wip", r: "2021-08-01~experimental", cmp: 1,
	}}
	for i := range tests {
		c.Logf("test %d %#v", i, tests[i])
		lv, err := ParseVersion(tests[i].l)
		c.Assert(err, qt.IsNil)
		rv, err := ParseVersion(tests[i].r)
		c.Assert(err, qt.IsNil)
		c.Assert(lv.Compare(rv), qt.Equals, tests[i].cmp)
	}
}

func TestVersionDateStrings(t *testing.T) {
	c := qt.New(t)
	c.Assert(VersionDateStrings([]*Version{
		mustParseVersion("2021-06-01~wip"),
		mustParseVersion("2021-06-01~beta"),
		mustParseVersion("2021-06-10~beta"),
		mustParseVersion("2021-06-10"),
		mustParseVersion("2021-07-12~wip"),
		mustParseVersion("2021-07-12~experimental"),
		mustParseVersion("2021-07-12~beta"),
	}), qt.ContentEquals, []string{"2021-06-01", "2021-06-10", "2021-07-12"})
}
