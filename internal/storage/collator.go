package storage

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"

	"github.com/snyk/vervet/v8"
)

// Collator is an aggregate of service specs and uniqueVersions scraped by VU. It
// is responsible for collating uniqueVersions and specs from all services VU
// manages.
// This is the top level resource all storage classes should use for producing collated data.
type Collator struct {
	// revisions is a map of service name to the service's revisions.
	revisions map[string]*ServiceRevisions
	// uniqueVersions is API versions of all services.
	uniqueVersions vervet.VersionSlice

	// excludePatterns identifies elements to be removed from the collated OpenAPI output.
	excludePatterns vervet.ExcludePatterns
	// OpenAPI3 document overlay applied to collated result
	overlay string
}

// NewCollator returns a new Collator instance.
func NewCollator(options ...CollatorOption) (*Collator, error) {
	coll := &Collator{
		revisions: map[string]*ServiceRevisions{},
	}
	for i := range options {
		err := options[i](coll)
		if err != nil {
			return nil, err
		}
	}
	return coll, nil
}

// CollatorOption defines an optional setting when creating a new Collator.
type CollatorOption func(*Collator) error

// CollatorExcludePattern is a CollatorOption which specifies an exclude
// pattern to apply when collating OpenAPI document objects.
func CollatorExcludePattern(excludePatterns vervet.ExcludePatterns) CollatorOption {
	return func(c *Collator) error {
		c.excludePatterns = excludePatterns
		return nil
	}
}

// CollatorOverlay is a CollatorOption which specifies an OpenAPI document
// overlay to apply to the collated result. Top-level fields in the overlay
// replaces top-level fields in the collated result. Paths are merged.
func CollatorOverlay(overlay string) CollatorOption {
	return func(c *Collator) error {
		// Load the overlays early to validate the config.
		//
		// Why don't we reuse these parsed overlays in the long-lived
		// collator instance for successive merges? It may cause strange
		// effects that would be hard to debug and troubleshoot.
		// kin-openapi/openapi3 structures have private state references to
		// the loader which loaded them, which "come along for the ride" in
		// a merge. Re-parsing on each merge is safe and known to be
		// reliable.
		l := openapi3.NewLoader()
		_, err := l.LoadFromData([]byte(overlay))
		if err != nil {
			return err
		}
		c.overlay = overlay
		return nil
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
func (c *Collator) Collate() (map[vervet.Version]openapi3.T, error) {
	specs := make(map[vervet.Version]openapi3.T)
	sort.Sort(c.uniqueVersions)

	for _, version := range c.uniqueVersions {
		revisions := make(ContentRevisions, 0)
		for service, serviceRevisions := range c.revisions {
			rev, err := serviceRevisions.ResolveLatestRevision(version)
			if err != nil {
				// don't halt execution if we can't resolve version for this
				// service - it is possible for a service to not have this
				// version available.
				log.Trace().Err(err).Msgf("could not resolve version %s for service %s", version, service)
				continue
			}
			revisions = append(revisions, rev)
		}

		sort.Sort(revisions)
		if len(revisions) > 0 {
			spec, err := mergeRevisions(revisions)
			if err != nil {
				log.Error().Err(err).Msgf("could not merge revision for version %s", version)
				collatorMergeError.WithLabelValues(version.String()).Inc()
				return nil, err
			}
			if err := vervet.RemoveElements(spec, c.excludePatterns); err != nil {
				log.Error().Err(err).Msgf("could not merge revision for version %s", version)
				collatorMergeError.WithLabelValues(version.String()).Inc()
				return nil, err
			}

			// Overrides sunset header documentation until we can provide a more definitive fix
			overrideSunsetHeader(spec)

			if err := c.applyOverlay(spec); err != nil {
				log.Error().Err(err).Msgf("failed to merge overlay for version %s", version)
				collatorMergeError.WithLabelValues(version.String()).Inc()
				return nil, err
			}
			specs[version] = *spec
		}
	}

	return specs, nil
}

func overrideSunsetHeader(doc *openapi3.T) {
	const headerDescription = "A header containing the date of when the underlying endpoint will be removed. " +
		"This header is only present if the endpoint has been deprecated. " +
		"For information purposes only. " +
		"Returned as a date in the format: YYYY-MM-DD"
	const example = "2021-08-02"
	const schemaFormat = "date"

	for _, path := range doc.Paths.Map() {
		for _, operation := range path.Operations() {
			for _, responses := range operation.Responses.Map() {
				if responses.Value == nil {
					continue
				}

				if sunsetHeader, ok := responses.Value.Headers["sunset"]; ok {
					if sunsetHeader.Value == nil {
						continue
					}
					sunsetHeader.Value.Description = headerDescription
					sunsetHeader.Value.Example = example
					if sunsetHeader.Value.Schema.Value == nil {
						continue
					}
					sunsetHeader.Value.Schema.Value.Format = schemaFormat
				}
			}
		}
	}
	if doc.Components != nil {
		if sunsetHeader, ok := doc.Components.Headers["SunsetHeader"]; ok {
			if sunsetHeader.Value == nil {
				return
			}
			sunsetHeader.Value.Description = headerDescription
		}
	}
}

func mergeRevisions(revisions ContentRevisions) (*openapi3.T, error) {
	collator := vervet.NewCollator(vervet.StrictTags(false), vervet.UseFirstRoute(true))
	var haveOpenAPI, haveOpenAPIVersion bool
	for _, revision := range revisions {
		loader := openapi3.NewLoader()
		// JSON will deserialize here correctly
		src, err := loader.LoadFromData(revision.Blob)
		if err != nil {
			return nil, fmt.Errorf(
				"could not load revision %s-%s-%s: %w",
				revision.Service,
				revision.Version,
				revision.Digest,
				err,
			)
		}

		// Each service will declare their own /openapi paths. Collate only the
		// first instances of these paths and remove them from subsequent specs
		// to prevent failing the collate on conflicting paths.
		if haveOpenAPI {
			src.Paths.Delete("/openapi")
		} else if pathItem := src.Paths.Value("/openapi"); pathItem != nil {
			haveOpenAPI = true
		}
		if haveOpenAPIVersion {
			src.Paths.Delete("/openapi/{version}")
		} else if pathItem := src.Paths.Value("/openapi/{version}"); pathItem != nil {
			haveOpenAPIVersion = true
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
			return nil, fmt.Errorf("could not collate revision %s-%s-%s: %w",
				revision.Service,
				revision.Version,
				revision.Digest,
				err,
			)
		}
	}
	return collator.Result(), nil
}

func (c *Collator) applyOverlay(spec *openapi3.T) error {
	l := openapi3.NewLoader()
	overlayDoc, err := l.LoadFromData([]byte(c.overlay))
	if err != nil {
		return err
	}
	return vervet.Merge(spec, overlayDoc, true)
}
