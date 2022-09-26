package vervet_test

import (
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	. "github.com/snyk/vervet/v4"
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

func TestVersionIndex(t *testing.T) {
	type matchTest struct {
		match       string
		atStability string
		mostStable  string
		err         string
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
			match:       "2021-06-10",
			atStability: "2021-06-10",
			mostStable:  "2021-06-10",
		}, {
			match:       "2021-06-10~beta",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10",
		}, {
			match:       "2021-06-11~beta",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10",
		}, {
			match:       "2021-06-11",
			atStability: "2021-06-10",
			mostStable:  "2021-06-10",
		}, {
			match:       "2021-06-10~experimental",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10",
		}, {
			match:       "2021-06-11~experimental",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10",
		}, {
			match: "2021-01-01",
			err:   "no matching version",
		}, {
			match:       "2022-01-01",
			atStability: "2021-06-10",
			mostStable:  "2021-06-10",
		}, {
			match:       "2022-01-01~experimental",
			atStability: "2021-07-12~experimental",
			mostStable:  "2021-07-12~beta",
		}, {
			match:       "2022-01-01~beta",
			atStability: "2021-07-12~beta",
			mostStable:  "2021-07-12~beta",
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
			match:       "2021-06-10~beta",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10~beta",
		}, {
			match:       "2021-06-10~experimental",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10~beta",
		}, {
			match:       "2021-06-11~wip",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10~beta",
		}, {
			match: "2021-01-01",
			err:   "no matching version",
		}, {
			match:       "2022-01-01~wip",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10~beta",
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
			match:       "2021-08-10~beta",
			atStability: "2021-06-10~beta",
			mostStable:  "2021-06-10~beta",
		}, {
			match:       "2022-02-10~wip",
			atStability: "2022-01-10~experimental",
			mostStable:  "2022-01-10~experimental",
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
			match:       "2021-10-12~wip",
			atStability: "2021-10-06",
			mostStable:  "2021-10-06",
		}},
	}, {
		versions: VersionSlice{
			MustParseVersion("2021-09-06~beta"),
			MustParseVersion("2021-09-07"),
			MustParseVersion("2021-09-08~beta"),
			MustParseVersion("2021-09-09~experimental"),
		},
		first: "2021-09-06~beta",
		last:  "2021-09-09~experimental",
		matchTests: []matchTest{{
			match:       "2021-09-10",
			atStability: "2021-09-07",
			mostStable:  "2021-09-07",
		}},
	}}
	c := qt.New(t)
	for _, t := range tests {
		index := NewVersionIndex(t.versions)
		c.Log(t.versions)
		c.Assert(t.versions[0].String(), qt.Equals, t.first)
		c.Assert(t.versions[len(t.versions)-1].String(), qt.Equals, t.last)
		for _, mt := range t.matchTests {
			c.Log(mt)
			match := MustParseVersion(mt.match)
			atStability, err := index.Resolve(match)
			if err != nil {
				c.Assert(err, qt.ErrorMatches, mt.err)
			} else {
				c.Assert(atStability.String(), qt.Equals, mt.atStability)
			}
			mostStable, err := index.ResolveForBuild(match)
			if err != nil {
				c.Assert(err, qt.ErrorMatches, mt.err)
			} else {
				c.Assert(mostStable.String(), qt.Equals, mt.mostStable)
			}
		}
	}
}

func TestVersionIndexResolveEmpty(t *testing.T) {
	c := qt.New(t)
	vi := NewVersionIndex(VersionSlice{})
	_, err := vi.Resolve(MustParseVersion("2021-10-31"))
	c.Assert(err, qt.ErrorMatches, "no matching version")
	_, err = vi.ResolveForBuild(MustParseVersion("2021-10-31"))
	c.Assert(err, qt.ErrorMatches, "no matching version")
}

func TestDeprecatedBy(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		base, deprecatedBy string
		result             bool
	}{{
		"2021-06-01", "2021-02-01", false,
	}, {
		"2021-06-01", "2021-02-01~experimental", false,
	}, {
		"2021-06-01", "2021-02-01~beta", false,
	}, {
		"2021-06-01", "2022-02-01~beta", false,
	}, {
		"2021-06-01", "2022-02-01~experimental", false,
	}, {
		"2021-06-01", "2021-06-01", false,
	}, {
		"2021-06-01", "2021-06-02", true,
	}, {
		"2021-06-01~experimental", "2021-06-02~beta", true,
	}, {
		"2021-06-01~beta", "2021-06-02", true,
	}, {
		"2021-06-01", "2021-06-01", false,
	}}
	for i, test := range tests {
		c.Logf("test#%d: %#v", i, test)
		base, deprecatedBy := MustParseVersion(test.base), MustParseVersion(test.deprecatedBy)
		c.Assert(base.DeprecatedBy(deprecatedBy), qt.Equals, test.result)
	}
}

func TestDeprecates(t *testing.T) {
	c := qt.New(t)
	versions := NewVersionIndex(VersionSlice{
		MustParseVersion("2021-06-01~experimental"),
		MustParseVersion("2021-06-07~beta"),
		MustParseVersion("2021-07-01"),
		MustParseVersion("2021-08-12~experimental"),
		MustParseVersion("2021-09-16~beta"),
		MustParseVersion("2021-10-31"),
	})
	tests := []struct {
		name         string
		target       Version
		deprecatedBy Version
		isDeprecated bool
		sunset       time.Time
	}{{
		name:         "beta deprecates experimental",
		target:       MustParseVersion("2021-06-01~experimental"),
		deprecatedBy: MustParseVersion("2021-06-07~beta"),
		isDeprecated: true,
		sunset:       time.Date(2021, time.June, 8, 0, 0, 0, 0, time.UTC),
	}, {
		name:         "ga deprecates beta",
		target:       MustParseVersion("2021-06-07~beta"),
		deprecatedBy: MustParseVersion("2021-07-01"),
		isDeprecated: true,
		sunset:       time.Date(2021, time.September, 30, 0, 0, 0, 0, time.UTC),
	}, {
		name:         "ga deprecates ga",
		target:       MustParseVersion("2021-07-01"),
		deprecatedBy: MustParseVersion("2021-10-31"),
		isDeprecated: true,
		sunset:       time.Date(2022, time.April, 30, 0, 0, 0, 0, time.UTC),
	}}
	for i, test := range tests {
		c.Logf("test#%d: %s", i, test.name)
		deprecates, ok := versions.Deprecates(test.target)
		c.Assert(ok, qt.Equals, test.isDeprecated)
		if test.isDeprecated {
			c.Assert(test.deprecatedBy, qt.DeepEquals, deprecates)
			sunset, ok := test.target.Sunset(deprecates)
			c.Assert(ok, qt.IsTrue)
			c.Assert(test.sunset, qt.Equals, sunset)
		}
	}
}

func TestLifecycleAtEnv(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		version   Version
		lifecycle string
		at        string
	}{{
		version:   MustParseVersion("2022-08-16~wip"),
		lifecycle: "sunset",
		at:        "2022-08-16",
	}, {
		version:   MustParseVersion("2022-08-16~experimental"),
		lifecycle: "deprecated",
		at:        "2022-08-16",
	}, {
		version:   MustParseVersion("2022-08-16~experimental"),
		lifecycle: "sunset",
		at:        "2022-11-16",
	}, {
		version:   MustParseVersion("2022-08-16~beta"),
		lifecycle: "released",
		at:        "2022-08-16",
	}, {
		version:   MustParseVersion("2022-08-16~beta"),
		lifecycle: "released",
		at:        "2022-11-16",
	}}
	for _, test := range tests {
		c.Run(fmt.Sprintf("test %v", test.version), func(c *qt.C) {
			c.Setenv("VERVET_LIFECYCLE_AT", test.at)
			lifecycle := test.version.LifecycleAt(time.Time{})
			c.Assert(lifecycle.String(), qt.Equals, test.lifecycle)
		})
	}
}

func TestLifecycleAtDefaultDate(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		version   Version
		lifecycle string
		at        string
	}{{
		version:   MustParseVersion("2022-08-16~wip"),
		lifecycle: "sunset",
	}, {
		version:   MustParseVersion("2022-08-16~experimental"),
		lifecycle: "deprecated",
	}, {
		version:   MustParseVersion("2022-05-16~experimental"),
		lifecycle: "sunset",
	}, {
		version:   MustParseVersion("2022-08-16~beta"),
		lifecycle: "released",
	}, {
		version:   MustParseVersion("2022-08-16~beta"),
		lifecycle: "released",
	}}
	for _, test := range tests {
		c.Run(fmt.Sprintf("test %v", test.version), func(c *qt.C) {
			c.Patch(TimeNow, func() time.Time { return time.Date(2022, time.September, 6, 14, 49, 50, 0, time.UTC) })
			lifecycle := test.version.LifecycleAt(time.Time{})
			c.Assert(lifecycle.String(), qt.Equals, test.lifecycle)
		})
	}
}

func TestLifecycleAtUnreleasedVersionStringPanics(t *testing.T) {
	c := qt.New(t)
	c.Patch(TimeNow, func() time.Time { return time.Date(2022, time.June, 6, 14, 49, 50, 0, time.UTC) })
	version := MustParseVersion("2022-10-16~wip")
	lifecycle := version.LifecycleAt(time.Time{})
	c.Assert(func() {
		c.Log(lifecycle.String())
	}, qt.PanicMatches, "invalid lifecycle.*")
}
