package service

import (
	"context"
	"fmt"
	"net/url"
	"sync"
)

// Service represents a target for vervet-underground to scrape.
type Service struct {
	Base string `json:"base"`
	URL  string `json:"url"`
	// TODO: track healthcheck?
}

// Registry is a registry of services scraped by vervet-underground.
type Registry struct {
	Services []Service

	loaders []Loader
	mu      sync.RWMutex
}

// NewRegistry returns a new service Registry.
func NewRegistry(loaders ...Loader) *Registry {
	return &Registry{
		Services: make([]Service, 0),
		loaders:  loaders,
	}
}

// Load loads services from loaders and replaces existing service entries.
func (r *Registry) Load() error {
	var loaded []string
	for _, loader := range r.loaders {
		services, err := loader(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to load services: %w", err)
		}
		loaded = append(loaded, services...)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.Services = nil
	if err := r.add(loaded...); err != nil {
		return fmt.Errorf("failed to add services to registry: %w", err)
	}
	return nil
}

// add a new service base to the Registry. Returns an error if base + /openapi is not a valid URL.
func (r *Registry) add(bases ...string) error {
	for _, base := range bases {
		u, err := url.Parse(base + "/openapi")
		if err != nil {
			return fmt.Errorf("invalid service %q: %w", base, err)
		}

		// Handle for local/smaller deployments and tests
		if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
			base = u.Host
		}
		r.Services = append(r.Services, Service{base, u.String()})
	}
	return nil
}
