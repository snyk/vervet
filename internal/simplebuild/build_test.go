package simplebuild_test

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/simplebuild"
)

func TestGetLatest(t *testing.T) {
	c := qt.New(t)

	c.Run("gets the latest version", func(c *qt.C) {
		before := vervet.MustParseVersion("3000-01-01")
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-03-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-02-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
		}
		op := vs.GetLatest(before.Date)
		c.Assert(op, qt.Equals, vs[1].Operation)
	})

	c.Run("filters to before given date", func(c *qt.C) {
		before := vervet.MustParseVersion("2024-02-15")
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-03-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-02-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
		}
		op := vs.GetLatest(before.Date)
		c.Assert(op, qt.Equals, vs[2].Operation)
	})

	c.Run("selects highest stability", func(c *qt.C) {
		before := vervet.MustParseVersion("2024-06-01")
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-02-01~beta"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-03-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-04-01~experimental"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)
		c.Assert(output[0].VersionDate, qt.Equals, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		c.Assert(output[0].Doc.Paths.Value("/foo").Get, qt.IsNotNil)
	})

	c.Run("copies servers to output", func(c *qt.C) {
		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			}},
		}
		expectedServer := &openapi3.Server{
			URL: "https://example.com",
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), openapi3.Servers{
			expectedServer,
		})
		c.Assert(output[0].VersionDate, qt.Equals, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		c.Assert(output[0].Doc.Paths.Value("/foo").Get, qt.IsNotNil)
		c.Assert(output[0].Doc.Servers, qt.HasLen, 1)
		c.Assert(output[0].Doc.Servers[0], qt.Equals, expectedServer)
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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      version,
				Operation:    getFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      version,
				Operation:    postFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      version,
				Operation:    getBar,
				ResourceName: "bar",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)
		c.Assert(output[0].VersionDate, qt.Equals, version.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[0].Doc.Paths.Value("/foo").Post, qt.Equals, postFoo)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versions[0],
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versions[1],
				Operation:    openapi3.NewOperation(),
				ResourceName: "bar",
			}, simplebuild.VersionedOp{
				Version:      versions[2],
				Operation:    openapi3.NewOperation(),
				ResourceName: "bar",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)

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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionA,
				Operation:    getFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionB,
				Operation:    postFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionC,
				Operation:    getBar,
				ResourceName: "bar",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output[0].VersionDate, qt.Equals, versionA.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[0].Doc.Paths.Value("/foo").Post, qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar"), qt.IsNil)

		c.Assert(output[1].VersionDate, qt.Equals, versionB.Date)
		c.Assert(output[1].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[1].Doc.Paths.Value("/foo").Post, qt.Equals, postFoo)
		c.Assert(output[1].Doc.Paths.Value("/bar"), qt.IsNil)

		c.Assert(output[2].VersionDate, qt.Equals, versionC.Date)
		c.Assert(output[2].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[2].Doc.Paths.Value("/foo").Post, qt.Equals, postFoo)
		c.Assert(output[2].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionA,
				Operation:    getFooOld,
				ResourceName: "foo",
			}, simplebuild.VersionedOp{
				Version:   versionC,
				Operation: getFooNew,
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionB,
				Operation:    getBar,
				ResourceName: "bar",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output[0].VersionDate, qt.Equals, versionA.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo").Get, qt.Equals, getFooOld)
		c.Assert(output[0].Doc.Paths.Value("/bar"), qt.IsNil)

		c.Assert(output[1].VersionDate, qt.Equals, versionB.Date)
		c.Assert(output[1].Doc.Paths.Value("/foo").Get, qt.Equals, getFooOld)
		c.Assert(output[1].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)

		c.Assert(output[2].VersionDate, qt.Equals, versionC.Date)
		c.Assert(output[2].Doc.Paths.Value("/foo").Get, qt.Equals, getFooNew)
		c.Assert(output[2].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
	})

	c.Run("does not generate versions before pivot date", func(c *qt.C) {
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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionA,
				Operation:    getFooOld,
				ResourceName: "foo",
			}, simplebuild.VersionedOp{
				Version:   versionC,
				Operation: getFooNew,
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionB,
				Operation:    getBar,
				ResourceName: "bar",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-02"), nil)

		slices.SortFunc(output, compareDocs)

		c.Assert(len(output), qt.Equals, 2)

		c.Assert(output[0].VersionDate, qt.Equals, versionB.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo").Get, qt.Equals, getFooOld)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)

		c.Assert(output[1].VersionDate, qt.Equals, versionC.Date)
		c.Assert(output[1].Doc.Paths.Value("/foo").Get, qt.Equals, getFooNew)
		c.Assert(output[1].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
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
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionGA,
				Operation:    getFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionBetaB,
				Operation:    postFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionBetaA,
				Operation:    getBar,
				ResourceName: "bar",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output[0].VersionDate, qt.Equals, versionBetaA.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo"), qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)

		c.Assert(output[1].VersionDate, qt.Equals, versionGA.Date)
		c.Assert(output[1].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[1].Doc.Paths.Value("/foo").Post, qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)

		c.Assert(output[2].VersionDate, qt.Equals, versionBetaB.Date)
		c.Assert(output[2].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[2].Doc.Paths.Value("/foo").Post, qt.Equals, postFoo)
		c.Assert(output[2].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
	})

	c.Run("experimental version are filtered out", func(c *qt.C) {
		versionBetaA := vervet.MustParseVersion("2024-01-01~beta")
		versionGA := vervet.MustParseVersion("2024-01-02")
		versionBetaB := vervet.MustParseVersion("2024-01-03~beta")
		versionExperimental := vervet.MustParseVersion("2024-01-04~experimental")
		versionExperimentalBeforePivotDate := vervet.MustParseVersion("2023-01-01~experimental")

		getFoo := openapi3.NewOperation()
		postFoo := openapi3.NewOperation()
		getBar := openapi3.NewOperation()

		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionGA,
				Operation:    getFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionBetaB,
				Operation:    postFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionBetaA,
				Operation:    getBar,
				ResourceName: "bar",
			}},
			simplebuild.OpKey{
				Path:   "/experimental-path",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionExperimental,
				Operation:    openapi3.NewOperation(),
				ResourceName: "experimental-path",
			}},
			simplebuild.OpKey{
				Path:   "/experimental-path-before-pivot-date",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionExperimentalBeforePivotDate,
				Operation:    openapi3.NewOperation(),
				ResourceName: "experimental-path-before-pivot-date",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output, qt.HasLen, 3)

		c.Assert(output[0].VersionDate, qt.Equals, versionBetaA.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo"), qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
		c.Assert(output[0].Doc.Paths.Value("/experimental-path-before-pivot-date"), qt.IsNil)

		c.Assert(output[1].VersionDate, qt.Equals, versionGA.Date)
		c.Assert(output[1].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[1].Doc.Paths.Value("/foo").Post, qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)

		c.Assert(output[2].VersionDate, qt.Equals, versionBetaB.Date)
		c.Assert(output[2].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[2].Doc.Paths.Value("/foo").Post, qt.Equals, postFoo)
		c.Assert(output[2].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
		c.Assert(output[2].Doc.Paths.Value("/experimental-path-before-pivot-date"), qt.IsNil)
		c.Assert(output[2].Doc.Paths.Value("/experimental-path"), qt.IsNil)
	})
	c.Run("wip version are filtered out", func(c *qt.C) {
		versionBetaA := vervet.MustParseVersion("2024-01-01~beta")
		versionGA := vervet.MustParseVersion("2024-01-02")
		versionBetaB := vervet.MustParseVersion("2024-01-03~beta")
		versionWip := vervet.MustParseVersion("2024-01-04~wip")
		versionWipBeforePivotDate := vervet.MustParseVersion("2023-01-01~wip")

		getFoo := openapi3.NewOperation()
		postFoo := openapi3.NewOperation()
		getBar := openapi3.NewOperation()

		ops := simplebuild.Operations{
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionGA,
				Operation:    getFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/foo",
				Method: "POST",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionBetaB,
				Operation:    postFoo,
				ResourceName: "foo",
			}},
			simplebuild.OpKey{
				Path:   "/bar",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionBetaA,
				Operation:    getBar,
				ResourceName: "bar",
			}},
			simplebuild.OpKey{
				Path:   "/wip-path",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionWip,
				Operation:    openapi3.NewOperation(),
				ResourceName: "wip-path",
			}},
			simplebuild.OpKey{
				Path:   "/wip-path-before-pivot-date",
				Method: "GET",
			}: simplebuild.VersionSet{simplebuild.VersionedOp{
				Version:      versionWipBeforePivotDate,
				Operation:    openapi3.NewOperation(),
				ResourceName: "wip-path-before-pivot-date",
			}},
		}
		output := ops.Build(vervet.MustParseVersion("2024-01-01"), nil)

		slices.SortFunc(output, compareDocs)

		c.Assert(output, qt.HasLen, 3)

		c.Assert(output[0].VersionDate, qt.Equals, versionBetaA.Date)
		c.Assert(output[0].Doc.Paths.Value("/foo"), qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
		c.Assert(output[0].Doc.Paths.Value("/wip-path-before-pivot-date"), qt.IsNil)

		c.Assert(output[1].VersionDate, qt.Equals, versionGA.Date)
		c.Assert(output[1].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[1].Doc.Paths.Value("/foo").Post, qt.IsNil)
		c.Assert(output[0].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)

		c.Assert(output[2].VersionDate, qt.Equals, versionBetaB.Date)
		c.Assert(output[2].Doc.Paths.Value("/foo").Get, qt.Equals, getFoo)
		c.Assert(output[2].Doc.Paths.Value("/foo").Post, qt.Equals, postFoo)
		c.Assert(output[2].Doc.Paths.Value("/bar").Get, qt.Equals, getBar)
		c.Assert(output[2].Doc.Paths.Value("/wip-path-before-pivot-date"), qt.IsNil)
		c.Assert(output[2].Doc.Paths.Value("/wip-path"), qt.IsNil)
	})
}

func TestAnnotate(t *testing.T) {
	c := qt.New(t)

	c.Run("adds version dates and resource name to operations", func(c *qt.C) {
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-02-01~beta"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-03-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "bar",
			},
		}
		vs.Annotate()
		c.Assert(vs[0].Operation.Extensions[vervet.ExtSnykApiVersion], qt.Equals, "2024-01-01")
		c.Assert(vs[0].Operation.Extensions[vervet.ExtSnykApiResource], qt.Equals, "foo")
		c.Assert(vs[1].Operation.Extensions[vervet.ExtSnykApiVersion], qt.Equals, "2024-02-01~beta")
		c.Assert(vs[1].Operation.Extensions[vervet.ExtSnykApiResource], qt.Equals, "foo")
		c.Assert(vs[2].Operation.Extensions[vervet.ExtSnykApiVersion], qt.Equals, "2024-03-01")
		c.Assert(vs[2].Operation.Extensions[vervet.ExtSnykApiResource], qt.Equals, "bar")

		// Check stability level for beta version
		c.Assert(vs[1].Operation.Extensions[vervet.ExtApiStabilityLevel], qt.Equals, "beta")
		c.Assert(vs[1].Operation.Extensions[vervet.ExtSnykApiStability], qt.Equals, "beta")
	})

	c.Run("adds a list of all other versions", func(c *qt.C) {
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-02-01~beta"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-03-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "bar",
			},
		}
		vs.Annotate()
		c.Assert(
			vs[0].Operation.Extensions[vervet.ExtSnykApiReleases],
			qt.DeepEquals,
			[]string{"2024-01-01", "2024-02-01~beta", "2024-03-01"},
		)
		c.Assert(
			vs[1].Operation.Extensions[vervet.ExtSnykApiReleases],
			qt.DeepEquals,
			[]string{"2024-01-01", "2024-02-01~beta", "2024-03-01"},
		)
		c.Assert(
			vs[2].Operation.Extensions[vervet.ExtSnykApiReleases],
			qt.DeepEquals,
			[]string{"2024-01-01", "2024-02-01~beta", "2024-03-01"},
		)

		c.Assert(vs[1].Operation.Extensions[vervet.ExtApiStabilityLevel], qt.Equals, "beta")
		c.Assert(vs[1].Operation.Extensions[vervet.ExtSnykApiStability], qt.Equals, "beta")
	})

	c.Run("adds deprecation annotations on older versions", func(c *qt.C) {
		vs := simplebuild.VersionSet{
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-01-01~beta"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-02-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "foo",
			},
			simplebuild.VersionedOp{
				Version:      vervet.MustParseVersion("2024-03-01"),
				Operation:    openapi3.NewOperation(),
				ResourceName: "bar",
			},
		}
		vs.Annotate()
		c.Assert(vs[0].Operation.Extensions[vervet.ExtSnykDeprecatedBy], qt.Equals, "2024-02-01")
		// beta sunsets after 91 days
		c.Assert(vs[0].Operation.Extensions[vervet.ExtSnykSunsetEligible], qt.Equals, "2024-05-02")
		c.Assert(vs[1].Operation.Extensions[vervet.ExtSnykDeprecatedBy], qt.Equals, "2024-03-01")
		// ga sunsets after 181 days
		c.Assert(vs[1].Operation.Extensions[vervet.ExtSnykSunsetEligible], qt.Equals, "2024-08-29")
		c.Assert(vs[2].Operation.Extensions[vervet.ExtSnykDeprecatedBy], qt.IsNil)
		c.Assert(vs[2].Operation.Extensions[vervet.ExtSnykSunsetEligible], qt.IsNil)

		c.Assert(vs[0].Operation.Extensions["x-stability-level"], qt.Equals, "beta")
		c.Assert(vs[0].Operation.Extensions["x-snyk-api-stability"], qt.Equals, "beta")
	})
}

func TestCheckBreakingChanges(t *testing.T) {
	c := qt.New(t)

	c.Run("detects breaking changes between versions", func(c *qt.C) {
		paths1 := openapi3.Paths{}
		paths1.Set("/foo", &openapi3.PathItem{
			Get: &openapi3.Operation{},
		})

		paths2 := openapi3.Paths{}
		paths2.Set("/foo", &openapi3.PathItem{
			Post: &openapi3.Operation{},
		})

		doc1 := &openapi3.T{
			Paths: &paths1,
		}
		doc2 := &openapi3.T{
			Paths: &paths2,
		}

		docs := simplebuild.DocSet{
			{
				VersionDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Doc:         doc1,
			},
			{
				VersionDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				Doc:         doc2,
			},
		}

		err := simplebuild.CheckBreakingChanges(docs)
		c.Assert(err, qt.IsNil, qt.Commentf("expected breaking change to be detected"))
	})

	c.Run("no breaking changes between versions", func(c *qt.C) {
		paths1 := openapi3.Paths{}
		paths1.Set("/foo", &openapi3.PathItem{
			Get: &openapi3.Operation{},
		})

		paths2 := openapi3.Paths{}
		paths2.Set("/foo", &openapi3.PathItem{
			Get: &openapi3.Operation{},
		})

		doc1 := &openapi3.T{
			Paths: &paths1,
		}
		doc2 := &openapi3.T{
			Paths: &paths2,
		}

		docs := simplebuild.DocSet{
			{
				VersionDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Doc:         doc1,
			},
			{
				VersionDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				Doc:         doc2,
			},
		}

		err := simplebuild.CheckBreakingChanges(docs)
		c.Assert(err, qt.Not(qt.IsNil), qt.Commentf("expected no breaking changes"))
	})
}

func TestCheckSingleVersionResource(t *testing.T) {
	c := qt.New(t)

	c.Run("no error when version is before or equal to the latest version", func(c *qt.C) {
		paths := []string{
			"internal/api/hidden/resources/apps/2023-07-31/spec.yaml",
		}
		latestVersion := vervet.MustParseVersion("2024-01-01")

		err := simplebuild.CheckSingleVersionResourceToBeBeforeLatestVersion(paths, latestVersion)
		c.Assert(err, qt.IsNil)
	})

	c.Run("error when version is after the latest version", func(c *qt.C) {
		paths := []string{
			"internal/api/hidden/resources/apps/2025-07-31/spec.yaml",
		}
		latestVersion := vervet.MustParseVersion("2024-01-01")

		err := simplebuild.CheckSingleVersionResourceToBeBeforeLatestVersion(paths, latestVersion)
		c.Assert(err, qt.ErrorMatches, "version .* is after the last released version .*")
	})

	c.Run("no error when multiple versions are present", func(c *qt.C) {
		paths := []string{
			"internal/api/hidden/resources/apps/2023-07-31/spec.yaml",
			"internal/api/hidden/resources/apps/2024-01-01/spec.yaml",
		}
		latestVersion := vervet.MustParseVersion("2024-01-01")

		err := simplebuild.CheckSingleVersionResourceToBeBeforeLatestVersion(paths, latestVersion)
		c.Assert(err, qt.IsNil)
	})

	c.Run("handles version parsing error gracefully", func(c *qt.C) {
		paths := []string{
			"internal/api/hidden/resources/apps/invalid-version/spec.yaml",
		}
		latestVersion := vervet.MustParseVersion("2024-01-01")

		err := simplebuild.CheckSingleVersionResourceToBeBeforeLatestVersion(paths, latestVersion)
		c.Assert(err, qt.ErrorMatches, "invalid version .*")
	})
}

func compareDocs(a, b simplebuild.VersionedDoc) int {
	return a.VersionDate.Compare(b.VersionDate)
}
func compareDates(a, b time.Time) int {
	return a.Compare(b)
}

func TestMapStabilityLevel(t *testing.T) {
	tests := []struct {
		name string
		args vervet.Stability
		want string
	}{
		{
			name: "stable",
			args: vervet.StabilityGA,
			want: "stable",
		},
		{
			name: "beta",
			args: vervet.StabilityBeta,
			want: "beta",
		},
		{
			name: "defaults to blank",
			args: vervet.StabilityExperimental,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := simplebuild.MapStabilityLevel(tt.args); got != tt.want {
				t.Errorf("MapStabilityLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildSkipsVersionCheckWhenFetchFails(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	// Create a temporary working directory.
	tempDir := t.TempDir()

	origWD, err := os.Getwd()
	c.Assert(err, qt.IsNil)
	err = os.Chdir(tempDir)
	c.Assert(err, qt.IsNil)

	defer func() {
		if err := os.Chdir(origWD); err != nil {
			t.Errorf("failed to change directory back: %v", err)
		}
	}()

	// Create an empty CODEOWNERS file to satisfy codeowners.FromFile.
	err = os.WriteFile("CODEOWNERS", []byte(""), 0644)
	c.Assert(err, qt.IsNil)

	dummyOutput := &config.Output{
		Paths: []string{tempDir},
	}
	dummyAPI := &config.API{
		Name:      "dummy-api",
		Output:    dummyOutput,
		Resources: []*config.ResourceSet{},
	}
	dummyProject := &config.Project{
		APIs: config.APIs{
			"dummy-api": dummyAPI,
		},
	}

	startDate := vervet.MustParseVersion("2020-01-01")

	// Versioning URL that will fail
	failingURL := "http://localhost:0"

	// Even though fetching the latest version fails, the build should continue and return nil.
	err = simplebuild.Build(ctx, dummyProject, startDate, failingURL, false)
	c.Assert(err, qt.IsNil)
}
