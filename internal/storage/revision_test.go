package storage_test

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/internal/storage"
)

func TestServiceRevisions_ResolveLatestRevision(t *testing.T) {
	c := qt.New(t)

	v20220301_ga := vervet.Version{
		Date:      time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}
	v20220315_beta := vervet.Version{
		Date:      time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityBeta,
	}
	v20220401_ga := vervet.Version{
		Date:      time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}
	v20220401_beta := vervet.Version{
		Date:      time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityBeta,
	}

	ut := storage.NewServiceRevisions()
	ut.Add(storage.ContentRevision{
		Version:   v20220301_ga,
		Digest:    "0301ga_0",
		Timestamp: time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
	})
	ut.Add(storage.ContentRevision{
		Version:   v20220301_ga,
		Digest:    "0301ga_1",
		Timestamp: time.Date(2022, 3, 1, 12, 0, 0, 0, time.UTC),
	})
	ut.Add(storage.ContentRevision{
		Version:   v20220301_ga,
		Digest:    "0301ga_2",
		Timestamp: time.Date(2022, 3, 4, 0, 0, 0, 0, time.UTC),
	})
	ut.Add(storage.ContentRevision{
		Version:   v20220315_beta,
		Digest:    "0315beta_0",
		Timestamp: time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
	})
	ut.Add(storage.ContentRevision{
		Version: v20220401_ga,
		Digest:  "0401ga_0",
	})
	ut.Add(storage.ContentRevision{
		Version: v20220401_beta,
		Digest:  "0401beta_0",
	})

	tcs := []struct {
		version        string
		expectedErr    string
		expectedDigest storage.Digest
	}{
		{
			version:        "2022-03-01",
			expectedDigest: "0301ga_2",
		},
		{
			version:     "2022-03-01~beta",
			expectedErr: "no revision found for resolved version: 2022-03-01~beta",
		},
		{
			version:        "2022-03-05",
			expectedDigest: "0301ga_2",
		},
		{
			version:        "2022-03-15",
			expectedDigest: "0301ga_2",
		},
		{
			version:        "2022-03-15~beta",
			expectedDigest: "0315beta_0",
		},
		{
			version:        "2022-04-01",
			expectedDigest: "0401ga_0",
		},
		{
			version:        "2022-04-05~beta",
			expectedDigest: "0401beta_0",
		},
		{
			version:     "2022-04-05~experimental",
			expectedErr: "no revision found for resolved version: 2022-04-01~experimental",
		},
		{
			version:     "2020-01-01",
			expectedErr: "no matching version",
		},
	}

	for _, tc := range tcs {
		c.Run("version "+tc.version, func(c *qt.C) {
			version, err := vervet.ParseVersion(tc.version)
			c.Check(err, qt.IsNil)

			revision, err := ut.ResolveLatestRevision(version)
			if tc.expectedErr != "" {
				c.Check(err, qt.ErrorMatches, tc.expectedErr)
				return
			}

			c.Check(err, qt.IsNil)
			c.Check(revision.Digest, qt.Equals, tc.expectedDigest)
		})
	}
}
