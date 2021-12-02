package vervet_test

import (
	"sort"
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
			c.Assert(func() { MustParseVersion(tests[i].vs) }, qt.PanicMatches, tests[i].err)
		} else {
			c.Assert(v.Date.Format("2006-01-02"), qt.Equals, tests[i].d)
			c.Assert(v.Stability, qt.Equals, tests[i].stab)
		}
	}
}

func TestVersionStringPanics(t *testing.T) {
	c := qt.New(t)
	var v Version
	c.Assert(func() {
		c.Log(v.String())
	}, qt.PanicMatches, "invalid stability.*")
	var s Stability
	c.Assert(func() {
		c.Log(s.String())
	}, qt.PanicMatches, "invalid stability.*")
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
		l: "9999-12-31", r: "9999-12-31~beta", cmp: 1,
	}, {
		// Compare date versions and special tags
		l: "9999-12-31~beta", r: "9999-12-31", cmp: -1,
	}, {
		l: "9999-12-31", r: "9999-12-31~experimental", cmp: 1,
	}, {
		l: "9999-12-31~experimental", r: "9999-12-31", cmp: -1,
		// Compare special tags
	}, {
		l: "2021-08-01~beta", r: "2021-08-01~experimental", cmp: 1,
	}, {
		l: "2021-08-01~experimental", r: "2021-08-01~beta", cmp: -1,
	}, {
		l: "2021-08-01~beta", r: "2021-08-01~beta", cmp: 0,
	}, {
		l: "2021-08-01~experimental", r: "2021-08-01~experimental", cmp: 0,
	}, {
		l: "2021-08-01~wip", r: "2021-08-01~experimental", cmp: -1,
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
	c.Assert(VersionDateStrings([]Version{
		MustParseVersion("2021-06-01~wip"),
		MustParseVersion("2021-06-01~beta"),
		MustParseVersion("2021-06-10~beta"),
		MustParseVersion("2021-06-10"),
		MustParseVersion("2021-07-12~wip"),
		MustParseVersion("2021-07-12~experimental"),
		MustParseVersion("2021-07-12~beta"),
	}), qt.ContentEquals, []string{"2021-06-01", "2021-06-10", "2021-07-12"})
}

func TestVersionSlice(t *testing.T) {
	type matchTest struct {
		match  string
		result string
		err    string
	}
	tests := []struct {
		versions   VersionSlice
		first      string
		last       string
		matchTests []matchTest
	}{{
		versions: VersionSlice{
			MustParseVersion("2021-07-12~experimental"),
			MustParseVersion("2021-06-01~beta"),
			MustParseVersion("2021-07-12~beta"),
			MustParseVersion("2021-06-10"),
			MustParseVersion("2021-06-01~wip"),
			MustParseVersion("2021-07-12~wip"),
			MustParseVersion("2021-06-10~beta"),
		},
		first: "2021-06-01~wip",
		last:  "2021-07-12~beta",
		matchTests: []matchTest{{
			match:  "2021-06-10",
			result: "2021-06-10",
		}, {
			match:  "2021-06-10~beta",
			result: "2021-06-10",
		}, {
			match:  "2021-06-10~experimental",
			result: "2021-06-10",
		}, {
			match:  "2021-06-11~experimental",
			result: "2021-06-10",
		}, {
			match: "2021-01-01",
			err:   "no matching version",
		}, {
			match:  "2022-01-01",
			result: "2021-06-10",
		}, {
			match:  "2022-01-01~experimental",
			result: "2021-07-12~beta",
		}},
	}, {
		versions: VersionSlice{
			MustParseVersion("2021-06-10~beta"),
		},
		first: "2021-06-10~beta",
		last:  "2021-06-10~beta",
		matchTests: []matchTest{{
			match: "2021-06-10",
			err:   "no matching version",
		}, {
			match:  "2021-06-10~beta",
			result: "2021-06-10~beta",
		}, {
			match:  "2021-06-10~experimental",
			result: "2021-06-10~beta",
		}, {
			match:  "2021-06-11~wip",
			result: "2021-06-10~beta",
		}, {
			match: "2021-01-01",
			err:   "no matching version",
		}, {
			match:  "2022-01-01~wip",
			result: "2021-06-10~beta",
		}},
	}, {
		versions: VersionSlice{
			MustParseVersion("2021-06-10~beta"),
			MustParseVersion("2022-01-10~experimental"),
		},
		first: "2021-06-10~beta",
		last:  "2022-01-10~experimental",
		matchTests: []matchTest{{
			match: "2021-04-10",
			err:   "no matching version",
		}, {
			match:  "2021-08-10~beta",
			result: "2021-06-10~beta",
		}, {
			match:  "2022-02-10~wip",
			result: "2022-01-10~experimental",
		}, {
			match: "2021-01-30",
			err:   "no matching version",
		}},
	}, {
		versions: VersionSlice{
			MustParseVersion("2021-09-06"),
			MustParseVersion("2021-10-06"),
		},
		first: "2021-09-06",
		last:  "2021-10-06",
		matchTests: []matchTest{{
			match:  "2021-10-12~wip",
			result: "2021-10-06",
		}},
	}}
	c := qt.New(t)
	for _, t := range tests {
		sort.Sort(t.versions)
		c.Assert(t.versions[0].String(), qt.Equals, t.first)
		c.Assert(t.versions[len(t.versions)-1].String(), qt.Equals, t.last)
		for _, mt := range t.matchTests {
			match := MustParseVersion(mt.match)
			result, err := t.versions.Resolve(match)
			if err != nil {
				c.Assert(err, qt.ErrorMatches, mt.err)
			} else {
				c.Assert(result.String(), qt.Equals, mt.result)
			}
		}
	}
}

func TestVersionSliceResolveEmpty(t *testing.T) {
	c := qt.New(t)
	_, err := VersionSlice{}.Resolve(MustParseVersion("2021-10-31"))
	c.Assert(err, qt.ErrorMatches, "no matching version")
}
