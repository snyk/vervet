package storage

import (
	"context"
	"time"

	qt "github.com/frankban/quicktest"
)

var t0 = time.Date(2021, time.December, 3, 20, 49, 51, 0, time.UTC)

const specPetfood = `{"components":{},"info":{"title":"ServiceA API","version":"0.0.0"},` +
	`"openapi":"3.0.0","paths":{"/petfood":{"get":{"operation":"getTest",` +
	`"responses":{"204":{"description":"An empty response"}},"summary":"Test endpoint"}}}}`

const specAnimals = `{"components":{},"info":{"title":"ServiceA API","version":"0.0.0"},` +
	`"openapi":"3.0.0","paths":{"/animals":{"get":{"operation":"getTest",` +
	`"responses":{"204":{"description":"An empty response"}},"summary":"Test endpoint"}}}}`

func AssertCollateVersion(c *qt.C, s Storage) {
	ctx := context.Background()

	err := s.NotifyVersion(ctx, "petfood", "2021-09-16", []byte(specPetfood), t0)
	c.Assert(err, qt.IsNil)
	err = s.NotifyVersion(ctx, "animals", "2021-09-16", []byte(specAnimals), t0)
	c.Assert(err, qt.IsNil)

	err = s.CollateVersions(ctx, map[string]bool{"petfood": true})
	c.Assert(err, qt.IsNil)

	after, err := s.Version(ctx, "2021-09-16")
	c.Assert(err, qt.IsNil)
	c.Assert(string(after), qt.Equals, specPetfood)
}
