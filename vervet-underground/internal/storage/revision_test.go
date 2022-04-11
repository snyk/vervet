package storage_test

import (
	qt "github.com/frankban/quicktest"
	"github.com/snyk/vervet"
	"testing"
	"time"
	"vervet-underground/internal/storage"
)

func TestServiceRevisions_ResolveLatestRevision(t *testing.T) {
	c := qt.New(t)

	v20220301_ga := vervet.Version{
		Date:      time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}
	v20220401_ga := vervet.Version{
		Date:      time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	ut := storage.NewServiceRevisions()
	ut.Add(storage.ContentRevision{Version: v20220301_ga, Digest: "1", Timestamp: time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC)})
	ut.Add(storage.ContentRevision{Version: v20220301_ga, Digest: "2", Timestamp: time.Date(2022, 3, 1, 12, 0, 0, 0, time.UTC)})
	ut.Add(storage.ContentRevision{Version: v20220401_ga, Digest: "3"})

	tcs := []struct {
		version        string
		expectedErr    string
		expectedDigest storage.Digest
	}{
		{
			version:        "2022-03-01",
			expectedDigest: "2",
		},
		{
			version:        "2022-03-01~beta",
			expectedDigest: "2",
		},
		{
			version:        "2022-03-05",
			expectedDigest: "2",
		},
		{
			version:        "2022-04-01",
			expectedDigest: "3",
		},
		{
			version:        "2022-04-05~beta",
			expectedDigest: "3",
		},
		{
			version:     "2020-01-01",
			expectedErr: "no matching version",
		},
	}

	for _, tc := range tcs {
		c.Run("version "+tc.version, func(c *qt.C) {
			version, err := vervet.ParseVersion(tc.version)
			c.Assert(err, qt.IsNil)

			revision, err := ut.ResolveLatestRevision(*version)
			if tc.expectedErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.expectedErr)
				return
			}

			c.Assert(err, qt.IsNil)
			c.Assert(revision.Digest, qt.Equals, tc.expectedDigest)
		})
	}

}
