package storage

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
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
			applyOverlay(spec)
			specs[version] = *spec
		}
	}

	return c.uniqueVersions, specs, nil
}

func mergeRevisions(revisions []ContentRevision) (*openapi3.T, error) {
	collator := vervet.NewCollator()
	var haveOpenAPI, haveOpenAPIVersion bool
	for _, revision := range revisions {
		loader := openapi3.NewLoader()
		// JSON will deserialize here correctly
		src, err := loader.LoadFromData(revision.Blob)
		if err != nil {
			return nil, fmt.Errorf("could not load revision %s-%s-%s: %w", revision.Service, revision.Version, revision.Digest, err)
		}

		// Each service will declare their own /openapi paths. Collate only the
		// first instances of these paths and remove them from subsequent specs
		// to prevent failing the collate on conflicting paths.
		if haveOpenAPI {
			delete(src.Paths, "/openapi")
		} else if _, ok := src.Paths["/openapi"]; ok {
			haveOpenAPI = true
		}
		if haveOpenAPIVersion {
			delete(src.Paths, "/openapi/{version}")
		} else if _, ok := src.Paths["/openapi/{version}"]; ok {
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
			return nil, fmt.Errorf("could not collate revision %s-%s-%s: %w", revision.Service, revision.Version, revision.Digest, err)
		}
	}
	return collator.Result(), nil
}

type InlineOverlay struct {
	Overlays []*Overlay `json:"overlays"`
}

type Overlay struct {
	Include string `json:"include"`
	Inline  string `json:"inline"`
}

func loadConfig() (*InlineOverlay, error) {
	configPath := ".vervet.yaml"
	var o InlineOverlay

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", configPath, err)
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read project configuration: %w", err)
	}
	err = yaml.Unmarshal(buf, &o)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project configuration: %w", err)
	}

	return &o, nil
}

func applyOverlay(spec *openapi3.T) (*openapi3.T, error) {
	o, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, doc := range o.Overlays {
		docString := os.ExpandEnv(doc.Inline)
		l := openapi3.NewLoader()
		newDoc, err := l.LoadFromData([]byte(docString))
		if err != nil {
			return nil, fmt.Errorf("failed to load template: %w", err)
		}
		vervet.Merge(spec, newDoc, true)
	}

	return spec, err
}
