package simplebuild_test

import (
	"slices"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/internal/simplebuild"
)

func TestGetLatest(t *testing.T) {
	c := qt.New(t)

	c.Run("gets the latest version", func(c *qt.C) {
		before := vervet.MustParseVersion("3000-01-01")
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-01-01"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-03-01"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-02-01"),
				Operation: openapi3.NewOperation(),
			},
		}
		op := vs.GetLatest(before.Date)
		c.Assert(op, qt.Equals, vs[1].Operation)
	})

	c.Run("filters to before given date", func(c *qt.C) {
		before := vervet.MustParseVersion("2024-02-15")
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-01-01"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-03-01"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-02-01"),
				Operation: openapi3.NewOperation(),
			},
		}
		op := vs.GetLatest(before.Date)
		c.Assert(op, qt.Equals, vs[2].Operation)
	})

	c.Run("selects highest stability", func(c *qt.C) {
		before := vervet.MustParseVersion("2024-06-01")
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-01-01"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-02-01~beta"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-03-01"),
				Operation: openapi3.NewOperation(),
			},
			simplebuild.VersionedOp{
				Version:   vervet.MustParseVersion("2024-04-01~experimental"),
				Operation: openapi3.NewOperation(),
			},
		}
		op := vs.GetLatest(before.Date)
		c.Assert(op, qt.Equals, vs[2].Operation)
	})
}

func TestBuild(t *testing.T) {
	c := qt.New(t)

	c.Run("copies paths to output", func(c *qt.C) {
		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   vervet.MustParseVersion("2024-01-01"),
				Operation: openapi3.NewOperation(),
			}},
		}
		output, err := ops.Build()
		c.Assert(err, qt.IsNil)
		c.Assert(output[0].VersionDate, qt.Equals, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		c.Assert(output[0].Doc.Paths["/foo"].Get, qt.IsNotNil)
	})

	c.Run("merges operations from the same version", func(c *qt.C) {
		version := vervet.MustParseVersion("2024-01-01")

		getFoo := openapi3.NewOperation()
		postFoo := openapi3.NewOperation()
		getBar := openapi3.NewOperation()

		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   version,
				Operation: getFoo,
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{{
				Version:   version,
				Operation: postFoo,
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   version,
				Operation: getBar,
			}},
		}
		output, err := ops.Build()
		c.Assert(err, qt.IsNil)
		c.Assert(output[0].VersionDate, qt.Equals, version.Date)
		c.Assert(output[0].Doc.Paths["/foo"].Get, qt.Equals, getFoo)
		c.Assert(output[0].Doc.Paths["/foo"].Post, qt.Equals, postFoo)
		c.Assert(output[0].Doc.Paths["/bar"].Get, qt.Equals, getBar)
	})

	c.Run("generates an output per unique version", func(c *qt.C) {
		versions := []vervet.Version{
			vervet.MustParseVersion("2024-01-01"),
			vervet.MustParseVersion("2024-01-02"),
			vervet.MustParseVersion("2024-01-03"),
		}
		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versions[0],
				Operation: openapi3.NewOperation(),
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versions[1],
				Operation: openapi3.NewOperation(),
			}, {
				Version:   versions[2],
				Operation: openapi3.NewOperation(),
			}},
		}
		output, err := ops.Build()
		c.Assert(err, qt.IsNil)

		inputVersions := make([]time.Time, len(versions))
		for idx, in := range versions {
			inputVersions[idx] = in.Date
		}
		slices.SortFunc(inputVersions, compareDates)

		outputVersions := make([]time.Time, len(output))
		for idx, out := range output {
			outputVersions[idx] = out.VersionDate
		}
		slices.SortFunc(outputVersions, compareDates)

		c.Assert(outputVersions, qt.DeepEquals, inputVersions)
	})

	c.Run("merges distinct operations from previous versions", func(c *qt.C) {
		versionA := vervet.MustParseVersion("2024-01-01")
		versionB := vervet.MustParseVersion("2024-01-02")
		versionC := vervet.MustParseVersion("2024-01-03")

		getFoo := openapi3.NewOperation()
		postFoo := openapi3.NewOperation()
		getBar := openapi3.NewOperation()

		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versionA,
				Operation: getFoo,
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{{
				Version:   versionB,
				Operation: postFoo,
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versionC,
				Operation: getBar,
			}},
		}
		output, err := ops.Build()
		c.Assert(err, qt.IsNil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output[0].VersionDate, qt.Equals, versionA.Date)
		c.Assert(output[0].Doc.Paths["/foo"].Get, qt.Equals, getFoo)
		c.Assert(output[0].Doc.Paths["/foo"].Post, qt.IsNil)
		c.Assert(output[0].Doc.Paths["/bar"], qt.IsNil)

		c.Assert(output[1].VersionDate, qt.Equals, versionB.Date)
		c.Assert(output[1].Doc.Paths["/foo"].Get, qt.Equals, getFoo)
		c.Assert(output[1].Doc.Paths["/foo"].Post, qt.Equals, postFoo)
		c.Assert(output[1].Doc.Paths["/bar"], qt.IsNil)

		c.Assert(output[2].VersionDate, qt.Equals, versionC.Date)
		c.Assert(output[2].Doc.Paths["/foo"].Get, qt.Equals, getFoo)
		c.Assert(output[2].Doc.Paths["/foo"].Post, qt.Equals, postFoo)
		c.Assert(output[2].Doc.Paths["/bar"].Get, qt.Equals, getBar)
	})

	c.Run("resolves operations to latest version with respect to output", func(c *qt.C) {
		versionA := vervet.MustParseVersion("2024-01-01")
		versionB := vervet.MustParseVersion("2024-01-02")
		versionC := vervet.MustParseVersion("2024-01-03")

		getFooOld := openapi3.NewOperation()
		getFooNew := openapi3.NewOperation()
		getBar := openapi3.NewOperation()

		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versionA,
				Operation: getFooOld,
			}, {
				Version:   versionC,
				Operation: getFooNew,
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versionB,
				Operation: getBar,
			}},
		}
		output, err := ops.Build()
		c.Assert(err, qt.IsNil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output[0].VersionDate, qt.Equals, versionA.Date)
		c.Assert(output[0].Doc.Paths["/foo"].Get, qt.Equals, getFooOld)
		c.Assert(output[0].Doc.Paths["/bar"], qt.IsNil)

		c.Assert(output[1].VersionDate, qt.Equals, versionB.Date)
		c.Assert(output[1].Doc.Paths["/foo"].Get, qt.Equals, getFooOld)
		c.Assert(output[1].Doc.Paths["/bar"].Get, qt.Equals, getBar)

		c.Assert(output[2].VersionDate, qt.Equals, versionC.Date)
		c.Assert(output[2].Doc.Paths["/foo"].Get, qt.Equals, getFooNew)
		c.Assert(output[2].Doc.Paths["/bar"].Get, qt.Equals, getBar)
	})

	c.Run("lower stabilities are merged into higher", func(c *qt.C) {
		versionBetaA := vervet.MustParseVersion("2024-01-01~beta")
		versionGA := vervet.MustParseVersion("2024-01-02")
		versionBetaB := vervet.MustParseVersion("2024-01-03~beta")

		getFoo := openapi3.NewOperation()
		postFoo := openapi3.NewOperation()
		getBar := openapi3.NewOperation()

		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versionGA,
				Operation: getFoo,
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{{
				Version:   versionBetaB,
				Operation: postFoo,
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{{
				Version:   versionBetaA,
				Operation: getBar,
			}},
		}
		output, err := ops.Build()
		c.Assert(err, qt.IsNil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output[0].VersionDate, qt.Equals, versionBetaA.Date)
		c.Assert(output[0].Doc.Paths["/foo"], qt.IsNil)
		c.Assert(output[0].Doc.Paths["/bar"].Get, qt.Equals, getBar)

		c.Assert(output[1].VersionDate, qt.Equals, versionGA.Date)
		c.Assert(output[1].Doc.Paths["/foo"].Get, qt.Equals, getFoo)
		c.Assert(output[1].Doc.Paths["/foo"].Post, qt.IsNil)
		c.Assert(output[0].Doc.Paths["/bar"].Get, qt.Equals, getBar)

		c.Assert(output[2].VersionDate, qt.Equals, versionBetaB.Date)
		c.Assert(output[2].Doc.Paths["/foo"].Get, qt.Equals, getFoo)
		c.Assert(output[2].Doc.Paths["/foo"].Post, qt.Equals, postFoo)
		c.Assert(output[2].Doc.Paths["/bar"].Get, qt.Equals, getBar)
	})
}

func compareDocs(a, b simplebuild.VersionedDoc) int {
	return a.VersionDate.Compare(b.VersionDate)
}
func compareDates(a, b time.Time) int {
	return a.Compare(b)
}
