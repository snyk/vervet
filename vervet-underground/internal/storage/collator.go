package storage

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"
)

// Collator is an aggregate of service specs and uniqueVersions scraped by VU. It is responsible for collating uniqueVersions and
// specs from all services VU manages.
// This is the top level resource all storage classes should use for producing collated data.
type Collator struct {
	// revisions is a map of service name to the service's revisions.
	revisions map[string]*ServiceRevisions
	// uniqueVersions is API versions of all services.
	uniqueVersions vervet.VersionSlice
	// excludePatterns identifies elements to be removed from the collated OpenAPI output.
	excludePatterns vervet.ExcludePatterns
}

// NewCollator returns a new Collator instance.
func NewCollator() *Collator {
	return NewCollatorExcludePatterns(vervet.ExcludePatterns{})
}

// NewCollatorExcludePatterns returns a new Collator instance with patterns for
// excluding elements from the output.
func NewCollatorExcludePatterns(excludePatterns vervet.ExcludePatterns) *Collator {
	return &Collator{
		revisions:       make(map[string]*ServiceRevisions),
		uniqueVersions:  nil,
		excludePatterns: excludePatterns,
	}
}

// Add a new service and revision to the Collator.
func (c *Collator) Add(service string, revision ContentRevision) {
	// Track service and its revision
	if _, ok := c.revisions[service]; !ok {
		c.revisions[service] = NewServiceRevisions()
	}
	c.revisions[service].Add(revision)

	// Track versions
	version := revision.Version
	var found bool
	for _, v := range c.uniqueVersions {
		if version == v {
			found = true
			break
		}
	}
	if !found {
		c.uniqueVersions = append(c.uniqueVersions, version)
	}
}

// Collate processes added service revisions to collate unified versions and OpenAPI specs for each version.
func (c *Collator) Collate() (vervet.VersionSlice, map[vervet.Version]openapi3.T, error) {
	specs := make(map[vervet.Version]openapi3.T)
	sort.Sort(c.uniqueVersions)

	for _, version := range c.uniqueVersions {
		revisions := make([]ContentRevision, 0)
		for service, serviceRevisions := range c.revisions {
			rev, err := serviceRevisions.ResolveLatestRevision(version)
			if err != nil {
				// don't halt execution if we can't resolve version for this service - it is possible for a service to not have this version available.
				log.Trace().Err(err).Msgf("could not resolve version %s for service %s", version, service)
				continue
			}
			revisions = append(revisions, rev)
		}

		if len(revisions) > 0 {
			spec, err := mergeRevisions(revisions)
			if err != nil {
				log.Error().Err(err).Msgf("could not merge revision for version %s", version)
				collatorMergeError.WithLabelValues(version.String()).Inc()
				return nil, nil, err
			}
			if err := vervet.RemoveElements(spec, c.excludePatterns); err != nil {
				log.Error().Err(err).Msgf("could not merge revision for version %s", version)
				collatorMergeError.WithLabelValues(version.String()).Inc()
				return nil, nil, err
			}
			specs[version] = *spec
		}
	}

	return c.uniqueVersions, specs, nil
}

func mergeRevisions(revisions []ContentRevision) (*openapi3.T, error) {
	collator := vervet.NewCollator()
	for _, revision := range revisions {
		loader := openapi3.NewLoader()
		// JSON will deserialize here correctly
		src, err := loader.LoadFromData(revision.Blob)
		if err != nil {
			return nil, fmt.Errorf("could not load revision %s-%s-%s: %w", revision.Service, revision.Version, revision.Digest, err)
		}

		rv := &vervet.ResourceVersion{
			Document: vervet.NewResolvedDocument(src, &url.URL{
				Scheme: "vu",
				Host:   revision.Service,
				Path:   revision.Version.String() + "@" + string(revision.Digest),
			}),
			Name:    revision.Service,
			Version: revision.Version,
		}
		if err := collator.Collate(rv); err != nil {
			return nil, fmt.Errorf("could not collate revision %s-%s-%s: %w", revision.Service, revision.Version, revision.Digest, err)
		}
	}
	return collator.Result(), nil
}
